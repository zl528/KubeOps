package api

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterManager struct {
	mu            sync.RWMutex
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
	restConfig    *rest.Config
	connected     bool
	clusterName   string
}

type ClusterInfo struct {
	Name    string `json:"name"`
	Server  string `json:"server"`
	Status  string `json:"status"`
	Active  bool   `json:"active"`
}

type MultiClusterManager struct {
	mu              sync.RWMutex
	clusters        map[string]*ClusterManager
	activeCluster   string
	db              *sql.DB
}

func (cm *ClusterManager) GetK8sClient() kubernetes.Interface {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.k8sClient
}

func (cm *ClusterManager) GetDynamicClient() dynamic.Interface {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.dynamicClient
}

func (cm *ClusterManager) GetRestConfig() *rest.Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.restConfig
}

func (cm *ClusterManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connected
}

func (cm *ClusterManager) GetClusterName() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clusterName
}

func (cm *ClusterManager) Connect(kubeconfig string, token string, server string, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var restConfig *rest.Config
	var err error

	if kubeconfig != "" {
		configObj, err := clientcmd.Load([]byte(kubeconfig))
		if err != nil {
			return fmt.Errorf("parse kubeconfig: %w", err)
		}

		for name, cluster := range configObj.Clusters {
			if cluster.InsecureSkipTLSVerify {
				cluster.CertificateAuthorityData = nil
				cluster.CertificateAuthority = ""
				configObj.Clusters[name] = cluster
			}
		}

		restConfig, err = clientcmd.NewDefaultClientConfig(*configObj, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return fmt.Errorf("build config from kubeconfig: %w", err)
		}
	} else if token != "" && server != "" {
		restConfig = &rest.Config{
			Host:        server,
			BearerToken: token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	} else if server != "" {
		restConfig = &rest.Config{
			Host: server,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	} else {
		return fmt.Errorf("kubeconfig, token+server, or server is required")
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create k8s client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create dynamic client: %w", err)
	}

	cm.k8sClient = k8sClient
	cm.dynamicClient = dynamicClient
	cm.restConfig = restConfig
	cm.connected = true
	cm.clusterName = name

	return nil
}

func (cm *ClusterManager) Disconnect() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	scheme := runtime.NewScheme()
	cm.k8sClient = fake.NewSimpleClientset()
	cm.dynamicClient = dynamicfake.NewSimpleDynamicClient(scheme)
	cm.restConfig = &rest.Config{Host: "http://localhost:8080"}
	cm.connected = false
	cm.clusterName = ""
}

var globalMultiClusterManager *MultiClusterManager

func GetMultiClusterManager() *MultiClusterManager {
	return globalMultiClusterManager
}

func InitMultiClusterManager(db *sql.DB) *MultiClusterManager {
	globalMultiClusterManager = &MultiClusterManager{
		clusters: make(map[string]*ClusterManager),
		db:       db,
	}
	globalMultiClusterManager.loadFromDB()
	return globalMultiClusterManager
}

func (m *MultiClusterManager) loadFromDB() {
	if m.db == nil {
		return
	}

	rows, err := m.db.Query("SELECT name, server, token, kubeconfig, status FROM cluster_connections")
	if err != nil {
		log.Printf("Failed to load clusters from database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, server, token, status string
		var kubeconfig sql.NullString
		if err := rows.Scan(&name, &server, &token, &kubeconfig, &status); err != nil {
			log.Printf("Failed to scan cluster: %v", err)
			continue
		}

		kubeconfigStr := ""
		if kubeconfig.Valid {
			kubeconfigStr = kubeconfig.String
		}

		if status == "connected" {
			cm := &ClusterManager{}
			if err := cm.Connect(kubeconfigStr, token, server, name); err != nil {
				log.Printf("Failed to reconnect cluster %s: %v", name, err)
				continue
			}
			m.clusters[name] = cm
			if m.activeCluster == "" {
				m.activeCluster = name
			}
			log.Printf("Reconnected to cluster: %s", name)
		}
	}
}

func (m *MultiClusterManager) GetActiveCluster() *ClusterManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.activeCluster == "" {
		return nil
	}
	return m.clusters[m.activeCluster]
}

func (m *MultiClusterManager) GetActiveClusterName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeCluster
}

func (m *MultiClusterManager) GetK8sClient() kubernetes.Interface {
	cm := m.GetActiveCluster()
	if cm == nil {
		return fake.NewSimpleClientset()
	}
	return cm.GetK8sClient()
}

func (m *MultiClusterManager) GetDynamicClient() dynamic.Interface {
	cm := m.GetActiveCluster()
	if cm == nil {
		scheme := runtime.NewScheme()
		return dynamicfake.NewSimpleDynamicClient(scheme)
	}
	return cm.GetDynamicClient()
}

func (m *MultiClusterManager) GetRestConfig() *rest.Config {
	cm := m.GetActiveCluster()
	if cm == nil {
		return &rest.Config{Host: "http://localhost:8080"}
	}
	return cm.GetRestConfig()
}

func (m *MultiClusterManager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeCluster != ""
}

func (m *MultiClusterManager) ListClusters() []ClusterInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var clusters []ClusterInfo
	for name, cm := range m.clusters {
		clusters = append(clusters, ClusterInfo{
			Name:   name,
			Server: cm.restConfig.Host,
			Status: "connected",
			Active: name == m.activeCluster,
		})
	}
	return clusters
}

func (m *MultiClusterManager) AddCluster(name, server, token, kubeconfig string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clusters[name]; exists {
		return fmt.Errorf("cluster %s already exists", name)
	}

	cm := &ClusterManager{}
	err := cm.Connect(kubeconfig, token, server, name)
	if err != nil {
		return fmt.Errorf("connect to cluster: %w", err)
	}

	m.clusters[name] = cm
	m.activeCluster = name

	if m.db != nil {
		m.db.Exec(
			`INSERT OR REPLACE INTO cluster_connections (name, server, token, kubeconfig, status) VALUES (?, ?, ?, ?, ?)`,
			name, server, token, kubeconfig, "connected",
		)
	}

	return nil
}

func (m *MultiClusterManager) RemoveCluster(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clusters[name]; !exists {
		return fmt.Errorf("cluster %s not found", name)
	}

	delete(m.clusters, name)

	if m.activeCluster == name {
		m.activeCluster = ""
		for k := range m.clusters {
			m.activeCluster = k
			break
		}
	}

	if m.db != nil {
		m.db.Exec("DELETE FROM cluster_connections WHERE name = ?", name)
	}

	return nil
}

func (m *MultiClusterManager) SwitchCluster(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clusters[name]; !exists {
		return fmt.Errorf("cluster %s not found", name)
	}

	m.activeCluster = name
	return nil
}

func (m *MultiClusterManager) Connect(name, server, token, kubeconfig string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cm, exists := m.clusters[name]
	if !exists {
		cm = &ClusterManager{}
		m.clusters[name] = cm
	}

	err := cm.Connect(kubeconfig, token, server, name)
	if err != nil {
		return err
	}

	m.activeCluster = name

	if m.db != nil {
		m.db.Exec(
			`INSERT OR REPLACE INTO cluster_connections (name, server, token, status) VALUES (?, ?, ?, ?)`,
			name, server, token, "connected",
		)
	}

	return nil
}

func (m *MultiClusterManager) Disconnect(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cm, exists := m.clusters[name]; exists {
		cm.Disconnect()
		delete(m.clusters, name)
	}

	if m.activeCluster == name {
		m.activeCluster = ""
		for k := range m.clusters {
			m.activeCluster = k
			break
		}
	}

	if m.db != nil {
		m.db.Exec("DELETE FROM cluster_connections WHERE name = ?", name)
	}
}

func (m *MultiClusterManager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, cm := range m.clusters {
		cm.Disconnect()
		delete(m.clusters, name)
	}
	m.activeCluster = ""

	if m.db != nil {
		m.db.Exec("DELETE FROM cluster_connections")
	}
}

func (m *MultiClusterManager) GetClientGetter() *ClientGetter {
	return &ClientGetter{manager: m}
}

type ClientGetter struct {
	manager *MultiClusterManager
}

func (cg *ClientGetter) GetK8sClient() kubernetes.Interface {
	return cg.manager.GetK8sClient()
}

func (cg *ClientGetter) GetDynamicClient() dynamic.Interface {
	return cg.manager.GetDynamicClient()
}

func (cg *ClientGetter) GetRestConfig() *rest.Config {
	return cg.manager.GetRestConfig()
}

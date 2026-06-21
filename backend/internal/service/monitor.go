package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type MonitorService struct {
	client        kubernetes.Interface
	prometheusURL string
}

var cachedPrometheusURL string

func NewMonitorService(client kubernetes.Interface, prometheusURL string) *MonitorService {
	return &MonitorService{
		client:        client,
		prometheusURL: prometheusURL,
	}
}

func (s *MonitorService) getPrometheusURL(ctx context.Context) string {
	if s.prometheusURL != "" {
		return s.prometheusURL
	}
	if cachedPrometheusURL != "" {
		return cachedPrometheusURL
	}
	// Auto-detect
	url := s.DetectPrometheusURL(ctx)
	log.Printf("[MONITOR] getPrometheusURL: detected URL=%s", url)
	if url != "" {
		cachedPrometheusURL = url
	}
	return url
}

func (s *MonitorService) DetectPrometheusURL(ctx context.Context) string {
	if s.prometheusURL != "" {
		return s.prometheusURL
	}
	// Try to find Prometheus service in cluster
	svc, err := s.client.CoreV1().Services("prometheus").Get(ctx, "prometheus-server", metav1.GetOptions{})
	if err == nil {
		for _, port := range svc.Spec.Ports {
			if port.NodePort > 0 {
				// Use NodePort
				nodes, _ := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
				if len(nodes.Items) > 0 {
					for _, addr := range nodes.Items[0].Status.Addresses {
						if addr.Type == corev1.NodeInternalIP {
							return fmt.Sprintf("http://%s:%d", addr.Address, port.NodePort)
						}
					}
				}
			}
		}
	}
	// Try common Prometheus URLs
	urls := []string{
		"http://prometheus-server.prometheus:9090",
		"http://prometheus:9090",
		"http://localhost:9090",
	}
	for _, url := range urls {
		resp, err := http.Get(url + "/api/v1/query?query=up")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return url
		}
	}
	return ""
}

type MonitoringCapabilities struct {
	BasicMetrics    bool `json:"basicMetrics"`
	ResourceMetrics bool `json:"resourceMetrics"`
	HistoricalData  bool `json:"historicalData"`
	AlertRules      bool `json:"alertRules"`
	PrometheusReady bool `json:"prometheusReady"`
}

type PrometheusStatus struct {
	Available    bool   `json:"available"`
	URL          string `json:"url,omitempty"`
	TargetsUp    int    `json:"targetsUp"`
	TargetsTotal int    `json:"targetsTotal"`
	AlertsCount  int    `json:"alertsCount"`
	Version      string `json:"version,omitempty"`
}

func (s *MonitorService) DetectCapabilities(ctx context.Context) (*MonitoringCapabilities, error) {
	caps := &MonitoringCapabilities{
		BasicMetrics: true,
	}

	// Check metrics-server
	_, err := s.client.Discovery().ServerResourcesForGroupVersion("metrics.k8s.io/v1beta1")
	if err == nil {
		caps.ResourceMetrics = true
	}

	// Auto-detect Prometheus URL
	promURL := s.DetectPrometheusURL(ctx)
	if promURL != "" {
		s.prometheusURL = promURL
		cachedPrometheusURL = promURL
		caps.HistoricalData = true
		caps.AlertRules = true
		caps.PrometheusReady = true
	}

	return caps, nil
}

func (s *MonitorService) GetPrometheusStatus(ctx context.Context) (*PrometheusStatus, error) {
	// Auto-detect if not set
	if s.prometheusURL == "" {
		s.prometheusURL = s.DetectPrometheusURL(ctx)
	}

	status := &PrometheusStatus{
		URL: s.prometheusURL,
	}

	if s.prometheusURL == "" {
		return status, nil
	}

	// Check if Prometheus is available
	resp, err := http.Get(s.prometheusURL + "/api/v1/status/config")
	if err != nil || resp.StatusCode != 200 {
		return status, nil
	}
	defer resp.Body.Close()
	status.Available = true

	// Get version
	body, _ := io.ReadAll(resp.Body)
	var configResp struct {
		Status string `json:"status"`
	}
	json.Unmarshal(body, &configResp)

	// Get targets
	targetsResp, err := http.Get(s.prometheusURL + "/api/v1/targets")
	if err == nil {
		defer targetsResp.Body.Close()
		var targets struct {
			Data struct {
				ActiveTargets []struct {
					Health string `json:"health"`
				} `json:"activeTargets"`
			} `json:"data"`
		}
		body, _ := io.ReadAll(targetsResp.Body)
		json.Unmarshal(body, &targets)
		status.TargetsTotal = len(targets.Data.ActiveTargets)
		for _, t := range targets.Data.ActiveTargets {
			if t.Health == "up" {
				status.TargetsUp++
			}
		}
	}

	// Get alerts count
	alertsResp, err := http.Get(s.prometheusURL + "/api/v1/alerts")
	if err == nil {
		defer alertsResp.Body.Close()
		var alerts struct {
			Data struct {
				Alerts []struct {
					State string `json:"state"`
				} `json:"alerts"`
			} `json:"data"`
		}
		body, _ := io.ReadAll(alertsResp.Body)
		json.Unmarshal(body, &alerts)
		status.AlertsCount = len(alerts.Data.Alerts)
	}

	return status, nil
}

func (s *MonitorService) GetPrometheusTargets(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := http.Get(s.prometheusURL + "/api/v1/targets")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ActiveTargets []struct {
				Labels      map[string]string `json:"labels"`
				ScrapeURL   string            `json:"scrapeUrl"`
				Health      string            `json:"health"`
				LastError   string            `json:"lastError"`
				LastScrape  string            `json:"lastScrape"`
			} `json:"activeTargets"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var targets []map[string]interface{}
	for _, t := range result.Data.ActiveTargets {
		targets = append(targets, map[string]interface{}{
			"url":     t.ScrapeURL,
			"health":  t.Health,
			"labels":  t.Labels,
			"error":   t.LastError,
			"lastScrape": t.LastScrape,
		})
	}
	return targets, nil
}

type NodeMetrics struct {
	Name           string  `json:"name"`
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
	CPUCores      float64 `json:"cpuCores"`
	MemoryBytes   int64   `json:"memoryBytes"`
	DiskUsage     float64 `json:"diskUsage,omitempty"`
	NetworkRx     int64   `json:"networkRx,omitempty"`
	NetworkTx     int64   `json:"networkTx,omitempty"`
}

type PodMetrics struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Node        string            `json:"node"`
	Containers  []ContainerMetrics `json:"containers"`
	CPUUsage    float64           `json:"cpuUsage"`
	MemoryUsage int64             `json:"memoryUsage"`
}

type ContainerMetrics struct {
	Name        string  `json:"name"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage int64   `json:"memoryUsage"`
}

type ClusterMetrics struct {
	TotalCPU     float64 `json:"totalCPU"`
	UsedCPU      float64 `json:"usedCPU"`
	TotalMemory  int64   `json:"totalMemory"`
	UsedMemory   int64   `json:"usedMemory"`
	TotalPods    int     `json:"totalPods"`
	RunningPods  int     `json:"runningPods"`
	TotalNodes   int     `json:"totalNodes"`
	ReadyNodes   int     `json:"readyNodes"`
}

func (s *MonitorService) GetClusterMetrics(ctx context.Context) (*ClusterMetrics, error) {
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods, err := s.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	metrics := &ClusterMetrics{
		TotalNodes:  len(nodes.Items),
		TotalPods:   len(pods.Items),
	}

	for _, node := range nodes.Items {
		if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
			metrics.TotalCPU += float64(cpu.MilliValue()) / 1000.0
		}
		if mem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
			metrics.TotalMemory += mem.Value()
		}
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				metrics.ReadyNodes++
			}
		}
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			metrics.RunningPods++
			for _, container := range pod.Spec.Containers {
				if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					metrics.UsedCPU += float64(cpu.MilliValue()) / 1000.0
				}
				if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					metrics.UsedMemory += mem.Value()
				}
			}
		}
	}

	return metrics, nil
}

func (s *MonitorService) GetNodeMetrics(ctx context.Context, nodeName string) (*NodeMetrics, error) {
	node, err := s.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metrics := &NodeMetrics{
		Name: node.Name,
	}

	if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
		metrics.CPUCores = float64(cpu.MilliValue()) / 1000.0
	}
	if mem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
		metrics.MemoryBytes = mem.Value()
	}

	return metrics, nil
}

func (s *MonitorService) GetPodMetrics(ctx context.Context, namespace, podName string) (*PodMetrics, error) {
	pod, err := s.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metrics := &PodMetrics{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Node:      pod.Spec.NodeName,
	}

	for _, container := range pod.Spec.Containers {
		cm := ContainerMetrics{
			Name: container.Name,
		}
		if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
			cm.CPUUsage = float64(cpu.MilliValue()) / 1000.0
			metrics.CPUUsage += cm.CPUUsage
		}
		if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
			cm.MemoryUsage = mem.Value()
			metrics.MemoryUsage += cm.MemoryUsage
		}
		metrics.Containers = append(metrics.Containers, cm)
	}

	return metrics, nil
}

type PrometheusQueryResult struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}

type PrometheusData struct {
	Type   string            `json:"type"`
	Result []json.RawMessage `json:"result"`
}

type PrometheusMetric struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

func (s *MonitorService) QueryPrometheus(query string) (interface{}, error) {
	prometheusURL := s.prometheusURL
	if prometheusURL == "" {
		prometheusURL = cachedPrometheusURL
	}
	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL not configured")
	}

	reqURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, url.QueryEscape(query))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed")
	}

	return result.Data, nil
}

func (s *MonitorService) GetCPUUsage(ctx context.Context) (float64, error) {
	query := `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
	promURL := s.getPrometheusURL(ctx)
	if promURL == "" {
		return 0, fmt.Errorf("prometheus URL not configured")
	}
	s.prometheusURL = promURL
	result, err := s.QueryPrometheus(query)
	if err != nil {
		log.Printf("[MONITOR] GetCPUUsage error: %v", err)
		return 0, err
	}
	log.Printf("[MONITOR] GetCPUUsage result: %v", result)

	if data, ok := result.(PrometheusData); ok {
		if len(data.Result) > 0 {
			var metric PrometheusMetric
			if err := json.Unmarshal(data.Result[0], &metric); err == nil {
				if len(metric.Value) >= 2 {
					if val, ok := metric.Value[1].(string); ok {
						var usage float64
						fmt.Sscanf(val, "%f", &usage)
						return usage, nil
					}
				}
			}
		}
	}

	return 0, nil
}

func (s *MonitorService) GetMemoryUsage(ctx context.Context) (float64, error) {
	query := `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
	promURL := s.getPrometheusURL(ctx)
	if promURL == "" {
		return 0, fmt.Errorf("prometheus URL not configured")
	}
	s.prometheusURL = promURL
	result, err := s.QueryPrometheus(query)
	if err != nil {
		return 0, err
	}

	if data, ok := result.(PrometheusData); ok {
		if len(data.Result) > 0 {
			var metric PrometheusMetric
			if err := json.Unmarshal(data.Result[0], &metric); err == nil {
				if len(metric.Value) >= 2 {
					if val, ok := metric.Value[1].(string); ok {
						var usage float64
						fmt.Sscanf(val, "%f", &usage)
						return usage, nil
					}
				}
			}
		}
	}

	return 0, nil
}

func (s *MonitorService) GetDiskUsage(ctx context.Context) (float64, error) {
	query := `(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100`
	promURL := s.getPrometheusURL(ctx)
	if promURL == "" {
		return 0, fmt.Errorf("prometheus URL not configured")
	}
	s.prometheusURL = promURL
	result, err := s.QueryPrometheus(query)
	if err != nil {
		return 0, err
	}

	if data, ok := result.(PrometheusData); ok {
		if len(data.Result) > 0 {
			var metric PrometheusMetric
			if err := json.Unmarshal(data.Result[0], &metric); err == nil {
				if len(metric.Value) >= 2 {
					if val, ok := metric.Value[1].(string); ok {
						var usage float64
						fmt.Sscanf(val, "%f", &usage)
						return usage, nil
					}
				}
			}
		}
	}

	return 0, nil
}

func (s *MonitorService) GetNetworkUsage(ctx context.Context) (rxBytes, txBytes int64, err error) {
	queryRx := `rate(node_network_receive_bytes_total{device!="lo"}[5m])`
	queryTx := `rate(node_network_transmit_bytes_total{device!="lo"}[5m])`

	promURL := s.getPrometheusURL(ctx)
	if promURL == "" {
		return 0, 0, fmt.Errorf("prometheus URL not configured")
	}
	s.prometheusURL = promURL

	rxResult, err := s.QueryPrometheus(queryRx)
	if err != nil {
		return 0, 0, err
	}

	txResult, err := s.QueryPrometheus(queryTx)
	if err != nil {
		return 0, 0, err
	}

	if data, ok := rxResult.(PrometheusData); ok && len(data.Result) > 0 {
		var metric PrometheusMetric
		if err := json.Unmarshal(data.Result[0], &metric); err == nil {
			if len(metric.Value) >= 2 {
				if val, ok := metric.Value[1].(string); ok {
					fmt.Sscanf(val, "%f", &rxBytes)
				}
			}
		}
	}

	if data, ok := txResult.(PrometheusData); ok && len(data.Result) > 0 {
		var metric PrometheusMetric
		if err := json.Unmarshal(data.Result[0], &metric); err == nil {
			if len(metric.Value) >= 2 {
				if val, ok := metric.Value[1].(string); ok {
					fmt.Sscanf(val, "%f", &txBytes)
				}
			}
		}
	}

	return rxBytes, txBytes, nil
}

type GrafanaDashboard struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	UID   string `json:"uid"`
}

func (s *MonitorService) GetGrafanaDashboards(grafanaURL string) ([]GrafanaDashboard, error) {
	if grafanaURL == "" {
		grafanaURL = "http://grafana:3000"
	}

	url := fmt.Sprintf("%s/api/search?type=dash-db", grafanaURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("grafana query failed: %w", err)
	}
	defer resp.Body.Close()

	var dashboards []GrafanaDashboard
	if err := json.NewDecoder(resp.Body).Decode(&dashboards); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return dashboards, nil
}

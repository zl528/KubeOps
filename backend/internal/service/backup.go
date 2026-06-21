package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type BackupService struct {
	client        kubernetes.Interface
	dynamicClient  dynamic.Interface
	backupDir     string
	mu            sync.RWMutex
	backups       []BackupRecord
}

type BackupRecord struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Description string            `json:"description,omitempty"`
	Resources   []string          `json:"resources"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"createdAt"`
	CompletedAt *time.Time        `json:"completedAt,omitempty"`
	Size        int64             `json:"size,omitempty"`
	Error       string            `json:"error,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type BackupResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace,omitempty"`
	Data       map[string]interface{} `json:"data"`
}

type BackupContent struct {
	Record     BackupRecord      `json:"record"`
	Resources  []BackupResource  `json:"resources"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type RestoreRequest struct {
	BackupID    string   `json:"backupId"`
	Namespace   string   `json:"namespace,omitempty"`
	Resources   []string `json:"resources,omitempty"`
	DryRun      bool     `json:"dryRun,omitempty"`
	Overwrite   bool     `json:"overwrite,omitempty"`
}

type RestoreResult struct {
	Restored int      `json:"restored"`
	Skipped  int      `json:"skipped"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

func NewBackupService(client kubernetes.Interface, dynamicClient dynamic.Interface, backupDir string) *BackupService {
	if backupDir == "" {
		backupDir = "/tmp/kubeops-backups"
	}

	svc := &BackupService{
		client:        client,
		dynamicClient:  dynamicClient,
		backupDir:     backupDir,
		backups:       make([]BackupRecord, 0),
	}

	os.MkdirAll(backupDir, 0755)
	svc.loadBackups()

	return svc
}

func (s *BackupService) loadBackups() {
	filePath := filepath.Join(s.backupDir, "backups.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var backups []BackupRecord
	if err := json.Unmarshal(data, &backups); err != nil {
		return
	}

	s.mu.Lock()
	s.backups = backups
	s.mu.Unlock()
}

func (s *BackupService) saveBackups() error {
	filePath := filepath.Join(s.backupDir, "backups.json")
	data, err := json.MarshalIndent(s.backups, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

var resourceTypes = map[string]schema.GroupVersionResource{
	"deployments":  {Group: "apps", Version: "v1", Resource: "deployments"},
	"statefulsets": {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"daemonsets":   {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"services":     {Group: "", Version: "v1", Resource: "services"},
	"configmaps":   {Group: "", Version: "v1", Resource: "configmaps"},
	"secrets":      {Group: "", Version: "v1", Resource: "secrets"},
	"ingresses":    {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"roles":        {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	"rolebindings": {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	"serviceaccounts": {Group: "", Version: "v1", Resource: "serviceaccounts"},
	"persistentvolumeclaims": {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	"cronjobs":     {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"jobs":         {Group: "batch", Version: "v1", Resource: "jobs"},
	"horizontalpodautoscalers": {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"},
	"poddisruptionbudgets": {Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"},
}

func (s *BackupService) CreateBackup(ctx context.Context, name, namespace, description string, resources []string) (*BackupRecord, error) {
	record := BackupRecord{
		ID:          fmt.Sprintf("backup-%d", time.Now().UnixNano()),
		Name:        name,
		Namespace:   namespace,
		Description: description,
		Resources:   resources,
		Status:      "running",
		CreatedAt:   time.Now(),
	}

	if len(resources) == 0 {
		resources = []string{"deployments", "services", "configmaps", "secrets"}
		record.Resources = resources
	}

	s.mu.Lock()
	s.backups = append(s.backups, record)
	s.mu.Unlock()

	go s.performBackup(record, resources)

	return &record, nil
}

func (s *BackupService) performBackup(record BackupRecord, resources []string) {
	backupContent := BackupContent{
		Record:    record,
		Resources: make([]BackupResource, 0),
		Metadata: map[string]string{
			"version": "1.0",
			"time":    time.Now().Format(time.RFC3339),
		},
	}

	for _, resType := range resources {
		gvr, ok := resourceTypes[resType]
		if !ok {
			continue
		}

		var list *unstructured.UnstructuredList
		var err error

		if record.Namespace != "" {
			list, err = s.dynamicClient.Resource(gvr).Namespace(record.Namespace).List(context.Background(), metav1.ListOptions{})
		} else {
			list, err = s.dynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
		}

		if err != nil {
			s.updateBackupStatus(record.ID, "failed", fmt.Sprintf("failed to list %s: %v", resType, err))
			return
		}

		for _, item := range list.Items {
			backupRes := BackupResource{
				APIVersion: item.GetAPIVersion(),
				Kind:       item.GetKind(),
				Name:       item.GetName(),
				Namespace:  item.GetNamespace(),
				Data:       item.UnstructuredContent(),
			}
			backupContent.Resources = append(backupContent.Resources, backupRes)
		}
	}

	backupFile := filepath.Join(s.backupDir, record.ID+".json")
	data, err := json.MarshalIndent(backupContent, "", "  ")
	if err != nil {
		s.updateBackupStatus(record.ID, "failed", fmt.Sprintf("marshal error: %v", err))
		return
	}

	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		s.updateBackupStatus(record.ID, "failed", fmt.Sprintf("write error: %v", err))
		return
	}

	now := time.Now()
	s.updateBackupStatus(record.ID, "completed", "")
	s.updateBackupSize(record.ID, int64(len(data)))
	s.updateBackupCompletedAt(record.ID, &now)
}

func (s *BackupService) updateBackupStatus(id, status, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.backups {
		if b.ID == id {
			s.backups[i].Status = status
			s.backups[i].Error = errMsg
			break
		}
	}
	s.saveBackups()
}

func (s *BackupService) updateBackupSize(id string, size int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.backups {
		if b.ID == id {
			s.backups[i].Size = size
			break
		}
	}
	s.saveBackups()
}

func (s *BackupService) updateBackupCompletedAt(id string, t *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.backups {
		if b.ID == id {
			s.backups[i].CompletedAt = t
			break
		}
	}
	s.saveBackups()
}

func (s *BackupService) ListBackups(ctx context.Context) []BackupRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]BackupRecord, len(s.backups))
	copy(result, s.backups)

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

func (s *BackupService) GetBackup(ctx context.Context, id string) (*BackupRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.backups {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("backup not found: %s", id)
}

func (s *BackupService) DeleteBackup(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.backups {
		if b.ID == id {
			backupFile := filepath.Join(s.backupDir, id+".json")
			os.Remove(backupFile)

			s.backups = append(s.backups[:i], s.backups[i+1:]...)
			return s.saveBackups()
		}
	}
	return fmt.Errorf("backup not found: %s", id)
}

func (s *BackupService) GetBackupContent(ctx context.Context, id string) (*BackupContent, error) {
	backupFile := filepath.Join(s.backupDir, id+".json")
	data, err := os.ReadFile(backupFile)
	if err != nil {
		return nil, fmt.Errorf("backup file not found: %s", id)
	}

	var content BackupContent
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("parse backup content: %w", err)
	}

	return &content, nil
}

func (s *BackupService) Restore(ctx context.Context, req RestoreRequest) (*RestoreResult, error) {
	content, err := s.GetBackupContent(ctx, req.BackupID)
	if err != nil {
		return nil, err
	}

	result := &RestoreResult{}

	for _, res := range content.Resources {
		if len(req.Resources) > 0 {
			found := false
			for _, r := range req.Resources {
				if res.Kind == r || resourceKindToKey(res.Kind) == r {
					found = true
					break
				}
			}
			if !found {
				result.Skipped++
				continue
			}
		}

		gvr, ok := resourceTypes[resourceKindToKey(res.Kind)]
		if !ok {
			result.Skipped++
			continue
		}

		namespace := res.Namespace
		if req.Namespace != "" {
			namespace = req.Namespace
		}

		obj := &unstructured.Unstructured{
			Object: res.Data,
		}

		if req.DryRun {
			result.Restored++
			continue
		}

		var createErr error
		if namespace != "" {
			_, createErr = s.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
		} else {
			_, createErr = s.dynamicClient.Resource(gvr).Create(ctx, obj, metav1.CreateOptions{})
		}

		if createErr != nil {
			if req.Overwrite {
				if namespace != "" {
					_, createErr = s.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
				} else {
					_, createErr = s.dynamicClient.Resource(gvr).Update(ctx, obj, metav1.UpdateOptions{})
				}
			}

			if createErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("%s/%s: %v", res.Kind, res.Name, createErr))
				continue
			}
		}

		result.Restored++
	}

	return result, nil
}

func (s *BackupService) ExportBackup(ctx context.Context, id string) ([]byte, error) {
	content, err := s.GetBackupContent(ctx, id)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(content, "", "  ")
}

func (s *BackupService) ImportBackup(ctx context.Context, data []byte, name string) (*BackupRecord, error) {
	var content BackupContent
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("parse import data: %w", err)
	}

	record := BackupRecord{
		ID:          fmt.Sprintf("backup-%d", time.Now().UnixNano()),
		Name:        name,
		Description: "Imported backup",
		Resources:   []string{"imported"},
		Status:      "completed",
		CreatedAt:   time.Now(),
		Size:        int64(len(data)),
	}

	content.Record = record

	backupFile := filepath.Join(s.backupDir, record.ID+".json")
	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return nil, fmt.Errorf("write backup file: %w", err)
	}

	s.mu.Lock()
	s.backups = append(s.backups, record)
	s.mu.Unlock()
	s.saveBackups()

	return &record, nil
}

func resourceKindToKey(kind string) string {
	kindMap := map[string]string{
		"Deployment":                   "deployments",
		"StatefulSet":                  "statefulsets",
		"DaemonSet":                    "daemonsets",
		"Service":                      "services",
		"ConfigMap":                    "configmaps",
		"Secret":                       "secrets",
		"Ingress":                      "ingresses",
		"Role":                         "roles",
		"RoleBinding":                  "rolebindings",
		"ServiceAccount":               "serviceaccounts",
		"PersistentVolumeClaim":        "persistentvolumeclaims",
		"CronJob":                      "cronjobs",
		"Job":                          "jobs",
		"HorizontalPodAutoscaler":      "horizontalpodautoscalers",
		"PodDisruptionBudget":          "poddisruptionbudgets",
	}

	if key, ok := kindMap[kind]; ok {
		return key
	}
	return ""
}

// Unused imports
var _ = runtime.Object(nil)
var _ = schema.GroupVersionResource{}

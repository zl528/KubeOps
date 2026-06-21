package service

import (
	"context"
	"os"
	"testing"
)

func TestBackupService_ListBackups(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backup-test")
	defer os.RemoveAll(dir)

	client := createTestClient()
	svc := NewBackupService(client, nil, dir)

	backups := svc.ListBackups(context.Background())
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

func TestBackupService_DeleteBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backup-test")
	defer os.RemoveAll(dir)

	client := createTestClient()
	svc := NewBackupService(client, nil, dir)

	err := svc.DeleteBackup(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backup")
	}
}

func TestBackupService_GetBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backup-test")
	defer os.RemoveAll(dir)

	client := createTestClient()
	svc := NewBackupService(client, nil, dir)

	_, err := svc.GetBackup(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backup")
	}
}

func TestBackupService_GetBackupContent(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backup-test")
	defer os.RemoveAll(dir)

	client := createTestClient()
	svc := NewBackupService(client, nil, dir)

	_, err := svc.GetBackupContent(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backup")
	}
}

func TestBackupService_ImportBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backup-test")
	defer os.RemoveAll(dir)

	client := createTestClient()
	svc := NewBackupService(client, nil, dir)

	data := []byte(`{"record":{"id":"test","name":"test"},"resources":[]}`)
	record, err := svc.ImportBackup(context.Background(), data, "test-import")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.Name != "test-import" {
		t.Errorf("expected name 'test-import', got %s", record.Name)
	}

	if record.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", record.Status)
	}

	backups := svc.ListBackups(context.Background())
	if len(backups) != 1 {
		t.Errorf("expected 1 backup after import, got %d", len(backups))
	}
}

func TestBackupService_ExportBackup(t *testing.T) {
	dir, _ := os.MkdirTemp("", "backup-test")
	defer os.RemoveAll(dir)

	client := createTestClient()
	svc := NewBackupService(client, nil, dir)

	data := []byte(`{"record":{"id":"test","name":"test"},"resources":[]}`)
	svc.ImportBackup(context.Background(), data, "test-export")

	_, err := svc.ExportBackup(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backup")
	}
}

func TestResourceKindToKey(t *testing.T) {
	tests := []struct {
		kind     string
		expected string
	}{
		{"Deployment", "deployments"},
		{"Service", "services"},
		{"ConfigMap", "configmaps"},
		{"Secret", "secrets"},
		{"Ingress", "ingresses"},
		{"Role", "roles"},
		{"RoleBinding", "rolebindings"},
		{"ServiceAccount", "serviceaccounts"},
		{"PersistentVolumeClaim", "persistentvolumeclaims"},
		{"CronJob", "cronjobs"},
		{"Job", "jobs"},
		{"HorizontalPodAutoscaler", "horizontalpodautoscalers"},
		{"Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			result := resourceKindToKey(tt.kind)
			if result != tt.expected {
				t.Errorf("resourceKindToKey(%s) = %s, want %s", tt.kind, result, tt.expected)
			}
		})
	}
}

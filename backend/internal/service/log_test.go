package service

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func createTestContext() context.Context {
	return context.Background()
}

func createTestClient() *fake.Clientset {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:latest"},
				{Name: "sidecar", Image: "sidecar:latest"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	return fake.NewSimpleClientset(pod)
}

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogEntry
	}{
		{
			name:  "error log",
			input: "2024-01-15T10:30:00Z ERROR Something went wrong",
			expected: LogEntry{
				Level:   "error",
				Message: "2024-01-15T10:30:00Z ERROR Something went wrong",
			},
		},
		{
			name:  "warn log",
			input: "2024-01-15T10:30:00Z WARNING High memory usage",
			expected: LogEntry{
				Level:   "warn",
				Message: "2024-01-15T10:30:00Z WARNING High memory usage",
			},
		},
		{
			name:  "info log",
			input: "2024-01-15T10:30:00Z INFO Server started",
			expected: LogEntry{
				Level:   "info",
				Message: "2024-01-15T10:30:00Z INFO Server started",
			},
		},
		{
			name:  "debug log",
			input: "2024-01-15T10:30:00Z DEBUG Processing request",
			expected: LogEntry{
				Level:   "debug",
				Message: "2024-01-15T10:30:00Z DEBUG Processing request",
			},
		},
		{
			name:  "plain log",
			input: "Just a plain log message",
			expected: LogEntry{
				Message: "Just a plain log message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLine(tt.input)
			if result.Level != tt.expected.Level {
				t.Errorf("expected level %s, got %s", tt.expected.Level, result.Level)
			}
		})
	}
}

func TestLogService_ListPodContainers(t *testing.T) {
	client := createTestClient()
	svc := NewLogService(client)

	containers, err := svc.ListPodContainers(createTestContext(), "default", "test-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}

	if containers[0] != "nginx" {
		t.Errorf("expected first container 'nginx', got %s", containers[0])
	}
}

package service

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterService_GetOverview(t *testing.T) {
	nodes := &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
					},
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("8Gi"),
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
					},
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("8Gi"),
					},
				},
			},
		},
	}

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-2", Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodPending},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-3", Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodFailed},
			},
		},
	}

	client := fake.NewSimpleClientset(nodes, pods)
	svc := NewClusterService(client)

	overview, err := svc.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if overview.TotalNodes != 2 {
		t.Errorf("expected 2 total nodes, got %d", overview.TotalNodes)
	}
	if overview.ReadyNodes != 1 {
		t.Errorf("expected 1 ready node, got %d", overview.ReadyNodes)
	}
	if overview.RunningPods != 1 {
		t.Errorf("expected 1 running pod, got %d", overview.RunningPods)
	}
	if overview.PendingPods != 1 {
		t.Errorf("expected 1 pending pod, got %d", overview.PendingPods)
	}
	if overview.FailedPods != 1 {
		t.Errorf("expected 1 failed pod, got %d", overview.FailedPods)
	}
}

func TestClusterService_ListNodes(t *testing.T) {
	nodes := &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "master-1",
					Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
					},
					NodeInfo: corev1.NodeSystemInfo{
						KubeletVersion: "v1.28.0",
					},
					Allocatable: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("4"),
						corev1.ResourceMemory: resource.MustParse("8Gi"),
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(nodes)
	svc := NewClusterService(client)

	result, err := svc.ListNodes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}

	if result[0].Name != "master-1" {
		t.Errorf("expected node name master-1, got %s", result[0].Name)
	}
	if result[0].Status != "Ready" {
		t.Errorf("expected status Ready, got %s", result[0].Status)
	}
}

func TestClusterService_ListNamespaces(t *testing.T) {
	namespaces := &corev1.NamespaceList{
		Items: []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "default",
					CreationTimestamp: metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
				},
				Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "kube-system",
					CreationTimestamp: metav1.Time{Time: time.Now().Add(-48 * time.Hour)},
				},
				Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
			},
		},
	}

	client := fake.NewSimpleClientset(namespaces)
	svc := NewClusterService(client)

	result, err := svc.ListNamespaces(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 namespaces, got %d", len(result))
	}
}

func TestClusterService_ListPods(t *testing.T) {
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "nginx-1",
					Namespace:         "default",
					CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "nginx", Ready: true, RestartCount: 0},
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "node-1",
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx:latest"},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pods)
	svc := NewClusterService(client)

	result, err := svc.ListPods(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 pod, got %d", len(result))
	}

	if result[0].Name != "nginx-1" {
		t.Errorf("expected pod name nginx-1, got %s", result[0].Name)
	}
	if result[0].Status != "Running" {
		t.Errorf("expected status Running, got %s", result[0].Status)
	}
	if result[0].Node != "node-1" {
		t.Errorf("expected node node-1, got %s", result[0].Node)
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "seconds ago",
			input:    time.Now().Add(-30 * time.Second),
			expected: "30s",
		},
		{
			name:     "minutes ago",
			input:    time.Now().Add(-5 * time.Minute),
			expected: "5m",
		},
		{
			name:     "hours ago",
			input:    time.Now().Add(-3 * time.Hour),
			expected: "3h",
		},
		{
			name:     "days ago",
			input:    time.Now().Add(-7 * 24 * time.Hour),
			expected: "7d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

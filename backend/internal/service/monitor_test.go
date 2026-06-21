package service

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestMonitorService_GetClusterMetrics(t *testing.T) {
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
		},
	}

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "nginx",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(nodes, pods)
	svc := NewMonitorService(client, "")

	metrics, err := svc.GetClusterMetrics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.TotalNodes != 1 {
		t.Errorf("expected 1 total node, got %d", metrics.TotalNodes)
	}

	if metrics.ReadyNodes != 1 {
		t.Errorf("expected 1 ready node, got %d", metrics.ReadyNodes)
	}

	if metrics.TotalPods != 1 {
		t.Errorf("expected 1 total pod, got %d", metrics.TotalPods)
	}

	if metrics.RunningPods != 1 {
		t.Errorf("expected 1 running pod, got %d", metrics.RunningPods)
	}
}

func TestMonitorService_GetNodeMetrics(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}

	client := fake.NewSimpleClientset(node)
	svc := NewMonitorService(client, "")

	metrics, err := svc.GetNodeMetrics(context.Background(), "node-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.Name != "node-1" {
		t.Errorf("expected name 'node-1', got %s", metrics.Name)
	}

	if metrics.CPUCores != 4 {
		t.Errorf("expected 4 CPU cores, got %f", metrics.CPUCores)
	}
}

func TestMonitorService_GetPodMetrics(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
			Containers: []corev1.Container{
				{
					Name: "nginx",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	svc := NewMonitorService(client, "")

	metrics, err := svc.GetPodMetrics(context.Background(), "default", "nginx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.Name != "nginx" {
		t.Errorf("expected name 'nginx', got %s", metrics.Name)
	}

	if metrics.Namespace != "default" {
		t.Errorf("expected namespace 'default', got %s", metrics.Namespace)
	}

	if len(metrics.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(metrics.Containers))
	}

	if metrics.Containers[0].Name != "nginx" {
		t.Errorf("expected container name 'nginx', got %s", metrics.Containers[0].Name)
	}
}

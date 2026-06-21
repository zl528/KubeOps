package service

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestResourceService_ScaleDeployment(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "nginx"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx:latest"},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(deploy)
	svc := NewResourceService(client)

	err := svc.ScaleDeployment(context.Background(), "default", "nginx", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := client.AppsV1().Deployments("default").Get(context.Background(), "nginx", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *updated.Spec.Replicas != 5 {
		t.Errorf("expected 5 replicas, got %d", *updated.Spec.Replicas)
	}
}

func TestResourceService_ListConfigMaps(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-config",
			Namespace: "default",
			Labels:    map[string]string{"env": "test"},
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	client := fake.NewSimpleClientset(cm)
	svc := NewResourceService(client)

	result, err := svc.ListConfigMaps(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 configmap, got %d", len(result))
	}

	if result[0].Name != "my-config" {
		t.Errorf("expected name my-config, got %s", result[0].Name)
	}
	if len(result[0].Data) != 2 {
		t.Errorf("expected 2 data keys, got %d", len(result[0].Data))
	}
}

func TestResourceService_ListSecrets(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("secret123"),
		},
	}

	client := fake.NewSimpleClientset(secret)
	svc := NewResourceService(client)

	result, err := svc.ListSecrets(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(result))
	}

	if result[0].Name != "my-secret" {
		t.Errorf("expected name my-secret, got %s", result[0].Name)
	}
	if len(result[0].DataKeys) != 2 {
		t.Errorf("expected 2 data keys, got %d", len(result[0].DataKeys))
	}
}

func TestResourceService_ListResourceQuotas(t *testing.T) {
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compute-quota",
			Namespace: "default",
		},
		Status: corev1.ResourceQuotaStatus{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10"),
				corev1.ResourceMemory: resource.MustParse("20Gi"),
			},
			Used: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}

	client := fake.NewSimpleClientset(rq)
	svc := NewResourceService(client)

	result, err := svc.ListResourceQuotas(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 resource quota, got %d", len(result))
	}

	if result[0].Name != "compute-quota" {
		t.Errorf("expected name compute-quota, got %s", result[0].Name)
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

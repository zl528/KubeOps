package service

import (
	"context"
	"testing"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPolicyService_ListLimitRanges(t *testing.T) {
	lrs := &corev1.LimitRangeList{
		Items: []corev1.LimitRange{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mem-limit",
					Namespace: "default",
				},
				Spec: corev1.LimitRangeSpec{
					Limits: []corev1.LimitRangeItem{
						{
							Type: corev1.LimitTypeContainer,
							Max: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("512Mi"),
							},
							Min: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(lrs)
	svc := NewPolicyService(client)

	result, err := svc.ListLimitRanges(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 limit range, got %d", len(result))
	}

	if result[0].Name != "mem-limit" {
		t.Errorf("expected name mem-limit, got %s", result[0].Name)
	}

	if len(result[0].Limits) != 1 {
		t.Fatalf("expected 1 limit, got %d", len(result[0].Limits))
	}
}

func TestPolicyService_ListHPAs(t *testing.T) {
	hpas := &autoscalingv2.HorizontalPodAutoscalerList{
		Items: []autoscalingv2.HorizontalPodAutoscaler{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx-hpa",
					Namespace: "default",
				},
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						Kind: "Deployment",
						Name: "nginx",
					},
					MinReplicas: int32Ptr(1),
					MaxReplicas: 10,
					Metrics: []autoscalingv2.MetricSpec{
						{
							Type: autoscalingv2.ResourceMetricSourceType,
							Resource: &autoscalingv2.ResourceMetricSource{
								Name: corev1.ResourceCPU,
								Target: autoscalingv2.MetricTarget{
									Type:               autoscalingv2.UtilizationMetricType,
									AverageUtilization: int32Ptr(80),
								},
							},
						},
					},
				},
				Status: autoscalingv2.HorizontalPodAutoscalerStatus{
					CurrentReplicas: 3,
					DesiredReplicas: 3,
				},
			},
		},
	}

	client := fake.NewSimpleClientset(hpas)
	svc := NewPolicyService(client)

	result, err := svc.ListHPAs(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 hpa, got %d", len(result))
	}

	if result[0].Name != "nginx-hpa" {
		t.Errorf("expected name nginx-hpa, got %s", result[0].Name)
	}

	if result[0].MinReplicas != 1 {
		t.Errorf("expected min replicas 1, got %d", result[0].MinReplicas)
	}

	if result[0].MaxReplicas != 10 {
		t.Errorf("expected max replicas 10, got %d", result[0].MaxReplicas)
	}

	if result[0].Target != "Deployment/nginx" {
		t.Errorf("expected target Deployment/nginx, got %s", result[0].Target)
	}
}

func TestPolicyService_DeleteHPA(t *testing.T) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-hpa",
			Namespace: "default",
		},
	}

	client := fake.NewSimpleClientset(hpa)
	svc := NewPolicyService(client)

	err := svc.DeleteHPA(context.Background(), "default", "test-hpa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := svc.ListHPAs(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 hpas after delete, got %d", len(result))
	}
}

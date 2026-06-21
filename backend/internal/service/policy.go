package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PolicyService struct {
	client kubernetes.Interface
}

func NewPolicyService(client kubernetes.Interface) *PolicyService {
	return &PolicyService{client: client}
}

type LimitRangeInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Limits    []LimitRangeItem  `json:"limits"`
	Labels    map[string]string `json:"labels,omitempty"`
	Age       string            `json:"age"`
}

type LimitRangeItem struct {
	Type                 string `json:"type"`
	Min                  string `json:"min,omitempty"`
	Max                  string `json:"max,omitempty"`
	Default              string `json:"default,omitempty"`
	DefaultRequest       string `json:"defaultRequest,omitempty"`
	MaxLimitRequestRatio string `json:"maxLimitRequestRatio,omitempty"`
}

func (s *PolicyService) ListLimitRanges(ctx context.Context, namespace string) ([]LimitRangeInfo, error) {
	lrList, err := s.client.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []LimitRangeInfo
	for _, lr := range lrList.Items {
		info := LimitRangeInfo{
			Name:      lr.Name,
			Namespace: lr.Namespace,
			Labels:    lr.Labels,
			Age:       formatAge(lr.CreationTimestamp.Time),
		}

		for _, limit := range lr.Spec.Limits {
			item := LimitRangeItem{
				Type: string(limit.Type),
			}

			if limit.Min != nil {
				item.Min = limit.Min.Memory().String()
			}
			if limit.Max != nil {
				item.Max = limit.Max.Memory().String()
			}
			if limit.Default != nil {
				item.Default = limit.Default.Memory().String()
			}
			if limit.DefaultRequest != nil {
				item.DefaultRequest = limit.DefaultRequest.Memory().String()
			}

			info.Limits = append(info.Limits, item)
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *PolicyService) DeleteLimitRange(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().LimitRanges(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type CreateLimitRangeRequest struct {
	Namespace string              `json:"namespace"`
	Name      string              `json:"name"`
	Limits    []CreateLimitItem   `json:"limits"`
	Labels    map[string]string   `json:"labels,omitempty"`
}

type CreateLimitItem struct {
	Type                 string `json:"type"`
	MinCPU               string `json:"minCpu,omitempty"`
	MinMemory            string `json:"minMemory,omitempty"`
	MaxCPU               string `json:"maxCpu,omitempty"`
	MaxMemory            string `json:"maxMemory,omitempty"`
	DefaultCPU           string `json:"defaultCpu,omitempty"`
	DefaultMemory        string `json:"defaultMemory,omitempty"`
	DefaultRequestCPU    string `json:"defaultRequestCpu,omitempty"`
	DefaultRequestMemory string `json:"defaultRequestMemory,omitempty"`
	MaxLimitRequestRatio string `json:"maxLimitRequestRatio,omitempty"`
}

func (s *PolicyService) CreateLimitRange(ctx context.Context, req CreateLimitRangeRequest) error {
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
	}

	for _, item := range req.Limits {
		limit := corev1.LimitRangeItem{
			Type: corev1.LimitType(item.Type),
		}

		rl := corev1.ResourceList{}
		if item.MinCPU != "" {
			if q, err := resource.ParseQuantity(item.MinCPU); err == nil {
				rl[corev1.ResourceCPU] = q
			}
		}
		if item.MinMemory != "" {
			if q, err := resource.ParseQuantity(item.MinMemory); err == nil {
				rl[corev1.ResourceMemory] = q
			}
		}
		if len(rl) > 0 {
			limit.Min = rl
		}

		rl = corev1.ResourceList{}
		if item.MaxCPU != "" {
			if q, err := resource.ParseQuantity(item.MaxCPU); err == nil {
				rl[corev1.ResourceCPU] = q
			}
		}
		if item.MaxMemory != "" {
			if q, err := resource.ParseQuantity(item.MaxMemory); err == nil {
				rl[corev1.ResourceMemory] = q
			}
		}
		if len(rl) > 0 {
			limit.Max = rl
		}

		rl = corev1.ResourceList{}
		if item.DefaultCPU != "" {
			if q, err := resource.ParseQuantity(item.DefaultCPU); err == nil {
				rl[corev1.ResourceCPU] = q
			}
		}
		if item.DefaultMemory != "" {
			if q, err := resource.ParseQuantity(item.DefaultMemory); err == nil {
				rl[corev1.ResourceMemory] = q
			}
		}
		if len(rl) > 0 {
			limit.Default = rl
		}

		rl = corev1.ResourceList{}
		if item.DefaultRequestCPU != "" {
			if q, err := resource.ParseQuantity(item.DefaultRequestCPU); err == nil {
				rl[corev1.ResourceCPU] = q
			}
		}
		if item.DefaultRequestMemory != "" {
			if q, err := resource.ParseQuantity(item.DefaultRequestMemory); err == nil {
				rl[corev1.ResourceMemory] = q
			}
		}
		if len(rl) > 0 {
			limit.DefaultRequest = rl
		}

		lr.Spec.Limits = append(lr.Spec.Limits, limit)
	}

	_, err := s.client.CoreV1().LimitRanges(req.Namespace).Create(ctx, lr, metav1.CreateOptions{})
	return err
}

type HPAInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Target            string            `json:"target"`
	MinReplicas       int32             `json:"minReplicas"`
	MaxReplicas       int32             `json:"maxReplicas"`
	CurrentReplicas   int32             `json:"currentReplicas"`
	DesiredReplicas   int32             `json:"desiredReplicas"`
	Metrics           []HPAMetricInfo   `json:"metrics"`
	Labels            map[string]string `json:"labels,omitempty"`
	Age               string            `json:"age"`
}

type HPAMetricInfo struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Target     string `json:"target"`
	Current    string `json:"current,omitempty"`
}

func (s *PolicyService) ListHPAs(ctx context.Context, namespace string) ([]HPAInfo, error) {
	hpaList, err := s.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []HPAInfo
	for _, hpa := range hpaList.Items {
		info := HPAInfo{
			Name:            hpa.Name,
			Namespace:       hpa.Namespace,
			MinReplicas:     getMinReplicas(hpa.Spec.MinReplicas),
			MaxReplicas:     hpa.Spec.MaxReplicas,
			CurrentReplicas: getStatusCurrent(hpa.Status.CurrentReplicas),
			DesiredReplicas: getStatusDesired(hpa.Status.DesiredReplicas),
			Labels:          hpa.Labels,
			Age:             formatAge(hpa.CreationTimestamp.Time),
			Target:          fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name),
		}

		for _, metric := range hpa.Spec.Metrics {
			m := HPAMetricInfo{
				Type: string(metric.Type),
			}

			switch metric.Type {
			case autoscalingv2.ResourceMetricSourceType:
				if metric.Resource != nil {
					m.Name = string(metric.Resource.Name)
					if metric.Resource.Target.AverageUtilization != nil {
						m.Target = fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
					}
				}
			case autoscalingv2.PodsMetricSourceType:
				if metric.Pods != nil {
					m.Name = metric.Pods.Metric.Name
					if metric.Pods.Target.AverageValue != nil {
						m.Target = metric.Pods.Target.AverageValue.String()
					}
				}
			case autoscalingv2.ObjectMetricSourceType:
				if metric.Object != nil {
					m.Name = metric.Object.Metric.Name
					if metric.Object.Target.Value != nil {
						m.Target = metric.Object.Target.Value.String()
					}
				}
			}

			info.Metrics = append(info.Metrics, m)
		}

		for _, metric := range hpa.Status.CurrentMetrics {
			m := HPAMetricInfo{
				Type: string(metric.Type),
			}

			switch metric.Type {
			case autoscalingv2.ResourceMetricSourceType:
				if metric.Resource != nil {
					m.Name = string(metric.Resource.Name)
					m.Current = metric.Resource.Current.String()
				}
			case autoscalingv2.PodsMetricSourceType:
				if metric.Pods != nil {
					m.Name = metric.Pods.Metric.Name
					m.Current = metric.Pods.Current.String()
				}
			}

			info.Metrics = append(info.Metrics, m)
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *PolicyService) DeleteHPA(ctx context.Context, namespace, name string) error {
	return s.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type CreateHPARequest struct {
	Namespace         string            `json:"namespace"`
	Name              string            `json:"name"`
	TargetKind        string            `json:"targetKind"`
	TargetName        string            `json:"targetName"`
	MinReplicas       int32             `json:"minReplicas"`
	MaxReplicas       int32             `json:"maxReplicas"`
	MetricName        string            `json:"metricName"`
	MetricTarget      string            `json:"metricTarget"`
	Labels            map[string]string `json:"labels,omitempty"`
	Behavior          *HPABehavior      `json:"behavior,omitempty"`
}

type HPABehavior struct {
	ScaleUp  *HPAScalingRules `json:"scaleUp,omitempty"`
	ScaleDown *HPAScalingRules `json:"scaleDown,omitempty"`
}

type HPAScalingRules struct {
	StabilizationWindowSeconds *int32              `json:"stabilizationWindowSeconds,omitempty"`
	SelectPolicy               string              `json:"selectPolicy,omitempty"`
	Policies                   []HPAScalingPolicy  `json:"policies,omitempty"`
}

type HPAScalingPolicy struct {
	Type          string `json:"type"`
	Value         int32  `json:"value"`
	PeriodSeconds int32  `json:"periodSeconds"`
}

func (s *PolicyService) CreateHPA(ctx context.Context, req CreateHPARequest) error {
	minReplicas := &req.MinReplicas

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       req.TargetKind,
				Name:       req.TargetName,
				APIVersion: "apps/v1",
			},
			MinReplicas: minReplicas,
			MaxReplicas: req.MaxReplicas,
		},
	}

	if req.MetricTarget != "" {
		metricName := req.MetricName
		if metricName == "" {
			metricName = "cpu"
		}
		if q, err := resource.ParseQuantity(req.MetricTarget); err == nil {
			avgUtil := int32(q.Value())
			hpa.Spec.Metrics = []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceName(metricName),
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: &avgUtil,
						},
					},
				},
			}
		}
	}

	if req.Behavior != nil {
		hpa.Spec.Behavior = &autoscalingv2.HorizontalPodAutoscalerBehavior{}
		if req.Behavior.ScaleUp != nil {
			hpa.Spec.Behavior.ScaleUp = convertScalingRules(req.Behavior.ScaleUp)
		}
		if req.Behavior.ScaleDown != nil {
			hpa.Spec.Behavior.ScaleDown = convertScalingRules(req.Behavior.ScaleDown)
		}
	}

	_, err := s.client.AutoscalingV2().HorizontalPodAutoscalers(req.Namespace).Create(ctx, hpa, metav1.CreateOptions{})
	return err
}

func convertScalingRules(rules *HPAScalingRules) *autoscalingv2.HPAScalingRules {
	result := &autoscalingv2.HPAScalingRules{
		StabilizationWindowSeconds: rules.StabilizationWindowSeconds,
	}
	if rules.SelectPolicy != "" {
		sp := autoscalingv2.ScalingPolicySelect(rules.SelectPolicy)
		result.SelectPolicy = &sp
	}
	for _, p := range rules.Policies {
		result.Policies = append(result.Policies, autoscalingv2.HPAScalingPolicy{
			Type:          autoscalingv2.HPAScalingPolicyType(p.Type),
			Value:         p.Value,
			PeriodSeconds: p.PeriodSeconds,
		})
	}
	return result
}

func getMinReplicas(min *int32) int32 {
	if min != nil {
		return *min
	}
	return 1
}

func getStatusCurrent(current int32) int32 {
	return current
}

func getStatusDesired(desired int32) int32 {
	return desired
}

package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeops/ops-kubernetes/internal/model"
)

type ClusterService struct {
	client kubernetes.Interface
}

func NewClusterService(client kubernetes.Interface) *ClusterService {
	return &ClusterService{client: client}
}

func (s *ClusterService) GetOverview(ctx context.Context) (*model.ClusterOverview, error) {
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	nsList, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	pods, err := s.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	overview := &model.ClusterOverview{
		TotalNodes: len(nodes.Items),
		Namespaces: len(nsList.Items),
		TotalPods:  len(pods.Items),
	}

	for _, node := range nodes.Items {
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				overview.ReadyNodes++
			}
		}
	}

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			overview.RunningPods++
		case corev1.PodPending:
			overview.PendingPods++
		case corev1.PodFailed:
			overview.FailedPods++
		}
	}

	var totalCPU, usedCPU, totalMem, usedMem resource.Quantity
	for _, node := range nodes.Items {
		if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
			totalCPU.Add(cpu)
		}
		if mem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
			totalMem.Add(mem)
		}
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				usedCPU.Add(cpu)
			}
			if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				usedMem.Add(mem)
			}
		}
	}

	totalCPUCores := float64(totalCPU.MilliValue()) / 1000.0
	usedCPUCores := float64(usedCPU.MilliValue()) / 1000.0
	totalMemGB := float64(totalMem.Value()) / (1024 * 1024 * 1024)
	usedMemGB := float64(usedMem.Value()) / (1024 * 1024 * 1024)

	cpuPct := 0.0
	if totalCPUCores > 0 {
		cpuPct = (usedCPUCores / totalCPUCores) * 100
	}
	memPct := 0.0
	if totalMemGB > 0 {
		memPct = (usedMemGB / totalMemGB) * 100
	}

	overview.CPUUsage = model.ResourceUsage{
		Used:  fmt.Sprintf("%.2f cores", usedCPUCores),
		Total: fmt.Sprintf("%.2f cores", totalCPUCores),
		Pct:   fmt.Sprintf("%.1f%%", cpuPct),
	}
	overview.MemoryUsage = model.ResourceUsage{
		Used:  fmt.Sprintf("%.2f GB", usedMemGB),
		Total: fmt.Sprintf("%.2f GB", totalMemGB),
		Pct:   fmt.Sprintf("%.1f%%", memPct),
	}

	return overview, nil
}

func (s *ClusterService) ListNodes(ctx context.Context) ([]model.NodeInfo, error) {
	nodes, err := s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []model.NodeInfo
	for _, node := range nodes.Items {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
			}
		}

		roles := []string{}
		for key := range node.Labels {
			switch key {
			case "node-role.kubernetes.io/control-plane", "node-role.kubernetes.io/master":
				roles = append(roles, "control-plane")
			case "node-role.kubernetes.io/worker":
				roles = append(roles, "worker")
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "<none>")
		}

		result = append(result, model.NodeInfo{
			Name:          node.Name,
			Status:        status,
			Roles:         roles,
			Version:       node.Status.NodeInfo.KubeletVersion,
			OS:            node.Status.NodeInfo.OperatingSystem,
			KernelVersion: node.Status.NodeInfo.KernelVersion,
			CPU:           node.Status.Allocatable.Cpu().String(),
			Memory:        node.Status.Allocatable.Memory().String(),
			Labels:        node.Labels,
		})
	}

	return result, nil
}

func (s *ClusterService) GetNode(ctx context.Context, name string) (*model.NodeInfo, error) {
	node, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	status := "NotReady"
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
			status = "Ready"
		}
	}

	return &model.NodeInfo{
		Name:          node.Name,
		Status:        status,
		Version:       node.Status.NodeInfo.KubeletVersion,
		OS:            node.Status.NodeInfo.OperatingSystem,
		KernelVersion: node.Status.NodeInfo.KernelVersion,
		CPU:           node.Status.Allocatable.Cpu().String(),
		Memory:        node.Status.Allocatable.Memory().String(),
		Labels:        node.Labels,
	}, nil
}

func (s *ClusterService) ListNamespaces(ctx context.Context) ([]model.NamespaceInfo, error) {
	nsList, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []model.NamespaceInfo
	for _, ns := range nsList.Items {
		result = append(result, model.NamespaceInfo{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Labels: ns.Labels,
			Age:    formatAge(ns.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *ClusterService) ListPods(ctx context.Context, namespace string) ([]model.PodInfo, error) {
	pods, err := s.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []model.PodInfo
	for _, pod := range pods.Items {
		var containers []model.ContainerInfo
		for _, c := range pod.Spec.Containers {
			ready := false
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == c.Name {
					ready = cs.Ready
					break
				}
			}
			containers = append(containers, model.ContainerInfo{
				Name:  c.Name,
				Image: c.Image,
				Ready: ready,
			})
		}

		var restarts int32
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}

		result = append(result, model.PodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     string(pod.Status.Phase),
			Node:       pod.Spec.NodeName,
			Restarts:   restarts,
			Age:        formatAge(pod.CreationTimestamp.Time),
			Containers: containers,
		})
	}
	return result, nil
}

func (s *ClusterService) ListDeployments(ctx context.Context, namespace string) ([]model.DeploymentInfo, error) {
	deploys, err := s.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []model.DeploymentInfo
	for _, d := range deploys.Items {
		result = append(result, model.DeploymentInfo{
			Name:      d.Name,
			Namespace: d.Namespace,
			Replicas:  fmt.Sprintf("%d/%d", *d.Spec.Replicas, d.Status.Replicas),
			Ready:     fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas),
			UpToDate:  d.Status.UpdatedReplicas,
			Available: d.Status.AvailableReplicas,
			Age:       formatAge(d.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *ClusterService) ListServices(ctx context.Context, namespace string) ([]model.ServiceInfo, error) {
	svcList, err := s.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []model.ServiceInfo
	for _, svc := range svcList.Items {
		var ports []model.ServicePortInfo
		for _, p := range svc.Spec.Ports {
			ports = append(ports, model.ServicePortInfo{
				Port:       p.Port,
				TargetPort: p.TargetPort.IntVal,
				Protocol:   string(p.Protocol),
				NodePort:   p.NodePort,
			})
		}

		result = append(result, model.ServiceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Age:       formatAge(svc.CreationTimestamp.Time),
		})
	}
	return result, nil
}

type ServiceDetail struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"clusterIP"`
	Ports       []model.ServicePortInfo `json:"ports"`
	Selector    map[string]string `json:"selector"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Age         string            `json:"age"`
}

func (s *ClusterService) GetService(ctx context.Context, namespace, name string) (*ServiceDetail, error) {
	svc, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var ports []model.ServicePortInfo
	for _, p := range svc.Spec.Ports {
		ports = append(ports, model.ServicePortInfo{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: p.TargetPort.IntVal,
			Protocol:   string(p.Protocol),
			NodePort:   p.NodePort,
		})
	}

	return &ServiceDetail{
		Name:        svc.Name,
		Namespace:   svc.Namespace,
		Type:        string(svc.Spec.Type),
		ClusterIP:   svc.Spec.ClusterIP,
		Ports:       ports,
		Selector:    svc.Spec.Selector,
		Labels:      svc.Labels,
		Annotations: svc.Annotations,
		Age:         formatAge(svc.CreationTimestamp.Time),
	}, nil
}

func (s *ClusterService) DeleteService(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *ClusterService) ListEvents(ctx context.Context, namespace string) ([]model.EventInfo, error) {
	events, err := s.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []model.EventInfo
	for _, e := range events.Items {
		result = append(result, model.EventInfo{
			Type:     e.Type,
			Reason:   e.Reason,
			Message:  e.Message,
			Count:    e.Count,
			LastTime: e.LastTimestamp.Time.Format(time.RFC3339),
			Object:   fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
		})
	}
	return result, nil
}

func (s *ClusterService) DrainNode(ctx context.Context, name string) error {
	node, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	if node.Spec.Unschedulable {
		return fmt.Errorf("node %s is already cordoned", name)
	}

	node.Spec.Unschedulable = true
	_, err = s.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cordon node: %w", err)
	}

	podList, err := s.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", name),
	})
	if err != nil {
		return fmt.Errorf("list pods on node: %w", err)
	}

	for _, pod := range podList.Items {
		if pod.DeletionTimestamp != nil {
			continue
		}
		if pod.Spec.NodeName != name {
			continue
		}
		if pod.Namespace == "kube-system" {
			continue
		}
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations["graceful-shutdown"] = "true"
		_ = s.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: int64Ptr(30),
		})
	}

	return nil
}

func int64Ptr(i int64) *int64 { return &i }

func (s *ClusterService) UncordonNode(ctx context.Context, name string) error {
	node, err := s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	node.Spec.Unschedulable = false
	_, err = s.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("uncordon node: %w", err)
	}

	return nil
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

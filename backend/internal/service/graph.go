package service

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type GraphService struct {
	client kubernetes.Interface
}

func NewGraphService(client kubernetes.Interface) *GraphService {
	return &GraphService{client: client}
}

type GraphNode struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Labels     map[string]string `json:"labels,omitempty"`
	Ports      []int32           `json:"ports,omitempty"`
	IP         string            `json:"ip,omitempty"`
	Replicas   string            `json:"replicas,omitempty"`
	Children   []string          `json:"children,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Label  string `json:"label,omitempty"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

func (s *GraphService) GetGraph(ctx context.Context, namespace string) (*GraphData, error) {
	graph := &GraphData{
		Nodes: []GraphNode{},
		Edges: []GraphEdge{},
	}

	// Fetch Deployments
	deploys, err := s.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, d := range deploys.Items {
			node := GraphNode{
				ID:        "deployment-" + d.Namespace + "-" + d.Name,
				Type:      "deployment",
				Name:      d.Name,
				Namespace: d.Namespace,
				Labels:    d.Labels,
			}
			if d.Spec.Replicas != nil {
				node.Replicas = formatReplicas(d.Status.ReadyReplicas, *d.Spec.Replicas)
			}
			if d.Status.ReadyReplicas == *d.Spec.Replicas {
				node.Status = "running"
			} else {
				node.Status = "warning"
			}
			graph.Nodes = append(graph.Nodes, node)
		}
	}

	// Fetch StatefulSets
	stss, err := s.client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, sts := range stss.Items {
			node := GraphNode{
				ID:        "statefulset-" + sts.Namespace + "-" + sts.Name,
				Type:      "statefulset",
				Name:      sts.Name,
				Namespace: sts.Namespace,
				Labels:    sts.Labels,
			}
			if sts.Spec.Replicas != nil {
				node.Replicas = formatReplicas(sts.Status.ReadyReplicas, *sts.Spec.Replicas)
			}
			if sts.Status.ReadyReplicas == *sts.Spec.Replicas {
				node.Status = "running"
			} else {
				node.Status = "warning"
			}
			graph.Nodes = append(graph.Nodes, node)
		}
	}

	// Fetch DaemonSets
	dss, err := s.client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ds := range dss.Items {
			node := GraphNode{
				ID:        "daemonset-" + ds.Namespace + "-" + ds.Name,
				Type:      "daemonset",
				Name:      ds.Name,
				Namespace: ds.Namespace,
				Labels:    ds.Labels,
				Replicas:  formatReplicas(ds.Status.NumberReady, ds.Status.DesiredNumberScheduled),
			}
			if ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
				node.Status = "running"
			} else {
				node.Status = "warning"
			}
			graph.Nodes = append(graph.Nodes, node)
		}
	}

	// Fetch Services
	svcs, err := s.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, svc := range svcs.Items {
			node := GraphNode{
				ID:        "service-" + svc.Namespace + "-" + svc.Name,
				Type:      "service",
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Labels:    svc.Labels,
				IP:        svc.Spec.ClusterIP,
			}
			for _, p := range svc.Spec.Ports {
				node.Ports = append(node.Ports, p.Port)
			}
			graph.Nodes = append(graph.Nodes, node)

			// Connect Service to Pods via selector
			if len(svc.Spec.Selector) > 0 {
				pods, err := s.client.CoreV1().Pods(svc.Namespace).List(ctx, metav1.ListOptions{})
				if err == nil {
					for _, pod := range pods.Items {
						if matchLabels(pod.Labels, svc.Spec.Selector) {
							graph.Edges = append(graph.Edges, GraphEdge{
								Source: node.ID,
								Target: "pod-" + pod.Namespace + "-" + pod.Name,
								Type:   "selector",
								Label:  "选择器",
							})
						}
					}
				}
			}
		}
	}

	// Fetch Ingresses
	ings, err := s.client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ing := range ings.Items {
			node := GraphNode{
				ID:        "ingress-" + ing.Namespace + "-" + ing.Name,
				Type:      "ingress",
				Name:      ing.Name,
				Namespace: ing.Namespace,
				Labels:    ing.Labels,
			}
			for _, rule := range ing.Spec.Rules {
				if rule.Host != "" {
					node.IP = rule.Host
					break
				}
			}
			graph.Nodes = append(graph.Nodes, node)

			// Connect Ingress to Services
			for _, rule := range ing.Spec.Rules {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil {
						graph.Edges = append(graph.Edges, GraphEdge{
							Source: node.ID,
							Target: "service-" + ing.Namespace + "-" + path.Backend.Service.Name,
							Type:   "routing",
							Label:  rule.Host + path.Path,
						})
					}
				}
			}
		}
	}

	// Build ReplicaSet -> Deployment mapping
	rsToDeployment := map[string]string{}
	replicaSets, rsErr := s.client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if rsErr == nil {
		for _, rs := range replicaSets.Items {
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" {
					rsToDeployment[rs.Name] = owner.Name
				}
			}
		}
	}

	// Fetch Pods
	pods, err := s.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pod := range pods.Items {
			node := GraphNode{
				ID:        "pod-" + pod.Namespace + "-" + pod.Name,
				Type:      "pod",
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Labels:    pod.Labels,
				IP:        pod.Status.PodIP,
			}
			switch pod.Status.Phase {
			case "Running":
				node.Status = "running"
			case "Pending":
				node.Status = "pending"
			case "Failed":
				node.Status = "error"
			default:
				node.Status = "unknown"
			}
			graph.Nodes = append(graph.Nodes, node)

			// Find owner (Deployment/StatefulSet/DaemonSet)
			for _, owner := range pod.OwnerReferences {
				var ownerType string
				var ownerName string
				switch owner.Kind {
				case "ReplicaSet":
					// ReplicaSet -> find parent Deployment
					deployName, ok := rsToDeployment[owner.Name]
					if ok {
						ownerType = "deployment"
						ownerName = deployName
					} else {
						// Fallback: strip hash from ReplicaSet name
						ownerType = "deployment"
						ownerName = stripRSHash(owner.Name)
					}
				case "StatefulSet":
					ownerType = "statefulset"
					ownerName = owner.Name
				case "DaemonSet":
					ownerType = "daemonset"
					ownerName = owner.Name
				case "Job":
					ownerType = "job"
					ownerName = owner.Name
				}
				if ownerType != "" {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: ownerType + "-" + pod.Namespace + "-" + ownerName,
						Target: node.ID,
						Type:   "ownership",
						Label:  "创建",
					})
				}
			}

			// Connect Pod to PVC
			for _, vol := range pod.Spec.Volumes {
				if vol.PersistentVolumeClaim != nil {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: node.ID,
						Target: "pvc-" + pod.Namespace + "-" + vol.PersistentVolumeClaim.ClaimName,
						Type:   "volume",
						Label:  "挂载",
					})
				}
			}

			// Connect Pod to ConfigMap/Secret
			for _, vol := range pod.Spec.Volumes {
				if vol.ConfigMap != nil {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: "configmap-" + pod.Namespace + "-" + vol.ConfigMap.Name,
						Target: node.ID,
						Type:   "config",
						Label:  "配置",
					})
				}
				if vol.Secret != nil {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: "secret-" + pod.Namespace + "-" + vol.Secret.SecretName,
						Target: node.ID,
						Type:   "config",
						Label:  "密钥",
					})
				}
			}
		}
	}

	// Fetch ConfigMaps
	cms, err := s.client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, cm := range cms.Items {
			graph.Nodes = append(graph.Nodes, GraphNode{
				ID:        "configmap-" + cm.Namespace + "-" + cm.Name,
				Type:      "configmap",
				Name:      cm.Name,
				Namespace: cm.Namespace,
				Labels:    cm.Labels,
			})
		}
	}

	// Fetch Secrets
	secrets, err := s.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, secret := range secrets.Items {
			graph.Nodes = append(graph.Nodes, GraphNode{
				ID:        "secret-" + secret.Namespace + "-" + secret.Name,
				Type:      "secret",
				Name:      secret.Name,
				Namespace: secret.Namespace,
				Labels:    secret.Labels,
			})
		}
	}

	// Fetch PVCs
	pvcs, err := s.client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pvc := range pvcs.Items {
			graph.Nodes = append(graph.Nodes, GraphNode{
				ID:        "pvc-" + pvc.Namespace + "-" + pvc.Name,
				Type:      "pvc",
				Name:      pvc.Name,
				Namespace: pvc.Namespace,
				Labels:    pvc.Labels,
				Status:    string(pvc.Status.Phase),
			})
		}
	}

	// Fetch HPAs
	hpas, err := s.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, hpa := range hpas.Items {
			node := GraphNode{
				ID:        "hpa-" + hpa.Namespace + "-" + hpa.Name,
				Type:      "hpa",
				Name:      hpa.Name,
				Namespace: hpa.Namespace,
				Labels:    hpa.Labels,
			}
			graph.Nodes = append(graph.Nodes, node)

			// Connect HPA to target
			targetRef := hpa.Spec.ScaleTargetRef
			graph.Edges = append(graph.Edges, GraphEdge{
				Source: node.ID,
				Target: targetRef.Kind + "-" + hpa.Namespace + "-" + hpa.Name,
				Type:   "autoscale",
				Label:  "自动扩缩",
			})
		}
	}

	return graph, nil
}

func formatReplicas(ready, desired int32) string {
	return fmt.Sprintf("%d/%d", ready, desired)
}

// stripRSHash removes the hash suffix from ReplicaSet name
// e.g., "hello-app-c5558d946" -> "hello-app"
func stripRSHash(rsName string) string {
	// Find the last dash followed by a hash (10 chars alphanumeric)
	for i := len(rsName) - 1; i >= 0; i-- {
		if rsName[i] == '-' && i < len(rsName)-1 {
			// Check if the part after dash looks like a hash
			suffix := rsName[i+1:]
			if len(suffix) >= 8 && len(suffix) <= 12 {
				isHash := true
				for _, c := range suffix {
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
						isHash = false
						break
					}
				}
				if isHash {
					return rsName[:i]
				}
			}
		}
	}
	return rsName
}

func matchLabels(podLabels, selectorLabels map[string]string) bool {
	for k, v := range selectorLabels {
		if podLabels[k] != v {
			return false
		}
	}
	return true
}

package service

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type NetworkService struct {
	client kubernetes.Interface
}

func NewNetworkService(client kubernetes.Interface) *NetworkService {
	return &NetworkService{client: client}
}

type IngressInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Hosts     []string          `json:"hosts"`
	TLS       bool              `json:"tls"`
	ClassName string            `json:"className"`
	Labels    map[string]string `json:"labels,omitempty"`
	Age       string            `json:"age"`
}

func (s *NetworkService) ListIngresses(ctx context.Context, namespace string) ([]IngressInfo, error) {
	ingList, err := s.client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []IngressInfo
	for _, ing := range ingList.Items {
		var hosts []string
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}

		hasTLS := len(ing.Spec.TLS) > 0

		var className string
		if ing.Spec.IngressClassName != nil {
			className = *ing.Spec.IngressClassName
		}

		result = append(result, IngressInfo{
			Name:      ing.Name,
			Namespace: ing.Namespace,
			Hosts:     hosts,
			TLS:       hasTLS,
			ClassName: className,
			Labels:    ing.Labels,
			Age:       formatAge(ing.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *NetworkService) DeleteIngress(ctx context.Context, namespace, name string) error {
	return s.client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type IngressDetail struct {
	Name           string               `json:"name"`
	Namespace      string               `json:"namespace"`
	ClassName      string               `json:"className"`
	Labels         map[string]string    `json:"labels,omitempty"`
	Annotations    map[string]string    `json:"annotations,omitempty"`
	Hosts          []IngressHost        `json:"hosts"`
	TLS            []IngressTLS         `json:"tls"`
	DefaultBackend *IngressDefaultBackend `json:"defaultBackend,omitempty"`
}

type IngressHost struct {
	Host  string          `json:"host"`
	Paths []IngressPath   `json:"paths"`
}

type IngressPath struct {
	Path        string `json:"path"`
	PathType    string `json:"pathType"`
	ServiceName string `json:"serviceName"`
	ServicePort int32  `json:"servicePort"`
}

type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

type IngressDefaultBackend struct {
	ServiceName string `json:"serviceName"`
	ServicePort int32  `json:"servicePort"`
}

func (s *NetworkService) GetIngress(ctx context.Context, namespace, name string) (*IngressDetail, error) {
	ing, err := s.client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var className string
	if ing.Spec.IngressClassName != nil {
		className = *ing.Spec.IngressClassName
	}

	detail := &IngressDetail{
		Name:        ing.Name,
		Namespace:   ing.Namespace,
		ClassName:   className,
		Labels:      ing.Labels,
		Annotations: ing.Annotations,
		Hosts:       make([]IngressHost, 0),
		TLS:         make([]IngressTLS, 0),
	}

	hostMap := make(map[string]*IngressHost)
	for _, rule := range ing.Spec.Rules {
		host := rule.Host
		if hostMap[host] == nil {
			hostMap[host] = &IngressHost{Host: host, Paths: make([]IngressPath, 0)}
		}
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				var serviceName string
				var servicePort int32
				if path.Backend.Service != nil {
					serviceName = path.Backend.Service.Name
					if path.Backend.Service.Port.Number != 0 {
						servicePort = path.Backend.Service.Port.Number
					}
				}
				pathType := ""
				if path.PathType != nil {
					pathType = string(*path.PathType)
				}
				if pathType == "" {
					pathType = "Prefix"
				}
				hostMap[host].Paths = append(hostMap[host].Paths, IngressPath{
					Path:        path.Path,
					PathType:    pathType,
					ServiceName: serviceName,
					ServicePort: servicePort,
				})
			}
		}
	}
	for _, h := range hostMap {
		detail.Hosts = append(detail.Hosts, *h)
	}

	for _, tls := range ing.Spec.TLS {
		detail.TLS = append(detail.TLS, IngressTLS{
			Hosts:      tls.Hosts,
			SecretName: tls.SecretName,
		})
	}

	if ing.Spec.DefaultBackend != nil {
		db := &IngressDefaultBackend{}
		if ing.Spec.DefaultBackend.Service != nil {
			db.ServiceName = ing.Spec.DefaultBackend.Service.Name
			if ing.Spec.DefaultBackend.Service.Port.Number != 0 {
				db.ServicePort = ing.Spec.DefaultBackend.Service.Port.Number
			}
		}
		if db.ServiceName != "" {
			detail.DefaultBackend = db
		}
	}

	return detail, nil
}

type CreateIngressRequest struct {
	Namespace      string               `json:"namespace"`
	Name           string               `json:"name"`
	ClassName      string               `json:"className"`
	Labels         map[string]string    `json:"labels,omitempty"`
	Annotations    map[string]string    `json:"annotations,omitempty"`
	Hosts          []IngressHost        `json:"hosts"`
	TLS            []IngressTLS         `json:"tls"`
	DefaultBackend *IngressDefaultBackend `json:"defaultBackend,omitempty"`
}

func (s *NetworkService) CreateIngress(ctx context.Context, req CreateIngressRequest) error {
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: make([]networkingv1.IngressRule, 0),
			TLS:   make([]networkingv1.IngressTLS, 0),
		},
	}

	if req.ClassName != "" {
		ing.Spec.IngressClassName = &req.ClassName
	}

	for _, host := range req.Hosts {
		rule := networkingv1.IngressRule{
			Host: host.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: make([]networkingv1.HTTPIngressPath, 0),
				},
			},
		}
		for _, path := range host.Paths {
			pathType := networkingv1.PathType(path.PathType)
			if pathType == "" {
				pathType = networkingv1.PathTypePrefix
			}
			httpPath := networkingv1.HTTPIngressPath{
				Path:     path.Path,
				PathType: &pathType,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: path.ServiceName,
						Port: networkingv1.ServiceBackendPort{
							Number: path.ServicePort,
						},
					},
				},
			}
			rule.HTTP.Paths = append(rule.HTTP.Paths, httpPath)
		}
		ing.Spec.Rules = append(ing.Spec.Rules, rule)
	}

	for _, tls := range req.TLS {
		ing.Spec.TLS = append(ing.Spec.TLS, networkingv1.IngressTLS{
			Hosts:      tls.Hosts,
			SecretName: tls.SecretName,
		})
	}

	if req.DefaultBackend != nil && req.DefaultBackend.ServiceName != "" {
		ing.Spec.DefaultBackend = &networkingv1.IngressBackend{
			Service: &networkingv1.IngressServiceBackend{
				Name: req.DefaultBackend.ServiceName,
				Port: networkingv1.ServiceBackendPort{
					Number: req.DefaultBackend.ServicePort,
				},
			},
		}
	}

	_, err := s.client.NetworkingV1().Ingresses(req.Namespace).Create(ctx, ing, metav1.CreateOptions{})
	return err
}

type NetworkPolicyInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	PodSelector map[string]string `json:"podSelector"`
	Ingress   []NetworkPolicyRule `json:"ingress"`
	Egress    []NetworkPolicyRule `json:"egress"`
	Labels    map[string]string `json:"labels,omitempty"`
	Age       string            `json:"age"`
}

type NetworkPolicyRule struct {
	From []NetworkPolicyPeer `json:"from,omitempty"`
	To   []NetworkPolicyPeer `json:"to,omitempty"`
	Ports []NetworkPolicyPort `json:"ports,omitempty"`
}

type NetworkPolicyPeer struct {
	NamespaceSelector map[string]string `json:"namespaceSelector,omitempty"`
	PodSelector       map[string]string `json:"podSelector,omitempty"`
	IPBlock           string            `json:"ipBlock,omitempty"`
}

type NetworkPolicyPort struct {
	Port     string `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

func (s *NetworkService) ListNetworkPolicies(ctx context.Context, namespace string) ([]NetworkPolicyInfo, error) {
	npList, err := s.client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []NetworkPolicyInfo
	for _, np := range npList.Items {
		info := NetworkPolicyInfo{
			Name:      np.Name,
			Namespace: np.Namespace,
			Labels:    np.Labels,
			Age:       formatAge(np.CreationTimestamp.Time),
		}

		if np.Spec.PodSelector.MatchLabels != nil {
			info.PodSelector = np.Spec.PodSelector.MatchLabels
		}

		for _, rule := range np.Spec.Ingress {
			ingressRule := NetworkPolicyRule{}
			for _, from := range rule.From {
				peer := NetworkPolicyPeer{}
				if from.NamespaceSelector != nil && from.NamespaceSelector.MatchLabels != nil {
					peer.NamespaceSelector = from.NamespaceSelector.MatchLabels
				}
				if from.PodSelector != nil && from.PodSelector.MatchLabels != nil {
					peer.PodSelector = from.PodSelector.MatchLabels
				}
				if from.IPBlock != nil && from.IPBlock.CIDR != "" {
					peer.IPBlock = from.IPBlock.CIDR
				}
				ingressRule.From = append(ingressRule.From, peer)
			}
			for _, port := range rule.Ports {
				p := NetworkPolicyPort{}
				if port.Port != nil {
					p.Port = port.Port.String()
				}
				if port.Protocol != nil {
					p.Protocol = string(*port.Protocol)
				}
				ingressRule.Ports = append(ingressRule.Ports, p)
			}
			info.Ingress = append(info.Ingress, ingressRule)
		}

		for _, rule := range np.Spec.Egress {
			egressRule := NetworkPolicyRule{}
			for _, to := range rule.To {
				peer := NetworkPolicyPeer{}
				if to.NamespaceSelector != nil && to.NamespaceSelector.MatchLabels != nil {
					peer.NamespaceSelector = to.NamespaceSelector.MatchLabels
				}
				if to.PodSelector != nil && to.PodSelector.MatchLabels != nil {
					peer.PodSelector = to.PodSelector.MatchLabels
				}
				if to.IPBlock != nil && to.IPBlock.CIDR != "" {
					peer.IPBlock = to.IPBlock.CIDR
				}
				egressRule.To = append(egressRule.To, peer)
			}
			for _, port := range rule.Ports {
				p := NetworkPolicyPort{}
				if port.Port != nil {
					p.Port = port.Port.String()
				}
				if port.Protocol != nil {
					p.Protocol = string(*port.Protocol)
				}
				egressRule.Ports = append(egressRule.Ports, p)
			}
			info.Egress = append(info.Egress, egressRule)
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *NetworkService) DeleteNetworkPolicy(ctx context.Context, namespace, name string) error {
	return s.client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *NetworkService) GetNetworkPolicy(ctx context.Context, namespace, name string) (*NetworkPolicyDetail, error) {
	np, err := s.client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	detail := &NetworkPolicyDetail{
		Name:      np.Name,
		Namespace: np.Namespace,
		Labels:    np.Labels,
	}

	if np.Spec.PodSelector.MatchLabels != nil {
		detail.PodSelector = np.Spec.PodSelector.MatchLabels
	}

	for _, pt := range np.Spec.PolicyTypes {
		detail.PolicyTypes = append(detail.PolicyTypes, string(pt))
	}

	for _, rule := range np.Spec.Ingress {
		ingressRule := NetworkPolicyRule{}
		for _, from := range rule.From {
			peer := NetworkPolicyPeer{}
			if from.NamespaceSelector != nil && from.NamespaceSelector.MatchLabels != nil {
				peer.NamespaceSelector = from.NamespaceSelector.MatchLabels
			}
			if from.PodSelector != nil && from.PodSelector.MatchLabels != nil {
				peer.PodSelector = from.PodSelector.MatchLabels
			}
			if from.IPBlock != nil && from.IPBlock.CIDR != "" {
				peer.IPBlock = from.IPBlock.CIDR
			}
			ingressRule.From = append(ingressRule.From, peer)
		}
		for _, port := range rule.Ports {
			p := NetworkPolicyPort{}
			if port.Port != nil {
				p.Port = port.Port.String()
			}
			if port.Protocol != nil {
				p.Protocol = string(*port.Protocol)
			}
			ingressRule.Ports = append(ingressRule.Ports, p)
		}
		detail.Ingress = append(detail.Ingress, ingressRule)
	}

	for _, rule := range np.Spec.Egress {
		egressRule := NetworkPolicyRule{}
		for _, to := range rule.To {
			peer := NetworkPolicyPeer{}
			if to.NamespaceSelector != nil && to.NamespaceSelector.MatchLabels != nil {
				peer.NamespaceSelector = to.NamespaceSelector.MatchLabels
			}
			if to.PodSelector != nil && to.PodSelector.MatchLabels != nil {
				peer.PodSelector = to.PodSelector.MatchLabels
			}
			if to.IPBlock != nil && to.IPBlock.CIDR != "" {
				peer.IPBlock = to.IPBlock.CIDR
			}
			egressRule.To = append(egressRule.To, peer)
		}
		for _, port := range rule.Ports {
			p := NetworkPolicyPort{}
			if port.Port != nil {
				p.Port = port.Port.String()
			}
			if port.Protocol != nil {
				p.Protocol = string(*port.Protocol)
			}
			egressRule.Ports = append(egressRule.Ports, p)
		}
		detail.Egress = append(detail.Egress, egressRule)
	}

	return detail, nil
}

type NetworkPolicyDetail struct {
	Name         string              `json:"name"`
	Namespace    string              `json:"namespace"`
	PodSelector  map[string]string   `json:"podSelector,omitempty"`
	Ingress      []NetworkPolicyRule `json:"ingress,omitempty"`
	Egress       []NetworkPolicyRule `json:"egress,omitempty"`
	Labels       map[string]string   `json:"labels,omitempty"`
	PolicyTypes  []string            `json:"policyTypes,omitempty"`
}

type CreateNetworkPolicyRequest struct {
	Name        string              `json:"name"`
	Namespace   string              `json:"namespace"`
	PodSelector map[string]string   `json:"podSelector,omitempty"`
	Ingress     []NetworkPolicyRule `json:"ingress,omitempty"`
	Egress      []NetworkPolicyRule `json:"egress,omitempty"`
	Labels      map[string]string   `json:"labels,omitempty"`
	PolicyTypes []string            `json:"policyTypes,omitempty"`
}

func (s *NetworkService) CreateNetworkPolicy(ctx context.Context, req CreateNetworkPolicyRequest) error {
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    req.Labels,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: req.PodSelector,
			},
			PolicyTypes: []networkingv1.PolicyType{},
		},
	}

	hasIngress := len(req.Ingress) > 0
	hasEgress := len(req.Egress) > 0

	if len(req.PolicyTypes) > 0 {
		for _, pt := range req.PolicyTypes {
			np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, networkingv1.PolicyType(pt))
		}
	} else {
		if hasIngress {
			np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
		}
		if hasEgress {
			np.Spec.PolicyTypes = append(np.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)
		}
		if !hasIngress && !hasEgress {
			np.Spec.PolicyTypes = []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			}
		}
	}

	if hasIngress {
		for _, rule := range req.Ingress {
			ingressRule := networkingv1.NetworkPolicyIngressRule{}
			for _, from := range rule.From {
				peer := networkingv1.NetworkPolicyPeer{}
				if len(from.NamespaceSelector) > 0 {
					peer.NamespaceSelector = &metav1.LabelSelector{MatchLabels: from.NamespaceSelector}
				}
				if len(from.PodSelector) > 0 {
					peer.PodSelector = &metav1.LabelSelector{MatchLabels: from.PodSelector}
				}
				if from.IPBlock != "" {
					peer.IPBlock = &networkingv1.IPBlock{CIDR: from.IPBlock}
				}
				ingressRule.From = append(ingressRule.From, peer)
			}
			for _, port := range rule.Ports {
				p := networkingv1.NetworkPolicyPort{}
				if port.Port != "" {
					var ios intstr.IntOrString
					if v, err := strconv.Atoi(port.Port); err == nil {
						ios = intstr.FromInt(v)
					} else {
						ios = intstr.FromString(port.Port)
					}
					p.Port = &ios
				}
				if port.Protocol != "" {
					proto := corev1.Protocol(port.Protocol)
					p.Protocol = &proto
				}
				ingressRule.Ports = append(ingressRule.Ports, p)
			}
			np.Spec.Ingress = append(np.Spec.Ingress, ingressRule)
		}
	}

	if hasEgress {
		for _, rule := range req.Egress {
			egressRule := networkingv1.NetworkPolicyEgressRule{}
			for _, to := range rule.To {
				peer := networkingv1.NetworkPolicyPeer{}
				if len(to.NamespaceSelector) > 0 {
					peer.NamespaceSelector = &metav1.LabelSelector{MatchLabels: to.NamespaceSelector}
				}
				if len(to.PodSelector) > 0 {
					peer.PodSelector = &metav1.LabelSelector{MatchLabels: to.PodSelector}
				}
				if to.IPBlock != "" {
					peer.IPBlock = &networkingv1.IPBlock{CIDR: to.IPBlock}
				}
				egressRule.To = append(egressRule.To, peer)
			}
			for _, port := range rule.Ports {
				p := networkingv1.NetworkPolicyPort{}
				if port.Port != "" {
					var ios intstr.IntOrString
					if v, err := strconv.Atoi(port.Port); err == nil {
						ios = intstr.FromInt(v)
					} else {
						ios = intstr.FromString(port.Port)
					}
					p.Port = &ios
				}
				if port.Protocol != "" {
					proto := corev1.Protocol(port.Protocol)
					p.Protocol = &proto
				}
				egressRule.Ports = append(egressRule.Ports, p)
			}
			np.Spec.Egress = append(np.Spec.Egress, egressRule)
		}
	}

	_, err := s.client.NetworkingV1().NetworkPolicies(req.Namespace).Create(ctx, np, metav1.CreateOptions{})
	return err
}

type EndpointInfo struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Addresses []string            `json:"addresses"`
	Ports     []EndpointPortInfo  `json:"ports"`
	Age       string              `json:"age"`
}

type EndpointPortInfo struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

func (s *NetworkService) ListEndpoints(ctx context.Context, namespace string) ([]EndpointInfo, error) {
	epList, err := s.client.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []EndpointInfo
	for _, ep := range epList.Items {
		info := EndpointInfo{
			Name:      ep.Name,
			Namespace: ep.Namespace,
			Age:       formatAge(ep.CreationTimestamp.Time),
		}

		for _, subset := range ep.Subsets {
			for _, addr := range subset.Addresses {
				info.Addresses = append(info.Addresses, addr.IP)
			}
			for _, port := range subset.Ports {
				info.Ports = append(info.Ports, EndpointPortInfo{
					Name:     port.Name,
					Port:     port.Port,
					Protocol: string(port.Protocol),
				})
			}
		}

		result = append(result, info)
	}
	return result, nil
}

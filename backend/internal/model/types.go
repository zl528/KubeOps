package model

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/apps/v1"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ClusterOverview struct {
	TotalNodes    int               `json:"totalNodes"`
	ReadyNodes    int               `json:"readyNodes"`
	Namespaces    int               `json:"namespaces"`
	TotalPods     int               `json:"totalPods"`
	RunningPods   int               `json:"runningPods"`
	PendingPods   int               `json:"pendingPods"`
	FailedPods    int               `json:"failedPods"`
	CPUUsage     ResourceUsage     `json:"cpuUsage"`
	MemoryUsage  ResourceUsage     `json:"memoryUsage"`
}

type ResourceUsage struct {
	Used  string `json:"used"`
	Total string `json:"total"`
	Pct   string `json:"percentage"`
}

type NodeInfo struct {
	Name            string            `json:"name"`
	Status          string            `json:"status"`
	Roles           []string          `json:"roles"`
	Version         string            `json:"version"`
	OS              string            `json:"os"`
	KernelVersion   string            `json:"kernelVersion"`
	CPU             string            `json:"cpu"`
	Memory          string            `json:"memory"`
	Labels          map[string]string `json:"labels,omitempty"`
	Taints          []TaintInfo       `json:"taints,omitempty"`
}

type TaintInfo struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

type NamespaceInfo struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels,omitempty"`
	Age    string            `json:"age"`
}

type PodInfo struct {
	Name      string         `json:"name"`
	Namespace string         `json:"namespace"`
	Status    string         `json:"status"`
	Node      string         `json:"node"`
	Restarts  int32          `json:"restarts"`
	Age       string         `json:"age"`
	Containers []ContainerInfo `json:"containers"`
}

type ContainerInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Ready bool   `json:"ready"`
}

type DeploymentInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  string `json:"replicas"`
	Ready     string `json:"ready"`
	UpToDate  int32  `json:"upToDate"`
	Available int32  `json:"available"`
	Age       string `json:"age"`
}

type ServiceInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"clusterIP"`
	Ports      []ServicePortInfo `json:"ports"`
	Age        string            `json:"age"`
}

type ServicePortInfo struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Protocol   string `json:"protocol"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

type IngressInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Hosts     string `json:"hosts"`
	TLS       bool   `json:"tls"`
	Age       string `json:"age"`
}

type EventInfo struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Count     int32  `json:"count"`
	LastTime  string `json:"lastTime"`
	Object    string `json:"object"`
}

type ResourceQuota struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Hard      map[string]string `json:"hard"`
	Used      map[string]string `json:"used"`
}

type PodLogRequest struct {
	Namespace  string `json:"namespace"`
	Pod        string `json:"pod"`
	Container  string `json:"container"`
	Lines      int64  `json:"lines"`
	TailLines  *int64 `json:"tailLines,omitempty"`
}

type ExecRequest struct {
	Namespace string   `json:"namespace"`
	Pod       string   `json:"pod"`
	Container string   `json:"container"`
	Command   []string `json:"command"`
}

type RollbackRequest struct {
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Revision   int    `json:"revision"`
}

type ScaleRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Replicas  int32  `json:"replicas"`
}

// Import usage hints
var (
	_ = corev1.Service{}
	_ = v1.Deployment{}
)

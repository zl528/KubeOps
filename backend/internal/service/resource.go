package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ResourceService struct {
	client kubernetes.Interface
}

func NewResourceService(client kubernetes.Interface) *ResourceService {
	return &ResourceService{client: client}
}

func (s *ResourceService) GetPodLogs(ctx context.Context, namespace, pod, container string, tailLines int64) (string, error) {
	opts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	}

	req := s.client.CoreV1().Pods(namespace).GetLogs(pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return "", fmt.Errorf("read pod logs: %w", err)
	}

	return buf.String(), nil
}

func (s *ResourceService) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	deploy, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment: %w", err)
	}

	deploy.Spec.Replicas = &replicas
	_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("scale deployment: %w", err)
	}

	return nil
}

func (s *ResourceService) RollbackDeployment(ctx context.Context, namespace, name string, revision int) error {
	deploy, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment: %w", err)
	}

	if deploy.Annotations == nil {
		deploy.Annotations = make(map[string]string)
	}
	deploy.Annotations["deployment.kubernetes.io/revision"] = fmt.Sprintf("%d", revision)

	_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("rollback deployment: %w", err)
	}

	return nil
}

func (s *ResourceService) DeleteDeployment(ctx context.Context, namespace, name string) error {
	return s.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *ResourceService) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (s *ResourceService) RestartDeployment(ctx context.Context, namespace, name string) error {
	deploy, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment: %w", err)
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("2006-01-02T15:04:05Z")

	_, err = s.client.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("restart deployment: %w", err)
	}

	return nil
}

type ConfigMapInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Data        map[string]string `json:"data,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

func (s *ResourceService) ListConfigMaps(ctx context.Context, namespace string) ([]ConfigMapInfo, error) {
	cmList, err := s.client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ConfigMapInfo
	for _, cm := range cmList.Items {
		result = append(result, ConfigMapInfo{
			Name:        cm.Name,
			Namespace:   cm.Namespace,
			Data:        cm.Data,
			Labels:      cm.Labels,
			Annotations: cm.Annotations,
			Age:         formatAge(cm.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *ResourceService) GetConfigMap(ctx context.Context, namespace, name string) (*ConfigMapInfo, error) {
	cm, err := s.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &ConfigMapInfo{
		Name:        cm.Name,
		Namespace:   cm.Namespace,
		Data:        cm.Data,
		Labels:      cm.Labels,
		Annotations: cm.Annotations,
		Age:         formatAge(cm.CreationTimestamp.Time),
	}, nil
}

func (s *ResourceService) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type ConfigMapCreateRequest struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Data        map[string]string `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func (s *ResourceService) CreateConfigMap(ctx context.Context, req ConfigMapCreateRequest) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Data: req.Data,
	}
	_, err := s.client.CoreV1().ConfigMaps(req.Namespace).Create(ctx, cm, metav1.CreateOptions{})
	return err
}

func (s *ResourceService) UpdateConfigMap(ctx context.Context, req ConfigMapCreateRequest) error {
	cm, err := s.client.CoreV1().ConfigMaps(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap: %w", err)
	}
	cm.Data = req.Data
	cm.Labels = req.Labels
	cm.Annotations = req.Annotations
	_, err = s.client.CoreV1().ConfigMaps(req.Namespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

type SecretInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	DataKeys    []string          `json:"dataKeys"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

func (s *ResourceService) ListSecrets(ctx context.Context, namespace string) ([]SecretInfo, error) {
	secretList, err := s.client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []SecretInfo
	for _, sec := range secretList.Items {
		keys := make([]string, 0, len(sec.Data))
		for k := range sec.Data {
			keys = append(keys, k)
		}
		result = append(result, SecretInfo{
			Name:        sec.Name,
			Namespace:   sec.Namespace,
			Type:        string(sec.Type),
			DataKeys:    keys,
			Labels:      sec.Labels,
			Annotations: sec.Annotations,
			Age:         formatAge(sec.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *ResourceService) DeleteSecret(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type SecretCreateRequest struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Data        map[string]string `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func (s *ResourceService) CreateSecret(ctx context.Context, req SecretCreateRequest) error {
	secretData := make(map[string][]byte)
	for k, v := range req.Data {
		secretData[k] = []byte(v)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Type: corev1.SecretType(req.Type),
		Data: secretData,
	}
	_, err := s.client.CoreV1().Secrets(req.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (s *ResourceService) UpdateSecret(ctx context.Context, req SecretCreateRequest) error {
	secret, err := s.client.CoreV1().Secrets(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get secret: %w", err)
	}
	secretData := make(map[string][]byte)
	for k, v := range req.Data {
		secretData[k] = []byte(v)
	}
	secret.Data = secretData
	if req.Type != "" {
		secret.Type = corev1.SecretType(req.Type)
	}
	secret.Labels = req.Labels
	secret.Annotations = req.Annotations
	_, err = s.client.CoreV1().Secrets(req.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

type ExecResult struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (s *ResourceService) ExecCommand(ctx context.Context, restConfig *rest.Config, namespace, pod, container string, command []string) (*ExecResult, error) {
	podObj, err := s.client.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod: %w", err)
	}

	if podObj.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("pod is not running, status: %s", podObj.Status.Phase)
	}

	result := &ExecResult{}
	result.Stdout = fmt.Sprintf("Exec command: %v on pod %s/%s", command, namespace, pod)
	return result, nil
}

func (s *ResourceService) DeletePod(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *ResourceService) RestartPod(ctx context.Context, namespace, name string) error {
	pod, err := s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get pod: %w", err)
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("2006-01-02T15:04:05Z")

	_, err = s.client.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("restart pod: %w", err)
	}

	return nil
}

func (s *ResourceService) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return s.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

type ResourceQuotaInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Hard      map[string]string `json:"hard"`
	Used      map[string]string `json:"used"`
}

func (s *ResourceService) ListResourceQuotas(ctx context.Context, namespace string) ([]ResourceQuotaInfo, error) {
	rqList, err := s.client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return []ResourceQuotaInfo{}, nil
		}
		return nil, err
	}

	var result []ResourceQuotaInfo
	for _, rq := range rqList.Items {
		hard := make(map[string]string)
		for k, v := range rq.Status.Hard {
			hard[k.String()] = v.String()
		}
		used := make(map[string]string)
		for k, v := range rq.Status.Used {
			used[k.String()] = v.String()
		}

		result = append(result, ResourceQuotaInfo{
			Name:      rq.Name,
			Namespace: rq.Namespace,
			Hard:      hard,
			Used:      used,
		})
	}
	return result, nil
}

func (s *ResourceService) GetResourceQuota(ctx context.Context, namespace, name string) (*ResourceQuotaInfo, error) {
	rq, err := s.client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	hard := make(map[string]string)
	for k, v := range rq.Status.Hard {
		hard[k.String()] = v.String()
	}
	used := make(map[string]string)
	for k, v := range rq.Status.Used {
		used[k.String()] = v.String()
	}

	return &ResourceQuotaInfo{
		Name:      rq.Name,
		Namespace: rq.Namespace,
		Hard:      hard,
		Used:      used,
	}, nil
}

type ResourceQuotaCreateRequest struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Hard      map[string]string `json:"hard"`
}

func (s *ResourceService) CreateResourceQuota(ctx context.Context, req ResourceQuotaCreateRequest) error {
	hard := make(corev1.ResourceList)
	for k, v := range req.Hard {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return fmt.Errorf("parse quantity %s: %w", v, err)
		}
		hard[corev1.ResourceName(k)] = q
	}

	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: hard,
		},
	}

	_, err := s.client.CoreV1().ResourceQuotas(req.Namespace).Create(ctx, rq, metav1.CreateOptions{})
	return err
}

func (s *ResourceService) UpdateResourceQuota(ctx context.Context, req ResourceQuotaCreateRequest) error {
	rq, err := s.client.CoreV1().ResourceQuotas(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	hard := make(corev1.ResourceList)
	for k, v := range req.Hard {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return fmt.Errorf("parse quantity %s: %w", v, err)
		}
		hard[corev1.ResourceName(k)] = q
	}

	rq.Spec.Hard = hard
	_, err = s.client.CoreV1().ResourceQuotas(req.Namespace).Update(ctx, rq, metav1.UpdateOptions{})
	return err
}

func (s *ResourceService) DeleteResourceQuota(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type UpdateDeploymentRequest struct {
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	Replicas        *int32            `json:"replicas,omitempty"`
	Image           string            `json:"image,omitempty"`
	ContainerName   string            `json:"containerName,omitempty"`
	EnvVars         map[string]string `json:"envVars,omitempty"`
	ResourceLimits  map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
}

func (s *ResourceService) UpdateDeployment(ctx context.Context, req UpdateDeploymentRequest) error {
	deploy, err := s.client.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment: %w", err)
	}

	if req.Replicas != nil {
		deploy.Spec.Replicas = req.Replicas
	}

	if len(deploy.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	containerIdx := 0
	if req.ContainerName != "" {
		found := false
		for i, c := range deploy.Spec.Template.Spec.Containers {
			if c.Name == req.ContainerName {
				containerIdx = i
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("container %s not found", req.ContainerName)
		}
	}

	if req.Image != "" {
		deploy.Spec.Template.Spec.Containers[containerIdx].Image = req.Image
	}

	if req.EnvVars != nil {
		envMap := make(map[string]corev1.EnvVar)
		for _, e := range deploy.Spec.Template.Spec.Containers[containerIdx].Env {
			envMap[e.Name] = e
		}
		for k, v := range req.EnvVars {
			envMap[k] = corev1.EnvVar{Name: k, Value: v}
		}
		envs := make([]corev1.EnvVar, 0, len(envMap))
		for _, e := range envMap {
			envs = append(envs, e)
		}
		deploy.Spec.Template.Spec.Containers[containerIdx].Env = envs
	}

	if req.ResourceLimits != nil || req.ResourceRequests != nil {
		deploy.Spec.Template.Spec.Containers[containerIdx].Resources = corev1.ResourceRequirements{
			Limits:   parseResourceList(req.ResourceLimits),
			Requests: parseResourceList(req.ResourceRequests),
		}
	}

	_, err = s.client.AppsV1().Deployments(req.Namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update deployment: %w", err)
	}

	return nil
}

func parseResourceList(m map[string]string) corev1.ResourceList {
	if len(m) == 0 {
		return nil
	}
	rl := make(corev1.ResourceList)
	for k, v := range m {
		if q, err := resource.ParseQuantity(v); err == nil {
			rl[corev1.ResourceName(k)] = q
		}
	}
	return rl
}

func (s *ResourceService) CreateNamespace(ctx context.Context, name string, labels map[string]string, annotations map[string]string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
	_, err := s.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

func (s *ResourceService) CreateService(ctx context.Context, namespace, name, svcType string, ports []ServicePortConfig, selector map[string]string, labels map[string]string, annotations map[string]string) error {
	var k8sPorts []corev1.ServicePort
	for _, p := range ports {
		k8sPorts = append(k8sPorts, corev1.ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: intstr.FromInt(int(p.TargetPort)),
			Protocol:   corev1.Protocol(p.Protocol),
			NodePort:   p.NodePort,
		})
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(svcType),
			Ports:    k8sPorts,
			Selector: selector,
		},
	}
	_, err := s.client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}

func (s *ResourceService) UpdateService(ctx context.Context, namespace, name, svcType string, ports []ServicePortConfig, selector map[string]string, labels map[string]string, annotations map[string]string) error {
	existing, err := s.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get service: %w", err)
	}

	if svcType != "" {
		existing.Spec.Type = corev1.ServiceType(svcType)
	}

	if len(ports) > 0 {
		var k8sPorts []corev1.ServicePort
		for _, p := range ports {
			k8sPorts = append(k8sPorts, corev1.ServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: intstr.FromInt(int(p.TargetPort)),
				Protocol:   corev1.Protocol(p.Protocol),
				NodePort:   p.NodePort,
			})
		}
		existing.Spec.Ports = k8sPorts
	}

	if selector != nil {
		existing.Spec.Selector = selector
	}

	if labels != nil {
		existing.Labels = labels
	}

	if annotations != nil {
		existing.Annotations = annotations
	}

	_, err = s.client.CoreV1().Services(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

type ServicePortConfig struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Protocol   string `json:"protocol"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

func (s *ResourceService) DeleteNamespace(ctx context.Context, name string) error {
	return s.client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

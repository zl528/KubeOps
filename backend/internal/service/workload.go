package service

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type WorkloadService struct {
	client kubernetes.Interface
}

func NewWorkloadService(client kubernetes.Interface) *WorkloadService {
	return &WorkloadService{client: client}
}

type StatefulSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  string `json:"replicas"`
	Ready     string `json:"ready"`
	Age       string `json:"age"`
}

func (s *WorkloadService) ListStatefulSets(ctx context.Context, namespace string) ([]StatefulSetInfo, error) {
	stsList, err := s.client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []StatefulSetInfo
	for _, sts := range stsList.Items {
		result = append(result, StatefulSetInfo{
			Name:      sts.Name,
			Namespace: sts.Namespace,
			Replicas:  fmt.Sprintf("%d/%d", *sts.Spec.Replicas, sts.Status.Replicas),
			Ready:     fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, sts.Status.Replicas),
			Age:       formatAge(sts.CreationTimestamp.Time),
		})
	}
	return result, nil
}

type StatefulSetDetail struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Replicas         int32             `json:"replicas"`
	ReadyReplicas    int32             `json:"readyReplicas"`
	AvailableReplicas int32            `json:"availableReplicas"`
	Image            string            `json:"image"`
	ServiceName      string            `json:"serviceName"`
	Labels           map[string]string `json:"labels"`
	Selector         map[string]string `json:"selector"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
	Age              string            `json:"age"`
}

func (s *WorkloadService) GetStatefulSet(ctx context.Context, namespace, name string) (*StatefulSetDetail, error) {
	sts, err := s.client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var image string
	var resourceLimits, resourceRequests map[string]string
	if len(sts.Spec.Template.Spec.Containers) > 0 {
		c := sts.Spec.Template.Spec.Containers[0]
		image = c.Image
		if c.Resources.Limits != nil {
			resourceLimits = make(map[string]string)
			for k, v := range c.Resources.Limits {
				resourceLimits[string(k)] = v.String()
			}
		}
		if c.Resources.Requests != nil {
			resourceRequests = make(map[string]string)
			for k, v := range c.Resources.Requests {
				resourceRequests[string(k)] = v.String()
			}
		}
	}

	return &StatefulSetDetail{
		Name:              sts.Name,
		Namespace:         sts.Namespace,
		Replicas:          *sts.Spec.Replicas,
		ReadyReplicas:     sts.Status.ReadyReplicas,
		AvailableReplicas: sts.Status.AvailableReplicas,
		Image:             image,
		ServiceName:       sts.Spec.ServiceName,
		Labels:            sts.Labels,
		Selector:          sts.Spec.Selector.MatchLabels,
		ResourceLimits:    resourceLimits,
		ResourceRequests:  resourceRequests,
		Age:               formatAge(sts.CreationTimestamp.Time),
	}, nil
}

func (s *WorkloadService) ScaleStatefulSet(ctx context.Context, namespace, name string, replicas int32) error {
	sts, err := s.client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get statefulset: %w", err)
	}

	sts.Spec.Replicas = &replicas
	_, err = s.client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("scale statefulset: %w", err)
	}

	return nil
}

func (s *WorkloadService) DeleteStatefulSet(ctx context.Context, namespace, name string) error {
	return s.client.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *WorkloadService) RestartStatefulSet(ctx context.Context, namespace, name string) error {
	sts, err := s.client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get statefulset: %w", err)
	}

	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = make(map[string]string)
	}
	sts.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("2006-01-02T15:04:05Z")

	_, err = s.client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("restart statefulset: %w", err)
	}

	return nil
}

type UpdateStatefulSetRequest struct {
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	Replicas        *int32            `json:"replicas,omitempty"`
	Image           string            `json:"image,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	ResourceLimits  map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
}

func (s *WorkloadService) UpdateStatefulSet(ctx context.Context, req UpdateStatefulSetRequest) error {
	sts, err := s.client.AppsV1().StatefulSets(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get statefulset: %w", err)
	}

	if req.Replicas != nil {
		sts.Spec.Replicas = req.Replicas
	}

	if req.Image != "" && len(sts.Spec.Template.Spec.Containers) > 0 {
		sts.Spec.Template.Spec.Containers[0].Image = req.Image
	}

	if req.Labels != nil {
		for k, v := range req.Labels {
			sts.Labels[k] = v
			sts.Spec.Template.Labels[k] = v
		}
	}

	if req.ResourceLimits != nil || req.ResourceRequests != nil {
		if len(sts.Spec.Template.Spec.Containers) > 0 {
			container := &sts.Spec.Template.Spec.Containers[0]
			if container.Resources.Limits == nil {
				container.Resources.Limits = make(corev1.ResourceList)
			}
			if container.Resources.Requests == nil {
				container.Resources.Requests = make(corev1.ResourceList)
			}
			for k, v := range req.ResourceLimits {
				container.Resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
			}
			for k, v := range req.ResourceRequests {
				container.Resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
			}
		}
	}

	_, err = s.client.AppsV1().StatefulSets(req.Namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update statefulset: %w", err)
	}

	return nil
}

type DaemonSetInfo struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Desired         int32  `json:"desired"`
	Current         int32  `json:"current"`
	Ready           int32  `json:"ready"`
	UpToDate        int32  `json:"upToDate"`
	Available       int32  `json:"available"`
	Age             string `json:"age"`
}

func (s *WorkloadService) ListDaemonSets(ctx context.Context, namespace string) ([]DaemonSetInfo, error) {
	dsList, err := s.client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []DaemonSetInfo
	for _, ds := range dsList.Items {
		result = append(result, DaemonSetInfo{
			Name:      ds.Name,
			Namespace: ds.Namespace,
			Desired:   ds.Status.DesiredNumberScheduled,
			Current:   ds.Status.CurrentNumberScheduled,
			Ready:     ds.Status.NumberReady,
			UpToDate:  ds.Status.UpdatedNumberScheduled,
			Available: ds.Status.NumberAvailable,
			Age:       formatAge(ds.CreationTimestamp.Time),
		})
	}
	return result, nil
}

type DaemonSetDetail struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Desired          int32             `json:"desired"`
	Current          int32             `json:"current"`
	Ready            int32             `json:"ready"`
	UpToDate         int32             `json:"upToDate"`
	Available        int32             `json:"available"`
	Image            string            `json:"image"`
	Labels           map[string]string `json:"labels"`
	Selector         map[string]string `json:"selector"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
	Age              string            `json:"age"`
}

func (s *WorkloadService) GetDaemonSet(ctx context.Context, namespace, name string) (*DaemonSetDetail, error) {
	ds, err := s.client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var image string
	var resourceLimits, resourceRequests map[string]string
	if len(ds.Spec.Template.Spec.Containers) > 0 {
		c := ds.Spec.Template.Spec.Containers[0]
		image = c.Image
		if c.Resources.Limits != nil {
			resourceLimits = make(map[string]string)
			for k, v := range c.Resources.Limits {
				resourceLimits[string(k)] = v.String()
			}
		}
		if c.Resources.Requests != nil {
			resourceRequests = make(map[string]string)
			for k, v := range c.Resources.Requests {
				resourceRequests[string(k)] = v.String()
			}
		}
	}

	return &DaemonSetDetail{
		Name:              ds.Name,
		Namespace:         ds.Namespace,
		Desired:           ds.Status.DesiredNumberScheduled,
		Current:           ds.Status.CurrentNumberScheduled,
		Ready:             ds.Status.NumberReady,
		UpToDate:          ds.Status.UpdatedNumberScheduled,
		Available:         ds.Status.NumberAvailable,
		Image:             image,
		Labels:            ds.Labels,
		Selector:          ds.Spec.Selector.MatchLabels,
		ResourceLimits:    resourceLimits,
		ResourceRequests:  resourceRequests,
		Age:               formatAge(ds.CreationTimestamp.Time),
	}, nil
}

func (s *WorkloadService) DeleteDaemonSet(ctx context.Context, namespace, name string) error {
	return s.client.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *WorkloadService) RestartDaemonSet(ctx context.Context, namespace, name string) error {
	ds, err := s.client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get daemonset: %w", err)
	}

	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = make(map[string]string)
	}
	ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("2006-01-02T15:04:05Z")

	_, err = s.client.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("restart daemonset: %w", err)
	}

	return nil
}

type UpdateDaemonSetRequest struct {
	Namespace        string            `json:"namespace"`
	Name             string            `json:"name"`
	Image            string            `json:"image,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
}

func (s *WorkloadService) UpdateDaemonSet(ctx context.Context, req UpdateDaemonSetRequest) error {
	ds, err := s.client.AppsV1().DaemonSets(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get daemonset: %w", err)
	}

	if req.Image != "" && len(ds.Spec.Template.Spec.Containers) > 0 {
		ds.Spec.Template.Spec.Containers[0].Image = req.Image
	}

	if req.Labels != nil {
		for k, v := range req.Labels {
			ds.Labels[k] = v
			ds.Spec.Template.Labels[k] = v
		}
	}

	if req.ResourceLimits != nil || req.ResourceRequests != nil {
		if len(ds.Spec.Template.Spec.Containers) > 0 {
			container := &ds.Spec.Template.Spec.Containers[0]
			if container.Resources.Limits == nil {
				container.Resources.Limits = make(corev1.ResourceList)
			}
			if container.Resources.Requests == nil {
				container.Resources.Requests = make(corev1.ResourceList)
			}
			for k, v := range req.ResourceLimits {
				container.Resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
			}
			for k, v := range req.ResourceRequests {
				container.Resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
			}
		}
	}

	_, err = s.client.AppsV1().DaemonSets(req.Namespace).Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update daemonset: %w", err)
	}

	return nil
}

type CronJobInfo struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Schedule          string `json:"schedule"`
	Suspend           bool   `json:"suspend"`
	Active            int32  `json:"active"`
	LastScheduleTime  string `json:"lastScheduleTime"`
	Age               string `json:"age"`
}

func (s *WorkloadService) ListCronJobs(ctx context.Context, namespace string) ([]CronJobInfo, error) {
	cjList, err := s.client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []CronJobInfo
	for _, cj := range cjList.Items {
		var lastSchedule string
		if cj.Status.LastScheduleTime != nil {
			lastSchedule = cj.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
		}

		var active int32
		if cj.Status.Active != nil {
			active = int32(len(cj.Status.Active))
		}

		result = append(result, CronJobInfo{
			Name:             cj.Name,
			Namespace:        cj.Namespace,
			Schedule:         cj.Spec.Schedule,
			Suspend:          *cj.Spec.Suspend,
			Active:           active,
			LastScheduleTime: lastSchedule,
			Age:              formatAge(cj.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *WorkloadService) DeleteCronJob(ctx context.Context, namespace, name string) error {
	return s.client.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *WorkloadService) SuspendCronJob(ctx context.Context, namespace, name string, suspend bool) error {
	cj, err := s.client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get cronjob: %w", err)
	}

	cj.Spec.Suspend = &suspend
	_, err = s.client.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update cronjob: %w", err)
	}

	return nil
}

func (s *WorkloadService) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	return s.client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

type UpdateCronJobRequest struct {
	Namespace         string            `json:"namespace"`
	Name              string            `json:"name"`
	Schedule          string            `json:"schedule,omitempty"`
	Suspend           *bool             `json:"suspend,omitempty"`
	ConcurrencyPolicy string            `json:"concurrencyPolicy,omitempty"`
	JobTemplateSpec   *JobTemplateSpec   `json:"jobTemplateSpec,omitempty"`
}

type JobTemplateSpec struct {
	Completions  *int32           `json:"completions,omitempty"`
	Parallelism  *int32           `json:"parallelism,omitempty"`
	BackoffLimit *int32           `json:"backoffLimit,omitempty"`
	Template     *PodTemplateSpec `json:"template,omitempty"`
}

type PodTemplateSpec struct {
	Containers []ContainerUpdate `json:"containers,omitempty"`
}

type ContainerUpdate struct {
	Name  string            `json:"name"`
	Image string            `json:"image,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
}

func (s *WorkloadService) UpdateCronJob(ctx context.Context, req UpdateCronJobRequest) error {
	cj, err := s.client.BatchV1().CronJobs(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get cronjob: %w", err)
	}

	if req.Schedule != "" {
		cj.Spec.Schedule = req.Schedule
	}
	if req.Suspend != nil {
		cj.Spec.Suspend = req.Suspend
	}
	if req.ConcurrencyPolicy != "" {
		cj.Spec.ConcurrencyPolicy = batchv1.ConcurrencyPolicy(req.ConcurrencyPolicy)
	}
	if req.JobTemplateSpec != nil {
		if req.JobTemplateSpec.Completions != nil {
			cj.Spec.JobTemplate.Spec.Completions = req.JobTemplateSpec.Completions
		}
		if req.JobTemplateSpec.Parallelism != nil {
			cj.Spec.JobTemplate.Spec.Parallelism = req.JobTemplateSpec.Parallelism
		}
		if req.JobTemplateSpec.BackoffLimit != nil {
			cj.Spec.JobTemplate.Spec.BackoffLimit = req.JobTemplateSpec.BackoffLimit
		}
		if req.JobTemplateSpec.Template != nil && len(req.JobTemplateSpec.Template.Containers) > 0 {
			for _, update := range req.JobTemplateSpec.Template.Containers {
				for i, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
					if c.Name == update.Name {
						if update.Image != "" {
							cj.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Image = update.Image
						}
						if update.Env != nil {
							var envVars []corev1.EnvVar
							for k, v := range update.Env {
								envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
							}
							cj.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Env = envVars
						}
						break
					}
				}
			}
		}
	}

	_, err = s.client.BatchV1().CronJobs(req.Namespace).Update(ctx, cj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update cronjob: %w", err)
	}

	return nil
}

type JobInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Completions  string `json:"completions"`
	Duration     string `json:"duration"`
	Status       string `json:"status"`
	Age          string `json:"age"`
}

func (s *WorkloadService) ListJobs(ctx context.Context, namespace string) ([]JobInfo, error) {
	jobList, err := s.client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []JobInfo
	for _, job := range jobList.Items {
		status := "Running"
		if job.Status.Succeeded > 0 {
			status = "Complete"
		}
		if job.Status.Failed > 0 {
			status = "Failed"
		}

		result = append(result, JobInfo{
			Name:        job.Name,
			Namespace:   job.Namespace,
			Completions: fmt.Sprintf("%d/%d", job.Status.Succeeded, *job.Spec.Completions),
			Status:      status,
			Age:         formatAge(job.CreationTimestamp.Time),
		})
	}
	return result, nil
}

type JobDetail struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Completions      int32             `json:"completions"`
	Parallelism      int32             `json:"parallelism"`
	BackoffLimit     int32             `json:"backoffLimit"`
	Succeeded        int32             `json:"succeeded"`
	Failed           int32             `json:"failed"`
	Active           int32             `json:"active"`
	Image            string            `json:"image"`
	Labels           map[string]string `json:"labels"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
	Age              string            `json:"age"`
}

func (s *WorkloadService) GetJob(ctx context.Context, namespace, name string) (*JobDetail, error) {
	job, err := s.client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var image string
	var resourceLimits, resourceRequests map[string]string
	if len(job.Spec.Template.Spec.Containers) > 0 {
		c := job.Spec.Template.Spec.Containers[0]
		image = c.Image
		if c.Resources.Limits != nil {
			resourceLimits = make(map[string]string)
			for k, v := range c.Resources.Limits {
				resourceLimits[string(k)] = v.String()
			}
		}
		if c.Resources.Requests != nil {
			resourceRequests = make(map[string]string)
			for k, v := range c.Resources.Requests {
				resourceRequests[string(k)] = v.String()
			}
		}
	}

	completions := int32(1)
	if job.Spec.Completions != nil {
		completions = *job.Spec.Completions
	}
	parallelism := int32(1)
	if job.Spec.Parallelism != nil {
		parallelism = *job.Spec.Parallelism
	}
	backoffLimit := int32(6)
	if job.Spec.BackoffLimit != nil {
		backoffLimit = *job.Spec.BackoffLimit
	}

	return &JobDetail{
		Name:              job.Name,
		Namespace:         job.Namespace,
		Completions:       completions,
		Parallelism:       parallelism,
		BackoffLimit:      backoffLimit,
		Succeeded:         job.Status.Succeeded,
		Failed:            job.Status.Failed,
		Active:            job.Status.Active,
		Image:             image,
		Labels:            job.Labels,
		ResourceLimits:    resourceLimits,
		ResourceRequests:  resourceRequests,
		Age:               formatAge(job.CreationTimestamp.Time),
	}, nil
}

func (s *WorkloadService) DeleteJob(ctx context.Context, namespace, name string) error {
	return s.client.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type UpdateJobRequest struct {
	Namespace        string            `json:"namespace"`
	Name             string            `json:"name"`
	Image            string            `json:"image,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
}

func (s *WorkloadService) UpdateJob(ctx context.Context, req UpdateJobRequest) error {
	job, err := s.client.BatchV1().Jobs(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}

	if req.Image != "" && len(job.Spec.Template.Spec.Containers) > 0 {
		job.Spec.Template.Spec.Containers[0].Image = req.Image
	}

	if req.Labels != nil {
		for k, v := range req.Labels {
			job.Labels[k] = v
			job.Spec.Template.Labels[k] = v
		}
	}

	if req.ResourceLimits != nil || req.ResourceRequests != nil {
		if len(job.Spec.Template.Spec.Containers) > 0 {
			container := &job.Spec.Template.Spec.Containers[0]
			if container.Resources.Limits == nil {
				container.Resources.Limits = make(corev1.ResourceList)
			}
			if container.Resources.Requests == nil {
				container.Resources.Requests = make(corev1.ResourceList)
			}
			for k, v := range req.ResourceLimits {
				container.Resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
			}
			for k, v := range req.ResourceRequests {
				container.Resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
			}
		}
	}

	_, err = s.client.BatchV1().Jobs(req.Namespace).Update(ctx, job, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}

	return nil
}

func (s *WorkloadService) CreateDeployment(ctx context.Context, req CreateDeploymentRequest) error {
	if req.Name == "" || req.Namespace == "" || req.Image == "" {
		return fmt.Errorf("name, namespace and image are required")
	}

	replicas := int32(1)
	if req.Replicas > 0 {
		replicas = req.Replicas
	}

	labels := map[string]string{"app": req.Name}
	if req.Labels != nil {
		for k, v := range req.Labels {
			labels[k] = v
		}
	}

	var ports []corev1.ContainerPort
	for _, p := range req.Ports {
		ports = append(ports, corev1.ContainerPort{
			ContainerPort: p,
		})
	}

	var envVars []corev1.EnvVar
	for k, v := range req.EnvVars {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	var resources corev1.ResourceRequirements
	if req.ResourceLimits != nil {
		resources.Limits = make(corev1.ResourceList)
		for k, v := range req.ResourceLimits {
			resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if req.ResourceRequests != nil {
		resources.Requests = make(corev1.ResourceList)
		for k, v := range req.ResourceRequests {
			resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}

	restartPolicy := corev1.RestartPolicyAlways
	switch req.RestartPolicy {
	case "OnFailure":
		restartPolicy = corev1.RestartPolicyOnFailure
	case "Never":
		restartPolicy = corev1.RestartPolicyNever
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:      req.Name,
				Image:     req.Image,
				Ports:     ports,
				Env:       envVars,
				Resources: resources,
			},
		},
		RestartPolicy: restartPolicy,
	}
	if req.NodeSelector != nil {
		podSpec.NodeSelector = req.NodeSelector
	}
	if req.ServiceAccountName != "" {
		podSpec.ServiceAccountName = req.ServiceAccountName
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: podSpec,
			},
		},
	}

	_, err := s.client.AppsV1().Deployments(req.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create deployment: %w", err)
	}
	return nil
}

func (s *WorkloadService) CreateStatefulSet(ctx context.Context, req CreateStatefulSetRequest) error {
	if req.Name == "" || req.Namespace == "" || req.Image == "" {
		return fmt.Errorf("name, namespace and image are required")
	}

	replicas := int32(1)
	if req.Replicas > 0 {
		replicas = req.Replicas
	}

	labels := map[string]string{"app": req.Name}
	if req.Labels != nil {
		for k, v := range req.Labels {
			labels[k] = v
		}
	}

	var containerPorts []corev1.ContainerPort
	if req.ContainerPort > 0 {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: req.ContainerPort,
		})
	}

	var envVars []corev1.EnvVar
	for k, v := range req.EnvVars {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	var resources corev1.ResourceRequirements
	if req.ResourceLimits != nil {
		resources.Limits = make(corev1.ResourceList)
		for k, v := range req.ResourceLimits {
			resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if req.ResourceRequests != nil {
		resources.Requests = make(corev1.ResourceList)
		for k, v := range req.ResourceRequests {
			resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}

	container := corev1.Container{
		Name:      req.Name,
		Image:     req.Image,
		Ports:     containerPorts,
		Env:       envVars,
		Resources: resources,
	}

	var volumeClaimTemplates []corev1.PersistentVolumeClaim
	for _, vc := range req.VolumeClaims {
		volumeClaimTemplates = append(volumeClaimTemplates, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:   vc.Name,
				Labels: labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(vc.Size),
					},
				},
				StorageClassName: &vc.StorageClass,
			},
		})
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: req.ServiceName,
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
				},
			},
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}

	_, err := s.client.AppsV1().StatefulSets(req.Namespace).Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create statefulset: %w", err)
	}
	return nil
}

func (s *WorkloadService) CreateDaemonSet(ctx context.Context, req CreateDaemonSetRequest) error {
	if req.Name == "" || req.Namespace == "" || req.Image == "" {
		return fmt.Errorf("name, namespace and image are required")
	}

	labels := map[string]string{"app": req.Name}
	if req.Labels != nil {
		for k, v := range req.Labels {
			labels[k] = v
		}
	}

	var containerPorts []corev1.ContainerPort
	if req.ContainerPort > 0 {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: req.ContainerPort,
		})
	}

	var envVars []corev1.EnvVar
	for k, v := range req.EnvVars {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	var resources corev1.ResourceRequirements
	if req.ResourceLimits != nil {
		resources.Limits = make(corev1.ResourceList)
		for k, v := range req.ResourceLimits {
			resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if req.ResourceRequests != nil {
		resources.Requests = make(corev1.ResourceList)
		for k, v := range req.ResourceRequests {
			resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:      req.Name,
				Image:     req.Image,
				Ports:     containerPorts,
				Env:       envVars,
				Resources: resources,
			},
		},
	}
	if len(req.NodeSelector) > 0 {
		podSpec.NodeSelector = req.NodeSelector
	}
	if len(req.Tolerations) > 0 {
		for _, t := range req.Tolerations {
			toleration := corev1.Toleration{
				Key:      t.Key,
				Operator: corev1.TolerationOperator(t.Operator),
				Value:    t.Value,
				Effect:   corev1.TaintEffect(t.Effect),
			}
			podSpec.Tolerations = append(podSpec.Tolerations, toleration)
		}
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: podSpec,
			},
		},
	}

	_, err := s.client.AppsV1().DaemonSets(req.Namespace).Create(ctx, ds, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create daemonset: %w", err)
	}
	return nil
}

func (s *WorkloadService) CreateCronJob(ctx context.Context, req CreateCronJobRequest) error {
	if req.Name == "" || req.Namespace == "" || req.Image == "" || req.Schedule == "" {
		return fmt.Errorf("name, namespace, image and schedule are required")
	}

	suspend := false
	if req.Suspend != nil {
		suspend = *req.Suspend
	}

	var backoffLimit int32 = 6
	var completions int32 = 1
	var parallelism int32 = 1

	var successfulJobsHistoryLimit *int32
	if req.SuccessfulJobsHistoryLimit != nil {
		successfulJobsHistoryLimit = req.SuccessfulJobsHistoryLimit
	}

	var failedJobsHistoryLimit *int32
	if req.FailedJobsHistoryLimit != nil {
		failedJobsHistoryLimit = req.FailedJobsHistoryLimit
	}

	cjCommand := []string{"/bin/sh", "-c"}
	if req.Command != "" {
		cjCommand = append(cjCommand, req.Command)
	} else {
		cjCommand = append(cjCommand, "echo hello")
	}

	concurrencyPolicy := batchv1.AllowConcurrent
	if req.ConcurrencyPolicy != "" {
		concurrencyPolicy = batchv1.ConcurrencyPolicy(req.ConcurrencyPolicy)
	}

	labels := map[string]string{"app": req.Name}
	for k, v := range req.Labels {
		labels[k] = v
	}

	var envVars []corev1.EnvVar
	for k, v := range req.EnvVars {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	var resources corev1.ResourceRequirements
	if req.ResourceLimits != nil {
		resources.Limits = make(corev1.ResourceList)
		for k, v := range req.ResourceLimits {
			resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if req.ResourceRequests != nil {
		resources.Requests = make(corev1.ResourceList)
		for k, v := range req.ResourceRequests {
			resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}

	restartPolicy := corev1.RestartPolicyOnFailure
	if req.RestartPolicy != "" {
		restartPolicy = corev1.RestartPolicy(req.RestartPolicy)
	}

	jobSpec := batchv1.JobSpec{
		Completions:  &completions,
		Parallelism:  &parallelism,
		BackoffLimit: &backoffLimit,
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      req.Name,
						Image:     req.Image,
						Command:   cjCommand,
						Env:       envVars,
						Resources: resources,
					},
				},
				RestartPolicy: restartPolicy,
			},
		},
	}
	if req.ActiveDeadlineSeconds != nil {
		jobSpec.ActiveDeadlineSeconds = req.ActiveDeadlineSeconds
	}
	if req.TTLSecondsAfterFinished != nil {
		jobSpec.TTLSecondsAfterFinished = req.TTLSecondsAfterFinished
	}

	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   req.Schedule,
			Suspend:                    &suspend,
			ConcurrencyPolicy:          concurrencyPolicy,
			SuccessfulJobsHistoryLimit: successfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     failedJobsHistoryLimit,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: jobSpec,
			},
		},
	}

	_, err := s.client.BatchV1().CronJobs(req.Namespace).Create(ctx, cj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create cronjob: %w", err)
	}
	return nil
}

func (s *WorkloadService) CreateJob(ctx context.Context, req CreateJobRequest) error {
	if req.Name == "" || req.Namespace == "" || req.Image == "" {
		return fmt.Errorf("name, namespace and image are required")
	}

	var backoffLimit int32 = 6
	if req.BackoffLimit != nil {
		backoffLimit = *req.BackoffLimit
	}

	var completions int32 = 1
	if req.Completions != nil {
		completions = *req.Completions
	}

	var parallelism int32 = 1
	if req.Parallelism != nil {
		parallelism = *req.Parallelism
	}

	jobCommand := []string{"/bin/sh", "-c", "echo hello"}
	if req.Command != "" {
		jobCommand = []string{"/bin/sh", "-c", req.Command}
	}

	labels := map[string]string{"app": req.Name}
	for k, v := range req.Labels {
		labels[k] = v
	}

	var envVars []corev1.EnvVar
	for k, v := range req.EnvVars {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	var resources corev1.ResourceRequirements
	if req.ResourceLimits != nil {
		resources.Limits = make(corev1.ResourceList)
		for k, v := range req.ResourceLimits {
			resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if req.ResourceRequests != nil {
		resources.Requests = make(corev1.ResourceList)
		for k, v := range req.ResourceRequests {
			resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}

	restartPolicy := corev1.RestartPolicyNever
	if req.RestartPolicy != "" {
		restartPolicy = corev1.RestartPolicy(req.RestartPolicy)
	}

	jobSpec := batchv1.JobSpec{
		Completions:  &completions,
		Parallelism:  &parallelism,
		BackoffLimit: &backoffLimit,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      req.Name,
						Image:     req.Image,
						Command:   jobCommand,
						Env:       envVars,
						Resources: resources,
					},
				},
				RestartPolicy: restartPolicy,
			},
		},
	}
	if req.ActiveDeadlineSeconds != nil {
		jobSpec.ActiveDeadlineSeconds = req.ActiveDeadlineSeconds
	}
	if req.TTLSecondsAfterFinished != nil {
		jobSpec.TTLSecondsAfterFinished = req.TTLSecondsAfterFinished
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Spec: jobSpec,
	}

	_, err := s.client.BatchV1().Jobs(req.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

type CreateDeploymentRequest struct {
	Name               string            `json:"name"`
	Namespace          string            `json:"namespace"`
	Replicas           int32             `json:"replicas"`
	Image              string            `json:"image"`
	Ports              []int32           `json:"ports"`
	Labels             map[string]string `json:"labels,omitempty"`
	EnvVars            map[string]string `json:"envVars,omitempty"`
	ResourceLimits     map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests   map[string]string `json:"resourceRequests,omitempty"`
	NodeSelector       map[string]string `json:"nodeSelector,omitempty"`
	ServiceAccountName string            `json:"serviceAccountName,omitempty"`
	RestartPolicy      string            `json:"restartPolicy,omitempty"`
}

type CreateStatefulSetRequest struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Replicas         int32             `json:"replicas"`
	Image            string            `json:"image"`
	ServiceName      string            `json:"serviceName"`
	ContainerPort    int32             `json:"containerPort,omitempty"`
	VolumeClaims     []VolumeClaim     `json:"volumeClaims,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	EnvVars          map[string]string `json:"envVars,omitempty"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
}

type VolumeClaim struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	StorageClass string `json:"storageClass,omitempty"`
}

type CreateDaemonSetRequest struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Image            string            `json:"image"`
	ContainerPort    int32             `json:"containerPort,omitempty"`
	NodeSelector     map[string]string `json:"nodeSelector,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	EnvVars          map[string]string `json:"envVars,omitempty"`
	ResourceLimits   map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests map[string]string `json:"resourceRequests,omitempty"`
	Tolerations      []Toleration      `json:"tolerations,omitempty"`
}

type Toleration struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect"`
}

type CreateCronJobRequest struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Schedule                   string            `json:"schedule"`
	Image                      string            `json:"image"`
	Command                    string            `json:"command,omitempty"`
	ConcurrencyPolicy          string            `json:"concurrencyPolicy,omitempty"`
	Suspend                    *bool             `json:"suspend,omitempty"`
	SuccessfulJobsHistoryLimit *int32            `json:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int32            `json:"failedJobsHistoryLimit,omitempty"`
	Labels                     map[string]string `json:"labels,omitempty"`
	EnvVars                    map[string]string `json:"envVars,omitempty"`
	ResourceLimits             map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests           map[string]string `json:"resourceRequests,omitempty"`
	RestartPolicy              string            `json:"restartPolicy,omitempty"`
	ActiveDeadlineSeconds      *int64            `json:"activeDeadlineSeconds,omitempty"`
	TTLSecondsAfterFinished    *int32            `json:"ttlSecondsAfterFinished,omitempty"`
}

type CreateJobRequest struct {
	Name                    string            `json:"name"`
	Namespace               string            `json:"namespace"`
	Image                   string            `json:"image"`
	Command                 string            `json:"command,omitempty"`
	Completions             *int32            `json:"completions,omitempty"`
	Parallelism             *int32            `json:"parallelism,omitempty"`
	BackoffLimit            *int32            `json:"backoffLimit,omitempty"`
	Labels                  map[string]string `json:"labels,omitempty"`
	EnvVars                 map[string]string `json:"envVars,omitempty"`
	ResourceLimits          map[string]string `json:"resourceLimits,omitempty"`
	ResourceRequests        map[string]string `json:"resourceRequests,omitempty"`
	RestartPolicy           string            `json:"restartPolicy,omitempty"`
	ActiveDeadlineSeconds   *int64            `json:"activeDeadlineSeconds,omitempty"`
	TTLSecondsAfterFinished *int32            `json:"ttlSecondsAfterFinished,omitempty"`
}

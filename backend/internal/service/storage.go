package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type StorageService struct {
	client kubernetes.Interface
}

func NewStorageService(client kubernetes.Interface) *StorageService {
	return &StorageService{client: client}
}

type PersistentVolumeInfo struct {
	Name        string            `json:"name"`
	Capacity    string            `json:"capacity"`
	AccessModes []string          `json:"accessModes"`
	ReclaimPolicy string          `json:"reclaimPolicy"`
	Status      string            `json:"status"`
	Claim       string            `json:"claim,omitempty"`
	StorageClass string           `json:"storageClass,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Age         string            `json:"age"`
}

func (s *StorageService) ListPersistentVolumes(ctx context.Context) ([]PersistentVolumeInfo, error) {
	pvList, err := s.client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []PersistentVolumeInfo
	for _, pv := range pvList.Items {
		info := PersistentVolumeInfo{
			Name:         pv.Name,
			AccessModes:  formatAccessModes(pv.Spec.AccessModes),
			ReclaimPolicy: string(pv.Spec.PersistentVolumeReclaimPolicy),
			Status:       string(pv.Status.Phase),
			Labels:       pv.Labels,
			Age:          formatAge(pv.CreationTimestamp.Time),
		}

		if capacity, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
			info.Capacity = capacity.String()
		}

		if pv.Spec.StorageClassName != "" {
			info.StorageClass = pv.Spec.StorageClassName
		}

		if pv.Spec.ClaimRef != nil {
			info.Claim = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}

		if pv.Status.Phase == corev1.VolumeFailed {
			info.Reason = pv.Status.Reason
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *StorageService) GetPersistentVolume(ctx context.Context, name string) (*PersistentVolumeInfo, error) {
	pv, err := s.client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	info := &PersistentVolumeInfo{
		Name:         pv.Name,
		AccessModes:  formatAccessModes(pv.Spec.AccessModes),
		ReclaimPolicy: string(pv.Spec.PersistentVolumeReclaimPolicy),
		Status:       string(pv.Status.Phase),
		Labels:       pv.Labels,
		Age:          formatAge(pv.CreationTimestamp.Time),
	}

	if capacity, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
		info.Capacity = capacity.String()
	}

	if pv.Spec.StorageClassName != "" {
		info.StorageClass = pv.Spec.StorageClassName
	}

	if pv.Spec.ClaimRef != nil {
		info.Claim = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}

	return info, nil
}

func (s *StorageService) DeletePersistentVolume(ctx context.Context, name string) error {
	return s.client.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
}

type CreatePersistentVolumeRequest struct {
	Name          string            `json:"name"`
	Capacity      string            `json:"capacity"`
	AccessModes   []string          `json:"accessModes"`
	StorageClass  string            `json:"storageClass,omitempty"`
	ReclaimPolicy string            `json:"reclaimPolicy,omitempty"`
	HostPath      string            `json:"hostPath,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	MountOptions  []string          `json:"mountOptions,omitempty"`
	VolumeMode    string            `json:"volumeMode,omitempty"`
}

func (s *StorageService) CreatePersistentVolume(ctx context.Context, req CreatePersistentVolumeRequest) error {
	capacity, err := resource.ParseQuantity(req.Capacity)
	if err != nil {
		return fmt.Errorf("invalid capacity: %w", err)
	}

	var accessModes []corev1.PersistentVolumeAccessMode
	for _, mode := range req.AccessModes {
		accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(mode))
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: capacity,
			},
			AccessModes:      accessModes,
			MountOptions:     req.MountOptions,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{},
			},
		},
	}

	if req.StorageClass != "" {
		pv.Spec.StorageClassName = req.StorageClass
	}

	if req.ReclaimPolicy != "" {
		pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimPolicy(req.ReclaimPolicy)
	}

	if req.HostPath != "" {
		pv.Spec.PersistentVolumeSource.HostPath = &corev1.HostPathVolumeSource{
			Path: req.HostPath,
		}
	}

	if req.VolumeMode != "" {
		vm := corev1.PersistentVolumeMode(req.VolumeMode)
		pv.Spec.VolumeMode = &vm
	}

	_, err = s.client.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{})
	return err
}

type PersistentVolumeClaimInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	Volume       string            `json:"volume,omitempty"`
	Capacity     string            `json:"capacity,omitempty"`
	AccessModes  []string          `json:"accessModes"`
	StorageClass string            `json:"storageClass,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Age          string            `json:"age"`
}

func (s *StorageService) ListPersistentVolumeClaims(ctx context.Context, namespace string) ([]PersistentVolumeClaimInfo, error) {
	pvcList, err := s.client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []PersistentVolumeClaimInfo
	for _, pvc := range pvcList.Items {
		info := PersistentVolumeClaimInfo{
			Name:         pvc.Name,
			Namespace:    pvc.Namespace,
			Status:       string(pvc.Status.Phase),
			AccessModes:  formatAccessModes(pvc.Status.AccessModes),
			Labels:       pvc.Labels,
			Age:          formatAge(pvc.CreationTimestamp.Time),
		}

		if pvc.Spec.VolumeName != "" {
			info.Volume = pvc.Spec.VolumeName
		}

		if pvc.Spec.StorageClassName != nil {
			info.StorageClass = *pvc.Spec.StorageClassName
		}

		if capacity, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			info.Capacity = capacity.String()
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *StorageService) DeletePersistentVolumeClaim(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type CreatePersistentVolumeClaimRequest struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Capacity     string            `json:"capacity"`
	AccessModes  []string          `json:"accessModes"`
	StorageClass string            `json:"storageClass,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	VolumeMode   string            `json:"volumeMode,omitempty"`
	DataSource   *DataSourceRef    `json:"dataSource,omitempty"`
}

type DataSourceRef struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	APIGroup string `json:"apiGroup,omitempty"`
}

func (s *StorageService) CreatePersistentVolumeClaim(ctx context.Context, req CreatePersistentVolumeClaimRequest) error {
	capacity, err := resource.ParseQuantity(req.Capacity)
	if err != nil {
		return fmt.Errorf("invalid capacity: %w", err)
	}

	var accessModes []corev1.PersistentVolumeAccessMode
	for _, mode := range req.AccessModes {
		accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(mode))
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: capacity,
				},
			},
		},
	}

	if req.StorageClass != "" {
		pvc.Spec.StorageClassName = &req.StorageClass
	}

	if req.VolumeMode != "" {
		vm := corev1.PersistentVolumeMode(req.VolumeMode)
		pvc.Spec.VolumeMode = &vm
	}

	if req.DataSource != nil {
		ds := &corev1.TypedLocalObjectReference{
			Name: req.DataSource.Name,
			Kind: req.DataSource.Kind,
		}
		if req.DataSource.APIGroup != "" {
			ds.APIGroup = &req.DataSource.APIGroup
		}
		pvc.Spec.DataSource = ds
	}

	_, err = s.client.CoreV1().PersistentVolumeClaims(req.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

type StorageClassInfo struct {
	Name                 string            `json:"name"`
	Provisioner          string            `json:"provisioner"`
	ReclaimPolicy        string            `json:"reclaimPolicy"`
	VolumeBindingMode    string            `json:"volumeBindingMode"`
	AllowVolumeExpansion bool              `json:"allowVolumeExpansion"`
	Labels               map[string]string `json:"labels,omitempty"`
	Age                  string            `json:"age"`
}

func (s *StorageService) ListStorageClasses(ctx context.Context) ([]StorageClassInfo, error) {
	scList, err := s.client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []StorageClassInfo
	for _, sc := range scList.Items {
		info := StorageClassInfo{
			Name:                 sc.Name,
			Provisioner:          sc.Provisioner,
			Labels:               sc.Labels,
			Age:                  formatAge(sc.CreationTimestamp.Time),
			AllowVolumeExpansion: sc.AllowVolumeExpansion != nil && *sc.AllowVolumeExpansion,
		}

		if sc.ReclaimPolicy != nil {
			info.ReclaimPolicy = string(*sc.ReclaimPolicy)
		}

		if sc.VolumeBindingMode != nil {
			info.VolumeBindingMode = string(*sc.VolumeBindingMode)
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *StorageService) DeleteStorageClass(ctx context.Context, name string) error {
	return s.client.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
}

type CreateStorageClassRequest struct {
	Name                 string            `json:"name"`
	Provisioner          string            `json:"provisioner"`
	ReclaimPolicy        string            `json:"reclaimPolicy,omitempty"`
	VolumeBindingMode    string            `json:"volumeBindingMode,omitempty"`
	AllowVolumeExpansion bool              `json:"allowVolumeExpansion"`
	Labels               map[string]string `json:"labels,omitempty"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	Parameters           map[string]string `json:"parameters,omitempty"`
}

func (s *StorageService) CreateStorageClass(ctx context.Context, req CreateStorageClassRequest) error {
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Provisioner:          req.Provisioner,
		AllowVolumeExpansion: &req.AllowVolumeExpansion,
		Parameters:           req.Parameters,
	}

	if req.ReclaimPolicy != "" {
		policy := corev1.PersistentVolumeReclaimPolicy(req.ReclaimPolicy)
		sc.ReclaimPolicy = &policy
	}

	if req.VolumeBindingMode != "" {
		mode := storagev1.VolumeBindingMode(req.VolumeBindingMode)
		sc.VolumeBindingMode = &mode
	}

	_, err := s.client.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{})
	return err
}

func formatAccessModes(modes []corev1.PersistentVolumeAccessMode) []string {
	var result []string
	for _, mode := range modes {
		result = append(result, string(mode))
	}
	return result
}

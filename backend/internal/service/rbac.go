package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RBACService struct {
	client kubernetes.Interface
}

func NewRBACService(client kubernetes.Interface) *RBACService {
	return &RBACService{client: client}
}

type RoleInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Rules       []PolicyRuleInfo  `json:"rules"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

type PolicyRuleInfo struct {
	APIGroups     []string `json:"apiGroups"`
	Resources     []string `json:"resources"`
	Verbs         []string `json:"verbs"`
	ResourceNames []string `json:"resourceNames,omitempty"`
	NonResourceURLs []string `json:"nonResourceURLs,omitempty"`
}

func (s *RBACService) ListRoles(ctx context.Context, namespace string) ([]RoleInfo, error) {
	roleList, err := s.client.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []RoleInfo
	for _, role := range roleList.Items {
		info := RoleInfo{
			Name:        role.Name,
			Namespace:   role.Namespace,
			Labels:      role.Labels,
			Annotations: role.Annotations,
			Age:         formatAge(role.CreationTimestamp.Time),
		}

		for _, rule := range role.Rules {
			info.Rules = append(info.Rules, PolicyRuleInfo{
				APIGroups:     rule.APIGroups,
				Resources:     rule.Resources,
				Verbs:         rule.Verbs,
				ResourceNames: rule.ResourceNames,
			})
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *RBACService) GetRole(ctx context.Context, namespace, name string) (*RoleInfo, error) {
	role, err := s.client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	info := &RoleInfo{
		Name:        role.Name,
		Namespace:   role.Namespace,
		Labels:      role.Labels,
		Annotations: role.Annotations,
		Age:         formatAge(role.CreationTimestamp.Time),
	}

	for _, rule := range role.Rules {
		info.Rules = append(info.Rules, PolicyRuleInfo{
			APIGroups:     rule.APIGroups,
			Resources:     rule.Resources,
			Verbs:         rule.Verbs,
			ResourceNames: rule.ResourceNames,
		})
	}

	return info, nil
}

func (s *RBACService) DeleteRole(ctx context.Context, namespace, name string) error {
	return s.client.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *RBACService) UpdateRole(ctx context.Context, namespace, name string, rules []PolicyRuleInfo) (*RoleInfo, error) {
	role, err := s.client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	role.Rules = nil
	for _, r := range rules {
		role.Rules = append(role.Rules, rbacv1.PolicyRule{
			APIGroups:       r.APIGroups,
			Resources:       r.Resources,
			Verbs:           r.Verbs,
			ResourceNames:   r.ResourceNames,
			NonResourceURLs: r.NonResourceURLs,
		})
	}

	updated, err := s.client.RbacV1().Roles(namespace).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	info := &RoleInfo{
		Name:        updated.Name,
		Namespace:   updated.Namespace,
		Labels:      updated.Labels,
		Annotations: updated.Annotations,
		Age:         formatAge(updated.CreationTimestamp.Time),
	}
	for _, rule := range updated.Rules {
		info.Rules = append(info.Rules, PolicyRuleInfo{
			APIGroups:       rule.APIGroups,
			Resources:       rule.Resources,
			Verbs:           rule.Verbs,
			ResourceNames:   rule.ResourceNames,
			NonResourceURLs: rule.NonResourceURLs,
		})
	}
	return info, nil
}

type ClusterRoleInfo struct {
	Name        string            `json:"name"`
	Rules       []PolicyRuleInfo  `json:"rules"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

func (s *RBACService) ListClusterRoles(ctx context.Context) ([]ClusterRoleInfo, error) {
	crList, err := s.client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ClusterRoleInfo
	for _, cr := range crList.Items {
		info := ClusterRoleInfo{
			Name:        cr.Name,
			Labels:      cr.Labels,
			Annotations: cr.Annotations,
			Age:         formatAge(cr.CreationTimestamp.Time),
		}

		for _, rule := range cr.Rules {
			info.Rules = append(info.Rules, PolicyRuleInfo{
				APIGroups:       rule.APIGroups,
				Resources:       rule.Resources,
				Verbs:           rule.Verbs,
				ResourceNames:   rule.ResourceNames,
				NonResourceURLs: rule.NonResourceURLs,
			})
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *RBACService) GetClusterRole(ctx context.Context, name string) (*ClusterRoleInfo, error) {
	cr, err := s.client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	info := &ClusterRoleInfo{
		Name:        cr.Name,
		Labels:      cr.Labels,
		Annotations: cr.Annotations,
		Age:         formatAge(cr.CreationTimestamp.Time),
	}

	for _, rule := range cr.Rules {
		info.Rules = append(info.Rules, PolicyRuleInfo{
			APIGroups:       rule.APIGroups,
			Resources:       rule.Resources,
			Verbs:           rule.Verbs,
			ResourceNames:   rule.ResourceNames,
			NonResourceURLs: rule.NonResourceURLs,
		})
	}

	return info, nil
}

func (s *RBACService) DeleteClusterRole(ctx context.Context, name string) error {
	return s.client.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *RBACService) UpdateClusterRole(ctx context.Context, name string, rules []PolicyRuleInfo) (*ClusterRoleInfo, error) {
	cr, err := s.client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cr.Rules = nil
	for _, r := range rules {
		cr.Rules = append(cr.Rules, rbacv1.PolicyRule{
			APIGroups:       r.APIGroups,
			Resources:       r.Resources,
			Verbs:           r.Verbs,
			ResourceNames:   r.ResourceNames,
			NonResourceURLs: r.NonResourceURLs,
		})
	}

	updated, err := s.client.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	info := &ClusterRoleInfo{
		Name:        updated.Name,
		Labels:      updated.Labels,
		Annotations: updated.Annotations,
		Age:         formatAge(updated.CreationTimestamp.Time),
	}
	for _, rule := range updated.Rules {
		info.Rules = append(info.Rules, PolicyRuleInfo{
			APIGroups:       rule.APIGroups,
			Resources:       rule.Resources,
			Verbs:           rule.Verbs,
			ResourceNames:   rule.ResourceNames,
			NonResourceURLs: rule.NonResourceURLs,
		})
	}
	return info, nil
}

type RoleBindingInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	RoleRef     RoleRefInfo       `json:"roleRef"`
	Subjects    []SubjectInfo     `json:"subjects"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

type RoleRefInfo struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type SubjectInfo struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

func (s *RBACService) ListRoleBindings(ctx context.Context, namespace string) ([]RoleBindingInfo, error) {
	rbList, err := s.client.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []RoleBindingInfo
	for _, rb := range rbList.Items {
		info := RoleBindingInfo{
			Name:        rb.Name,
			Namespace:   rb.Namespace,
			Labels:      rb.Labels,
			Annotations: rb.Annotations,
			Age:         formatAge(rb.CreationTimestamp.Time),
			RoleRef: RoleRefInfo{
				Kind: rb.RoleRef.Kind,
				Name: rb.RoleRef.Name,
			},
		}

		for _, subject := range rb.Subjects {
			info.Subjects = append(info.Subjects, SubjectInfo{
				Kind:      subject.Kind,
				Name:      subject.Name,
				Namespace: subject.Namespace,
			})
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *RBACService) DeleteRoleBinding(ctx context.Context, namespace, name string) error {
	return s.client.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type ClusterRoleBindingInfo struct {
	Name        string            `json:"name"`
	RoleRef     RoleRefInfo       `json:"roleRef"`
	Subjects    []SubjectInfo     `json:"subjects"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

func (s *RBACService) ListClusterRoleBindings(ctx context.Context) ([]ClusterRoleBindingInfo, error) {
	crbList, err := s.client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ClusterRoleBindingInfo
	for _, crb := range crbList.Items {
		info := ClusterRoleBindingInfo{
			Name:        crb.Name,
			Labels:      crb.Labels,
			Annotations: crb.Annotations,
			Age:         formatAge(crb.CreationTimestamp.Time),
			RoleRef: RoleRefInfo{
				Kind: crb.RoleRef.Kind,
				Name: crb.RoleRef.Name,
			},
		}

		for _, subject := range crb.Subjects {
			info.Subjects = append(info.Subjects, SubjectInfo{
				Kind:      subject.Kind,
				Name:      subject.Name,
				Namespace: subject.Namespace,
			})
		}

		result = append(result, info)
	}
	return result, nil
}

func (s *RBACService) DeleteClusterRoleBinding(ctx context.Context, name string) error {
	return s.client.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
}

type ServiceAccountInfo struct {
	Name                      string                      `json:"name"`
	Namespace                 string                      `json:"namespace"`
	Labels                    map[string]string           `json:"labels,omitempty"`
	Annotations               map[string]string           `json:"annotations,omitempty"`
	AutomountServiceAccountToken *bool                    `json:"automountServiceAccountToken,omitempty"`
	ImagePullSecrets          []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Age                       string                      `json:"age"`
}

func (s *RBACService) ListServiceAccounts(ctx context.Context, namespace string) ([]ServiceAccountInfo, error) {
	saList, err := s.client.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ServiceAccountInfo
	for _, sa := range saList.Items {
		result = append(result, ServiceAccountInfo{
			Name:                      sa.Name,
			Namespace:                 sa.Namespace,
			Labels:                    sa.Labels,
			Annotations:               sa.Annotations,
			AutomountServiceAccountToken: sa.AutomountServiceAccountToken,
			ImagePullSecrets:          sa.ImagePullSecrets,
			Age:                       formatAge(sa.CreationTimestamp.Time),
		})
	}
	return result, nil
}

func (s *RBACService) DeleteServiceAccount(ctx context.Context, namespace, name string) error {
	return s.client.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *RBACService) CreateRole(ctx context.Context, namespace, name string, labels, annotations map[string]string, rules []PolicyRuleInfo) (*RoleInfo, error) {
	var k8sRules []rbacv1.PolicyRule
	for _, r := range rules {
		k8sRules = append(k8sRules, rbacv1.PolicyRule{
			APIGroups:       r.APIGroups,
			Resources:       r.Resources,
			Verbs:           r.Verbs,
			ResourceNames:   r.ResourceNames,
			NonResourceURLs: r.NonResourceURLs,
		})
	}

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Rules: k8sRules,
	}

	created, err := s.client.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	info := &RoleInfo{
		Name:        created.Name,
		Namespace:   created.Namespace,
		Labels:      created.Labels,
		Annotations: created.Annotations,
		Age:         formatAge(created.CreationTimestamp.Time),
	}
	for _, rule := range created.Rules {
		info.Rules = append(info.Rules, PolicyRuleInfo{
			APIGroups:       rule.APIGroups,
			Resources:       rule.Resources,
			Verbs:           rule.Verbs,
			ResourceNames:   rule.ResourceNames,
			NonResourceURLs: rule.NonResourceURLs,
		})
	}
	return info, nil
}

func (s *RBACService) CreateClusterRole(ctx context.Context, name string, labels, annotations map[string]string, rules []PolicyRuleInfo) (*ClusterRoleInfo, error) {
	var k8sRules []rbacv1.PolicyRule
	for _, r := range rules {
		k8sRules = append(k8sRules, rbacv1.PolicyRule{
			APIGroups:       r.APIGroups,
			Resources:       r.Resources,
			Verbs:           r.Verbs,
			ResourceNames:   r.ResourceNames,
			NonResourceURLs: r.NonResourceURLs,
		})
	}

	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		Rules: k8sRules,
	}

	created, err := s.client.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	info := &ClusterRoleInfo{
		Name:        created.Name,
		Labels:      created.Labels,
		Annotations: created.Annotations,
		Age:         formatAge(created.CreationTimestamp.Time),
	}
	for _, rule := range created.Rules {
		info.Rules = append(info.Rules, PolicyRuleInfo{
			APIGroups:       rule.APIGroups,
			Resources:       rule.Resources,
			Verbs:           rule.Verbs,
			ResourceNames:   rule.ResourceNames,
			NonResourceURLs: rule.NonResourceURLs,
		})
	}
	return info, nil
}

func (s *RBACService) CreateRoleBinding(ctx context.Context, namespace, name string, labels map[string]string, roleRef RoleRefInfo, subjects []SubjectInfo) (*RoleBindingInfo, error) {
	var k8sSubjects []rbacv1.Subject
	for _, sub := range subjects {
		k8sSubjects = append(k8sSubjects, rbacv1.Subject{
			Kind:      sub.Kind,
			Name:      sub.Name,
			Namespace: sub.Namespace,
		})
	}

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleRef.Kind,
			Name:     roleRef.Name,
		},
		Subjects: k8sSubjects,
	}

	created, err := s.client.RbacV1().RoleBindings(namespace).Create(ctx, rb, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	info := &RoleBindingInfo{
		Name:        created.Name,
		Namespace:   created.Namespace,
		Labels:      created.Labels,
		Annotations: created.Annotations,
		Age:         formatAge(created.CreationTimestamp.Time),
		RoleRef: RoleRefInfo{
			Kind: created.RoleRef.Kind,
			Name: created.RoleRef.Name,
		},
	}
	for _, subject := range created.Subjects {
		info.Subjects = append(info.Subjects, SubjectInfo{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}
	return info, nil
}

func (s *RBACService) CreateClusterRoleBinding(ctx context.Context, name string, labels map[string]string, roleRef RoleRefInfo, subjects []SubjectInfo) (*ClusterRoleBindingInfo, error) {
	var k8sSubjects []rbacv1.Subject
	for _, sub := range subjects {
		k8sSubjects = append(k8sSubjects, rbacv1.Subject{
			Kind:      sub.Kind,
			Name:      sub.Name,
			Namespace: sub.Namespace,
		})
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleRef.Kind,
			Name:     roleRef.Name,
		},
		Subjects: k8sSubjects,
	}

	created, err := s.client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	info := &ClusterRoleBindingInfo{
		Name:        created.Name,
		Labels:      created.Labels,
		Annotations: created.Annotations,
		Age:         formatAge(created.CreationTimestamp.Time),
		RoleRef: RoleRefInfo{
			Kind: created.RoleRef.Kind,
			Name: created.RoleRef.Name,
		},
	}
	for _, subject := range created.Subjects {
		info.Subjects = append(info.Subjects, SubjectInfo{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}
	return info, nil
}

func (s *RBACService) CreateServiceAccount(ctx context.Context, namespace, name string, labels, annotations map[string]string, automountServiceAccountToken *bool, imagePullSecrets []corev1.LocalObjectReference) (*ServiceAccountInfo, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		AutomountServiceAccountToken: automountServiceAccountToken,
		ImagePullSecrets:             imagePullSecrets,
	}

	created, err := s.client.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &ServiceAccountInfo{
		Name:                      created.Name,
		Namespace:                 created.Namespace,
		Labels:                    created.Labels,
		Annotations:               created.Annotations,
		AutomountServiceAccountToken: created.AutomountServiceAccountToken,
		ImagePullSecrets:          created.ImagePullSecrets,
		Age:                       formatAge(created.CreationTimestamp.Time),
	}, nil
}

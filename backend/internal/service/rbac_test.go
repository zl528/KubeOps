package service

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRBACService_ListRoles(t *testing.T) {
	roles := &rbacv1.RoleList{
		Items: []rbacv1.Role{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-reader",
					Namespace: "default",
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{""},
						Resources: []string{"pods"},
						Verbs:     []string{"get", "list", "watch"},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(roles)
	svc := NewRBACService(client)

	result, err := svc.ListRoles(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 role, got %d", len(result))
	}

	if result[0].Name != "pod-reader" {
		t.Errorf("expected name pod-reader, got %s", result[0].Name)
	}

	if len(result[0].Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(result[0].Rules))
	}
}

func TestRBACService_ListClusterRoles(t *testing.T) {
	clusterRoles := &rbacv1.ClusterRoleList{
		Items: []rbacv1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "admin",
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{"*"},
						Resources: []string{"*"},
						Verbs:     []string{"*"},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(clusterRoles)
	svc := NewRBACService(client)

	result, err := svc.ListClusterRoles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 cluster role, got %d", len(result))
	}

	if result[0].Name != "admin" {
		t.Errorf("expected name admin, got %s", result[0].Name)
	}
}

func TestRBACService_ListRoleBindings(t *testing.T) {
	roleBindings := &rbacv1.RoleBindingList{
		Items: []rbacv1.RoleBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "read-pods",
					Namespace: "default",
				},
				RoleRef: rbacv1.RoleRef{
					Kind: "Role",
					Name: "pod-reader",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "User",
						Name:      "jane",
						Namespace: "default",
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(roleBindings)
	svc := NewRBACService(client)

	result, err := svc.ListRoleBindings(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 role binding, got %d", len(result))
	}

	if result[0].RoleRef.Name != "pod-reader" {
		t.Errorf("expected role ref pod-reader, got %s", result[0].RoleRef.Name)
	}

	if len(result[0].Subjects) != 1 {
		t.Fatalf("expected 1 subject, got %d", len(result[0].Subjects))
	}

	if result[0].Subjects[0].Name != "jane" {
		t.Errorf("expected subject jane, got %s", result[0].Subjects[0].Name)
	}
}

func TestRBACService_ListServiceAccounts(t *testing.T) {
	sas := &corev1.ServiceAccountList{
		Items: []corev1.ServiceAccount{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(sas)
	svc := NewRBACService(client)

	result, err := svc.ListServiceAccounts(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 service account, got %d", len(result))
	}

	if result[0].Name != "default" {
		t.Errorf("expected name default, got %s", result[0].Name)
	}
}

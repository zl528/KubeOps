package service

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestStorageService_ListPersistentVolumes(t *testing.T) {
	pvs := &corev1.PersistentVolumeList{
		Items: []corev1.PersistentVolume{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv-1",
				},
				Spec: corev1.PersistentVolumeSpec{
					Capacity: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
					StorageClassName: "standard",
				},
				Status: corev1.PersistentVolumeStatus{
					Phase: corev1.VolumeBound,
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pvs)
	svc := NewStorageService(client)

	result, err := svc.ListPersistentVolumes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 pv, got %d", len(result))
	}

	if result[0].Name != "pv-1" {
		t.Errorf("expected name pv-1, got %s", result[0].Name)
	}

	if result[0].Capacity != "10Gi" {
		t.Errorf("expected capacity 10Gi, got %s", result[0].Capacity)
	}

	if result[0].Status != "Bound" {
		t.Errorf("expected status Bound, got %s", result[0].Status)
	}
}

func TestStorageService_ListPersistentVolumeClaims(t *testing.T) {
	pvcs := &corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pvc",
					Namespace: "default",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				},
				Status: corev1.PersistentVolumeClaimStatus{
					Phase: corev1.ClaimBound,
					Capacity: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("5Gi"),
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pvcs)
	svc := NewStorageService(client)

	result, err := svc.ListPersistentVolumeClaims(context.Background(), "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 pvc, got %d", len(result))
	}

	if result[0].Name != "my-pvc" {
		t.Errorf("expected name my-pvc, got %s", result[0].Name)
	}

	if result[0].Status != "Bound" {
		t.Errorf("expected status Bound, got %s", result[0].Status)
	}
}

func TestStorageService_ListStorageClasses(t *testing.T) {
	scs := &storagev1.StorageClassList{
		Items: []storagev1.StorageClass{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "standard",
				},
				Provisioner: "kubernetes.io/no-provisioner",
			},
		},
	}

	client := fake.NewSimpleClientset(scs)
	svc := NewStorageService(client)

	result, err := svc.ListStorageClasses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 storage class, got %d", len(result))
	}

	if result[0].Name != "standard" {
		t.Errorf("expected name standard, got %s", result[0].Name)
	}

	if result[0].Provisioner != "kubernetes.io/no-provisioner" {
		t.Errorf("expected provisioner kubernetes.io/no-provisioner, got %s", result[0].Provisioner)
	}
}

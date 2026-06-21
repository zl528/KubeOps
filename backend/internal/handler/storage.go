package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type StorageHandler struct {
	clientGetter ClientGetter
}

func NewStorageHandler(clientGetter ClientGetter) *StorageHandler {
	return &StorageHandler{clientGetter: clientGetter}
}

func (h *StorageHandler) HandleListPersistentVolumes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	result, err := svc.ListPersistentVolumes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *StorageHandler) HandleGetPersistentVolume(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	result, err := svc.GetPersistentVolume(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *StorageHandler) HandleDeletePersistentVolume(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	if err := svc.DeletePersistentVolume(ctx, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *StorageHandler) HandleCreatePersistentVolume(w http.ResponseWriter, r *http.Request) {
	var req service.CreatePersistentVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Capacity == "" {
		writeError(w, http.StatusBadRequest, "name and capacity are required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	if err := svc.CreatePersistentVolume(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *StorageHandler) HandleListPersistentVolumeClaims(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	result, err := svc.ListPersistentVolumeClaims(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *StorageHandler) HandleDeletePersistentVolumeClaim(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	if err := svc.DeletePersistentVolumeClaim(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *StorageHandler) HandleCreatePersistentVolumeClaim(w http.ResponseWriter, r *http.Request) {
	var req service.CreatePersistentVolumeClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Namespace == "" || req.Capacity == "" {
		writeError(w, http.StatusBadRequest, "name, namespace, and capacity are required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	if err := svc.CreatePersistentVolumeClaim(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *StorageHandler) HandleListStorageClasses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	result, err := svc.ListStorageClasses(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *StorageHandler) HandleDeleteStorageClass(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteStorageClass(ctx, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *StorageHandler) HandleCreateStorageClass(w http.ResponseWriter, r *http.Request) {
	var req service.CreateStorageClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Provisioner == "" {
		writeError(w, http.StatusBadRequest, "name and provisioner are required")
		return
	}
	ctx := context.Background()
	svc := service.NewStorageService(h.clientGetter.GetK8sClient())
	if err := svc.CreateStorageClass(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

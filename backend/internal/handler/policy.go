package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type PolicyHandler struct {
	clientGetter ClientGetter
}

func NewPolicyHandler(clientGetter ClientGetter) *PolicyHandler {
	return &PolicyHandler{clientGetter: clientGetter}
}

func (h *PolicyHandler) HandleListLimitRanges(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewPolicyService(h.clientGetter.GetK8sClient())
	result, err := svc.ListLimitRanges(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *PolicyHandler) HandleDeleteLimitRange(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewPolicyService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteLimitRange(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *PolicyHandler) HandleCreateLimitRange(w http.ResponseWriter, r *http.Request) {
	var req service.CreateLimitRangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewPolicyService(h.clientGetter.GetK8sClient())
	if err := svc.CreateLimitRange(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *PolicyHandler) HandleListHPAs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewPolicyService(h.clientGetter.GetK8sClient())
	result, err := svc.ListHPAs(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *PolicyHandler) HandleDeleteHPA(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewPolicyService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteHPA(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *PolicyHandler) HandleCreateHPA(w http.ResponseWriter, r *http.Request) {
	var req service.CreateHPARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" || req.TargetKind == "" || req.TargetName == "" {
		writeError(w, http.StatusBadRequest, "namespace, name, targetKind, and targetName are required")
		return
	}
	if req.MinReplicas <= 0 {
		req.MinReplicas = 1
	}
	if req.MaxReplicas < req.MinReplicas {
		req.MaxReplicas = req.MinReplicas
	}
	ctx := context.Background()
	svc := service.NewPolicyService(h.clientGetter.GetK8sClient())
	if err := svc.CreateHPA(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

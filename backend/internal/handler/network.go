package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type NetworkHandler struct {
	clientGetter ClientGetter
}

func NewNetworkHandler(clientGetter ClientGetter) *NetworkHandler {
	return &NetworkHandler{clientGetter: clientGetter}
}

func (h *NetworkHandler) HandleListIngresses(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	result, err := svc.ListIngresses(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *NetworkHandler) HandleDeleteIngress(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteIngress(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *NetworkHandler) HandleGetIngress(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	result, err := svc.GetIngress(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *NetworkHandler) HandleCreateIngress(w http.ResponseWriter, r *http.Request) {
	var req service.CreateIngressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	if err := svc.CreateIngress(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *NetworkHandler) HandleListNetworkPolicies(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	result, err := svc.ListNetworkPolicies(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *NetworkHandler) HandleDeleteNetworkPolicy(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteNetworkPolicy(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *NetworkHandler) HandleGetNetworkPolicy(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	result, err := svc.GetNetworkPolicy(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *NetworkHandler) HandleCreateNetworkPolicy(w http.ResponseWriter, r *http.Request) {
	var req service.CreateNetworkPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	if err := svc.CreateNetworkPolicy(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *NetworkHandler) HandleListEndpoints(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewNetworkService(h.clientGetter.GetK8sClient())
	result, err := svc.ListEndpoints(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

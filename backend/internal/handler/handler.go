package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClientGetter interface {
	GetK8sClient() kubernetes.Interface
	GetRestConfig() *rest.Config
}

type Handler struct {
	clientGetter ClientGetter
	clusterSvc   *service.ClusterService
}

func NewHandler(clientGetter ClientGetter) *Handler {
	return &Handler{clientGetter: clientGetter}
}

func (h *Handler) HandleClusterOverview(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	overview, err := svc.GetOverview(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: overview})
}

func (h *Handler) HandleListNodes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	nodes, err := svc.ListNodes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: nodes})
}

func (h *Handler) HandleGetNode(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	node, err := svc.GetNode(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: node})
}

func (h *Handler) HandleDrainNode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	if err := svc.DrainNode(ctx, req.Name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "node drained successfully"})
}

func (h *Handler) HandleUncordonNode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	if err := svc.UncordonNode(ctx, req.Name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "node uncordoned successfully"})
}

func (h *Handler) HandleListNamespaces(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	namespaces, err := svc.ListNamespaces(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: namespaces})
}

func (h *Handler) HandleCreateNamespace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string            `json:"name"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.CreateNamespace(ctx, req.Name, req.Labels, req.Annotations); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *Handler) HandleDeleteNamespace(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteNamespace(ctx, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *Handler) HandleCreateService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace   string                      `json:"namespace"`
		Name        string                      `json:"name"`
		Type        string                      `json:"type"`
		Ports       []service.ServicePortConfig  `json:"ports"`
		Selector    map[string]string            `json:"selector"`
		Labels      map[string]string            `json:"labels,omitempty"`
		Annotations map[string]string            `json:"annotations,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	if req.Type == "" {
		req.Type = "ClusterIP"
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.CreateService(ctx, req.Namespace, req.Name, req.Type, req.Ports, req.Selector, req.Labels, req.Annotations); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *Handler) HandleUpdateService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace   string                      `json:"namespace"`
		Name        string                      `json:"name"`
		Type        string                      `json:"type,omitempty"`
		Ports       []service.ServicePortConfig  `json:"ports,omitempty"`
		Selector    map[string]string            `json:"selector,omitempty"`
		Labels      map[string]string            `json:"labels,omitempty"`
		Annotations map[string]string            `json:"annotations,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.UpdateService(ctx, req.Namespace, req.Name, req.Type, req.Ports, req.Selector, req.Labels, req.Annotations); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *Handler) HandleListPods(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	pods, err := svc.ListPods(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: pods})
}

func (h *Handler) HandleListDeployments(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	deploys, err := svc.ListDeployments(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: deploys})
}

func (h *Handler) HandleListServices(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	svcs, err := svc.ListServices(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: svcs})
}

func (h *Handler) HandleGetService(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	result, err := svc.GetService(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	events, err := svc.ListEvents(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: events})
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok"})
}

func (h *Handler) HandleDeleteService(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewClusterService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteService(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, model.Response{Code: code, Message: msg})
}

package handler

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type RBACHandler struct {
	clientGetter ClientGetter
}

func NewRBACHandler(clientGetter ClientGetter) *RBACHandler {
	return &RBACHandler{clientGetter: clientGetter}
}

func (h *RBACHandler) HandleListRoles(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.ListRoles(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleGetRole(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.GetRole(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleDeleteRole(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteRole(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *RBACHandler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string                  `json:"namespace"`
		Name      string                  `json:"name"`
		Rules     []service.PolicyRuleInfo `json:"rules"`
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
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.UpdateRole(ctx, req.Namespace, req.Name, req.Rules)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully", Data: result})
}

func (h *RBACHandler) HandleListClusterRoles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.ListClusterRoles(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleGetClusterRole(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.GetClusterRole(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleDeleteClusterRole(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteClusterRole(ctx, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *RBACHandler) HandleUpdateClusterRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string                  `json:"name"`
		Rules []service.PolicyRuleInfo `json:"rules"`
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
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.UpdateClusterRole(ctx, req.Name, req.Rules)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully", Data: result})
}

func (h *RBACHandler) HandleListRoleBindings(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.ListRoleBindings(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleDeleteRoleBinding(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteRoleBinding(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *RBACHandler) HandleListClusterRoleBindings(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.ListClusterRoleBindings(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleDeleteClusterRoleBinding(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteClusterRoleBinding(ctx, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *RBACHandler) HandleListServiceAccounts(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.ListServiceAccounts(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *RBACHandler) HandleDeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteServiceAccount(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *RBACHandler) HandleCreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace   string                  `json:"namespace"`
		Name        string                  `json:"name"`
		Labels      map[string]string       `json:"labels"`
		Annotations map[string]string       `json:"annotations"`
		Rules       []service.PolicyRuleInfo `json:"rules"`
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
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.CreateRole(ctx, req.Namespace, req.Name, req.Labels, req.Annotations, req.Rules)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully", Data: result})
}

func (h *RBACHandler) HandleCreateClusterRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                  `json:"name"`
		Labels      map[string]string       `json:"labels"`
		Annotations map[string]string       `json:"annotations"`
		Rules       []service.PolicyRuleInfo `json:"rules"`
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
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.CreateClusterRole(ctx, req.Name, req.Labels, req.Annotations, req.Rules)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully", Data: result})
}

func (h *RBACHandler) HandleCreateRoleBinding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string                `json:"namespace"`
		Name      string                `json:"name"`
		Labels    map[string]string     `json:"labels"`
		RoleRef   service.RoleRefInfo   `json:"roleRef"`
		Subjects  []service.SubjectInfo `json:"subjects"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	if req.RoleRef.Name == "" {
		writeError(w, http.StatusBadRequest, "roleRef name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.CreateRoleBinding(ctx, req.Namespace, req.Name, req.Labels, req.RoleRef, req.Subjects)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully", Data: result})
}

func (h *RBACHandler) HandleCreateClusterRoleBinding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string                `json:"name"`
		Labels   map[string]string     `json:"labels"`
		RoleRef  service.RoleRefInfo   `json:"roleRef"`
		Subjects []service.SubjectInfo `json:"subjects"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.RoleRef.Name == "" {
		writeError(w, http.StatusBadRequest, "roleRef name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.CreateClusterRoleBinding(ctx, req.Name, req.Labels, req.RoleRef, req.Subjects)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully", Data: result})
}

func (h *RBACHandler) HandleCreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace                    string                       `json:"namespace"`
		Name                         string                       `json:"name"`
		Labels                       map[string]string            `json:"labels"`
		Annotations                  map[string]string            `json:"annotations"`
		AutomountServiceAccountToken *bool                        `json:"automountServiceAccountToken"`
		ImagePullSecrets             []corev1.LocalObjectReference `json:"imagePullSecrets"`
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
	svc := service.NewRBACService(h.clientGetter.GetK8sClient())
	result, err := svc.CreateServiceAccount(ctx, req.Namespace, req.Name, req.Labels, req.Annotations, req.AutomountServiceAccountToken, req.ImagePullSecrets)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully", Data: result})
}

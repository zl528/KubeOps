package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type ResourceHandler struct {
	clientGetter ClientGetter
}

func NewResourceHandler(clientGetter ClientGetter) *ResourceHandler {
	return &ResourceHandler{clientGetter: clientGetter}
}

func (h *ResourceHandler) HandleGetPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")
	tailLines := int64(100)

	if tl := r.URL.Query().Get("tailLines"); tl != "" {
		if v, err := strconv.ParseInt(tl, 10, 64); err == nil {
			tailLines = v
		}
	}

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	logs, err := svc.GetPodLogs(ctx, namespace, pod, container, tailLines)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: map[string]string{"logs": logs}})
}

func (h *ResourceHandler) HandleScaleDeployment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Replicas  int32  `json:"replicas"`
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
	if err := svc.ScaleDeployment(ctx, req.Namespace, req.Name, req.Replicas); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "scaled successfully"})
}

func (h *ResourceHandler) HandleRollbackDeployment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Revision  int    `json:"revision"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.RollbackDeployment(ctx, req.Namespace, req.Name, req.Revision); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "rolled back successfully"})
}

func (h *ResourceHandler) HandleUpdateDeployment(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateDeploymentRequest
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
	if err := svc.UpdateDeployment(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *ResourceHandler) HandleDeleteDeployment(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteDeployment(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *ResourceHandler) HandleRestartDeployment(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.RestartDeployment(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "restarted successfully"})
}

func (h *ResourceHandler) HandleGetDeployment(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	deploy, err := svc.GetDeployment(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: deploy})
}

func (h *ResourceHandler) HandleListConfigMaps(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	cms, err := svc.ListConfigMaps(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: cms})
}

func (h *ResourceHandler) HandleGetConfigMap(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	cm, err := svc.GetConfigMap(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: cm})
}

func (h *ResourceHandler) HandleDeleteConfigMap(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteConfigMap(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *ResourceHandler) HandleCreateConfigMap(w http.ResponseWriter, r *http.Request) {
	var req service.ConfigMapCreateRequest
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
	if err := svc.CreateConfigMap(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *ResourceHandler) HandleUpdateConfigMap(w http.ResponseWriter, r *http.Request) {
	var req service.ConfigMapCreateRequest
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
	if err := svc.UpdateConfigMap(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *ResourceHandler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	secrets, err := svc.ListSecrets(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: secrets})
}

func (h *ResourceHandler) HandleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteSecret(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *ResourceHandler) HandleCreateSecret(w http.ResponseWriter, r *http.Request) {
	var req service.SecretCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	if req.Type == "" {
		req.Type = "Opaque"
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.CreateSecret(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *ResourceHandler) HandleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	var req service.SecretCreateRequest
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
	if err := svc.UpdateSecret(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *ResourceHandler) HandleDeletePod(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.DeletePod(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *ResourceHandler) HandleRestartPod(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.RestartPod(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "restarted successfully"})
}

func (h *ResourceHandler) HandleGetPod(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	pod, err := svc.GetPod(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: pod})
}

func (h *ResourceHandler) HandleListResourceQuotas(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	quotas, err := svc.ListResourceQuotas(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: quotas})
}

func (h *ResourceHandler) HandleGetResourceQuota(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	rq, err := svc.GetResourceQuota(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: rq})
}

func (h *ResourceHandler) HandleCreateResourceQuota(w http.ResponseWriter, r *http.Request) {
	var req service.ResourceQuotaCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Namespace == "" {
		writeError(w, http.StatusBadRequest, "name and namespace are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.CreateResourceQuota(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *ResourceHandler) HandleUpdateResourceQuota(w http.ResponseWriter, r *http.Request) {
	var req service.ResourceQuotaCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Namespace == "" {
		writeError(w, http.StatusBadRequest, "name and namespace are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.UpdateResourceQuota(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *ResourceHandler) HandleDeleteResourceQuota(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewResourceService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteResourceQuota(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

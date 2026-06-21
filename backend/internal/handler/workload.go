package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type WorkloadHandler struct {
	clientGetter ClientGetter
}

func NewWorkloadHandler(clientGetter ClientGetter) *WorkloadHandler {
	return &WorkloadHandler{clientGetter: clientGetter}
}

func (h *WorkloadHandler) HandleListStatefulSets(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.ListStatefulSets(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleGetStatefulSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.GetStatefulSet(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleScaleStatefulSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	replicasStr := r.URL.Query().Get("replicas")

	if namespace == "" || name == "" || replicasStr == "" {
		writeError(w, http.StatusBadRequest, "namespace, name and replicas are required")
		return
	}

	var replicas int32
	if _, err := fmt.Sscanf(replicasStr, "%d", &replicas); err != nil {
		writeError(w, http.StatusBadRequest, "invalid replicas value")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.ScaleStatefulSet(ctx, namespace, name, replicas); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "scaled successfully"})
}

func (h *WorkloadHandler) HandleDeleteStatefulSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteStatefulSet(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *WorkloadHandler) HandleUpdateStatefulSet(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateStatefulSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.UpdateStatefulSet(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *WorkloadHandler) HandleRestartStatefulSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.RestartStatefulSet(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "restarted successfully"})
}

func (h *WorkloadHandler) HandleListDaemonSets(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.ListDaemonSets(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleGetDaemonSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.GetDaemonSet(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleDeleteDaemonSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteDaemonSet(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *WorkloadHandler) HandleRestartDaemonSet(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.RestartDaemonSet(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "restarted successfully"})
}

func (h *WorkloadHandler) HandleUpdateDaemonSet(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateDaemonSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.UpdateDaemonSet(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *WorkloadHandler) HandleListCronJobs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.ListCronJobs(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleDeleteCronJob(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteCronJob(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *WorkloadHandler) HandleSuspendCronJob(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	suspend := r.URL.Query().Get("suspend") == "true"

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.SuspendCronJob(ctx, namespace, name, suspend); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *WorkloadHandler) HandleListJobs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.ListJobs(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	result, err := svc.GetJob(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *WorkloadHandler) HandleGetCronJob(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	cj, err := svc.GetCronJob(ctx, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: cj})
}

func (h *WorkloadHandler) HandleUpdateCronJob(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateCronJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.UpdateCronJob(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *WorkloadHandler) HandleDeleteJob(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}

	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.DeleteJob(ctx, namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *WorkloadHandler) HandleUpdateJob(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "namespace and name are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.UpdateJob(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *WorkloadHandler) HandleCreateDeployment(w http.ResponseWriter, r *http.Request) {
	var req service.CreateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" || req.Image == "" {
		writeError(w, http.StatusBadRequest, "namespace, name and image are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.CreateDeployment(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *WorkloadHandler) HandleCreateStatefulSet(w http.ResponseWriter, r *http.Request) {
	var req service.CreateStatefulSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" || req.Image == "" {
		writeError(w, http.StatusBadRequest, "namespace, name and image are required")
		return
	}
	if req.ServiceName == "" {
		req.ServiceName = req.Name
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.CreateStatefulSet(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *WorkloadHandler) HandleCreateDaemonSet(w http.ResponseWriter, r *http.Request) {
	var req service.CreateDaemonSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" || req.Image == "" {
		writeError(w, http.StatusBadRequest, "namespace, name and image are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.CreateDaemonSet(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *WorkloadHandler) HandleCreateCronJob(w http.ResponseWriter, r *http.Request) {
	var req service.CreateCronJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" || req.Image == "" || req.Schedule == "" {
		writeError(w, http.StatusBadRequest, "namespace, name, image and schedule are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.CreateCronJob(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *WorkloadHandler) HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	var req service.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Name == "" || req.Image == "" {
		writeError(w, http.StatusBadRequest, "namespace, name and image are required")
		return
	}
	ctx := context.Background()
	svc := service.NewWorkloadService(h.clientGetter.GetK8sClient())
	if err := svc.CreateJob(ctx, req); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

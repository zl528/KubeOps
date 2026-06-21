package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type ExecHandler struct {
	clientGetter ClientGetter
}

func NewExecHandler(clientGetter ClientGetter) *ExecHandler {
	return &ExecHandler{clientGetter: clientGetter}
}

func (h *ExecHandler) HandleExecCommand(w http.ResponseWriter, r *http.Request) {
	var req service.PodExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}
	if len(req.Command) == 0 {
		req.Command = []string{"/bin/sh"}
	}

	ctx := context.Background()
	svc := service.NewExecService(h.clientGetter.GetK8sClient(), h.clientGetter.GetRestConfig())
	result, err := svc.ExecCommand(ctx, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *ExecHandler) HandlePortForward(w http.ResponseWriter, r *http.Request) {
	var req service.PortForwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Namespace == "" || req.Pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}
	if req.LocalPort == 0 || req.RemotePort == 0 {
		writeError(w, http.StatusBadRequest, "localPort and remotePort are required")
		return
	}

	ctx := context.Background()
	svc := service.NewExecService(h.clientGetter.GetK8sClient(), h.clientGetter.GetRestConfig())
	result, err := svc.PortForward(ctx, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *ExecHandler) HandleStreamPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	ctx := context.Background()
	req := service.PodLogStreamRequest{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
	}

	svc := service.NewExecService(h.clientGetter.GetK8sClient(), h.clientGetter.GetRestConfig())
	if err := svc.StreamPodLogs(ctx, req, w); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

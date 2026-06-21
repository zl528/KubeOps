package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type LogHandler struct {
	clientGetter ClientGetter
}

func NewLogHandler(clientGetter ClientGetter) *LogHandler {
	return &LogHandler{clientGetter: clientGetter}
}

func (h *LogHandler) HandleGetPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")
	search := r.URL.Query().Get("search")
	level := r.URL.Query().Get("level")
	sinceTime := r.URL.Query().Get("sinceTime")

	tailLines := int64(1000)
	if tl := r.URL.Query().Get("tailLines"); tl != "" {
		if v, err := strconv.ParseInt(tl, 10, 64); err == nil {
			tailLines = v
		}
	}

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		TailLines: tailLines,
		Search:    search,
		Level:     level,
		SinceTime: sinceTime,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	result, err := svc.GetPodLogs(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *LogHandler) HandleStreamPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		Follow:    true,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "response writer does not support flushing")
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	if err := svc.StreamPodLogs(ctx, query, w); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	flusher.Flush()
}

func (h *LogHandler) HandleDownloadPodLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	buf, err := svc.DownloadPodLogs(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := pod + ".log"
	if container != "" {
		filename = pod + "-" + container + ".log"
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

func (h *LogHandler) HandleListContainers(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	containers, err := svc.ListPodContainers(ctx, namespace, pod)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: containers})
}

func (h *LogHandler) HandleSearchLogs(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")
	keyword := r.URL.Query().Get("keyword")

	tailLines := int64(1000)
	if tl := r.URL.Query().Get("tailLines"); tl != "" {
		if v, err := strconv.ParseInt(tl, 10, 64); err == nil {
			tailLines = v
		}
	}

	if namespace == "" || pod == "" || keyword == "" {
		writeError(w, http.StatusBadRequest, "namespace, pod and keyword are required")
		return
	}

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		TailLines: tailLines,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	result, err := svc.SearchLogs(ctx, query, keyword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *LogHandler) HandleGetLogsByLevel(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")
	level := r.URL.Query().Get("level")

	tailLines := int64(1000)
	if tl := r.URL.Query().Get("tailLines"); tl != "" {
		if v, err := strconv.ParseInt(tl, 10, 64); err == nil {
			tailLines = v
		}
	}

	if namespace == "" || pod == "" || level == "" {
		writeError(w, http.StatusBadRequest, "namespace, pod and level are required")
		return
	}

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		TailLines: tailLines,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	result, err := svc.GetLogsByLevel(ctx, query, level)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *LogHandler) HandleGetPreviousLogs(w http.ResponseWriter, r *http.Request) {
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

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		TailLines: tailLines,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	result, err := svc.GetPreviousPodLogs(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *LogHandler) HandleGetLogsFromTime(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	pod := r.URL.Query().Get("pod")
	container := r.URL.Query().Get("container")

	if namespace == "" || pod == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	query := service.LogQuery{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
	}

	ctx := context.Background()
	svc := service.NewLogService(h.clientGetter.GetK8sClient())
	result, err := svc.GetPodLogsFromTime(ctx, query, time.Now().Add(-1*time.Hour))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

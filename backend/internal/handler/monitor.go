package handler

import (
	"context"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type MonitorHandler struct {
	clientGetter ClientGetter
}

func NewMonitorHandler(clientGetter ClientGetter) *MonitorHandler {
	return &MonitorHandler{clientGetter: clientGetter}
}

func (h *MonitorHandler) HandleGetClusterMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetClusterMetrics(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *MonitorHandler) HandleGetNodeMetrics(w http.ResponseWriter, r *http.Request) {
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetNodeMetrics(ctx, nodeName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *MonitorHandler) HandleGetPodMetrics(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")
	if namespace == "" || podName == "" {
		writeError(w, http.StatusBadRequest, "namespace and pod are required")
		return
	}

	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetPodMetrics(ctx, namespace, podName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *MonitorHandler) HandleDetectCapabilities(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	caps, err := svc.DetectCapabilities(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: caps})
}

func (h *MonitorHandler) HandleGetPrometheusStatus(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	status, err := svc.GetPrometheusStatus(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: status})
}

func (h *MonitorHandler) HandleGetPrometheusTargets(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	targets, err := svc.GetPrometheusTargets(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: targets})
}

func (h *MonitorHandler) HandleQueryPrometheus(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.QueryPrometheus(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *MonitorHandler) HandleGetMetricNames(w http.ResponseWriter, r *http.Request) {
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.QueryPrometheus("{__name__=~\".+\",job=~\".+\"}")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *MonitorHandler) HandleGetLabelValues(w http.ResponseWriter, r *http.Request) {
	metricName := r.URL.Query().Get("metric")
	labelName := r.URL.Query().Get("label")
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")

	query := ""
	if metricName != "" && labelName != "" {
		query = "group by (" + labelName + ") (" + metricName + "{" + labelName + "=~\".+\"})"
	} else if labelName != "" {
		query = "group by (" + labelName + ") ({__name__=~\".+\", " + labelName + "=~\".+\"})"
	} else {
		writeError(w, http.StatusBadRequest, "label is required")
		return
	}
	
	result, err := svc.QueryPrometheus(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *MonitorHandler) HandleGetCPUUsage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetCPUUsage(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: map[string]float64{"usage": result}})
}

func (h *MonitorHandler) HandleGetMemoryUsage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetMemoryUsage(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: map[string]float64{"usage": result}})
}

func (h *MonitorHandler) HandleGetDiskUsage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetDiskUsage(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: map[string]float64{"usage": result}})
}

func (h *MonitorHandler) HandleGetNetworkUsage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	rxBytes, txBytes, err := svc.GetNetworkUsage(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: map[string]int64{"rxBytes": rxBytes, "txBytes": txBytes}})
}

func (h *MonitorHandler) HandleGetGrafanaDashboards(w http.ResponseWriter, r *http.Request) {
	grafanaURL := r.URL.Query().Get("grafanaURL")
	svc := service.NewMonitorService(h.clientGetter.GetK8sClient(), "")
	result, err := svc.GetGrafanaDashboards(grafanaURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

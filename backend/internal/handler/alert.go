package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type AlertHandler struct {
	alertSvc *service.AlertService
}

func NewAlertHandler(alertSvc *service.AlertService) *AlertHandler {
	return &AlertHandler{alertSvc: alertSvc}
}

func (h *AlertHandler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	result, err := h.alertSvc.ListRules(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *AlertHandler) HandleGetRule(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	result, err := h.alertSvc.GetRule(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *AlertHandler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	var rule service.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.CreateRule(ctx, rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *AlertHandler) HandleUpdateRule(w http.ResponseWriter, r *http.Request) {
	var rule service.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if rule.ID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.UpdateRule(ctx, rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *AlertHandler) HandleDeleteRule(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.DeleteRule(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *AlertHandler) HandleEnableRule(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	enabled := r.URL.Query().Get("enabled") == "true"
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.EnableRule(ctx, id, enabled); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "updated successfully"})
}

func (h *AlertHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")
	ctx := context.Background()
	result, err := h.alertSvc.ListAlerts(ctx, severity)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *AlertHandler) HandleGetAlert(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	result, err := h.alertSvc.GetAlert(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *AlertHandler) HandleResolveAlert(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.ResolveAlert(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "resolved successfully"})
}

func (h *AlertHandler) HandleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.DeleteAlert(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *AlertHandler) HandleGetAlertStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	result, err := h.alertSvc.GetAlertStats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *AlertHandler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	result, err := h.alertSvc.ListNotifications(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: result})
}

func (h *AlertHandler) HandleCreateNotification(w http.ResponseWriter, r *http.Request) {
	var config service.NotificationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.CreateNotification(ctx, config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "created successfully"})
}

func (h *AlertHandler) HandleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	ctx := context.Background()
	if err := h.alertSvc.DeleteNotification(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

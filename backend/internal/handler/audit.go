package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type AuditHandler struct {
	auditSvc *service.AuditService
}

func NewAuditHandler(auditSvc *service.AuditService) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc}
}

func (h *AuditHandler) HandleQueryLogs(w http.ResponseWriter, r *http.Request) {
	var query service.AuditLogQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	ctx := context.Background()
	logs, total := h.auditSvc.Query(ctx, query)

	writeJSON(w, http.StatusOK, model.Response{
		Code:    0,
		Message: "ok",
		Data: map[string]interface{}{
			"logs":  logs,
			"total": total,
			"page":  query.Page,
			"pageSize": query.PageSize,
		},
	})
}

func (h *AuditHandler) HandleGetLog(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ctx := context.Background()
	log, err := h.auditSvc.GetByID(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: log})
}

func (h *AuditHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	stats := h.auditSvc.GetStats(ctx)
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: stats})
}

func (h *AuditHandler) HandleExportLogs(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	user := r.URL.Query().Get("user")
	action := r.URL.Query().Get("action")
	resource := r.URL.Query().Get("resource")
	namespace := r.URL.Query().Get("namespace")

	query := service.AuditLogQuery{
		User:      user,
		Action:    action,
		Resource:  resource,
		Namespace: namespace,
		PageSize:  10000,
	}

	ctx := context.Background()
	data, contentType, err := h.auditSvc.Export(ctx, query, format)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := "audit-logs." + format
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(data)
}

func (h *AuditHandler) HandleCleanupLogs(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}

	ctx := context.Background()
	removed, err := h.auditSvc.Cleanup(ctx, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{
		Code:    0,
		Message: "cleaned up successfully",
		Data:    map[string]int{"removed": removed},
	})
}

func (h *AuditHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	stats := h.auditSvc.GetStats(ctx)

	users := make([]string, 0, len(stats.ByUser))
	for user := range stats.ByUser {
		if user != "" {
			users = append(users, user)
		}
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: users})
}

func (h *AuditHandler) HandleListActions(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	stats := h.auditSvc.GetStats(ctx)

	actions := make([]string, 0, len(stats.ByAction))
	for action := range stats.ByAction {
		if action != "" {
			actions = append(actions, action)
		}
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: actions})
}

package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type BackupHandler struct {
	backupSvc *service.BackupService
}

func NewBackupHandler(backupSvc *service.BackupService) *BackupHandler {
	return &BackupHandler{backupSvc: backupSvc}
}

func (h *BackupHandler) HandleCreateBackup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Namespace   string   `json:"namespace"`
		Description string   `json:"description"`
		Resources   []string `json:"resources"`
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
	record, err := h.backupSvc.CreateBackup(ctx, req.Name, req.Namespace, req.Description, req.Resources)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "backup started", Data: record})
}

func (h *BackupHandler) HandleListBackups(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	backups := h.backupSvc.ListBackups(ctx)
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: backups})
}

func (h *BackupHandler) HandleGetBackup(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ctx := context.Background()
	backup, err := h.backupSvc.GetBackup(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: backup})
}

func (h *BackupHandler) HandleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ctx := context.Background()
	if err := h.backupSvc.DeleteBackup(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "deleted successfully"})
}

func (h *BackupHandler) HandleGetBackupContent(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ctx := context.Background()
	content, err := h.backupSvc.GetBackupContent(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: content})
}

func (h *BackupHandler) HandleRestore(w http.ResponseWriter, r *http.Request) {
	var req service.RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.BackupID == "" {
		writeError(w, http.StatusBadRequest, "backupId is required")
		return
	}

	ctx := context.Background()
	result, err := h.backupSvc.Restore(ctx, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "restore completed", Data: result})
}

func (h *BackupHandler) HandleExportBackup(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	ctx := context.Background()
	data, err := h.backupSvc.ExportBackup(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=backup-"+id+".json")
	w.Write(data)
}

func (h *BackupHandler) HandleImportBackup(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "imported-backup"
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	ctx := context.Background()
	record, err := h.backupSvc.ImportBackup(ctx, body, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "imported successfully", Data: record})
}

func (h *BackupHandler) HandleListResources(w http.ResponseWriter, r *http.Request) {
	resources := []string{
		"deployments", "statefulsets", "daemonsets",
		"services", "configmaps", "secrets",
		"ingresses", "roles", "rolebindings",
		"serviceaccounts", "persistentvolumeclaims",
		"cronjobs", "jobs", "horizontalpodautoscalers",
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: resources})
}

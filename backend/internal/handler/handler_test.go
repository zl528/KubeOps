package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubeops/ops-kubernetes/internal/model"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := model.Response{Code: 0, Message: "ok"}

	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}

	var resp model.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Errorf("expected message ok, got %s", resp.Message)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "invalid request")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp model.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected code 400, got %d", resp.Code)
	}
	if resp.Message != "invalid request" {
		t.Errorf("expected message invalid request, got %s", resp.Message)
	}
}

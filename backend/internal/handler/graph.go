package handler

import (
	"context"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type GraphHandler struct {
	clientGetter ClientGetter
}

func NewGraphHandler(clientGetter ClientGetter) *GraphHandler {
	return &GraphHandler{clientGetter: clientGetter}
}

func (h *GraphHandler) HandleGetGraph(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ctx := context.Background()
	svc := service.NewGraphService(h.clientGetter.GetK8sClient())
	graph, err := svc.GetGraph(ctx, namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: graph})
}

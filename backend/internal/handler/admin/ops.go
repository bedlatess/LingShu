package admin

import (
	"net/http"

	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type OpsHandler struct {
	ops    service.OpsService
	alerts service.OpsAlertService
}

func NewOpsHandler(ops service.OpsService, alerts ...service.OpsAlertService) OpsHandler {
	handler := OpsHandler{ops: ops}
	if len(alerts) > 0 {
		handler.alerts = alerts[0]
	}
	return handler
}

func (h OpsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	item, err := h.ops.Dashboard(r.Context(), h.alerts)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

package admin

import (
	"net/http"

	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type OpsHandler struct {
	ops service.OpsService
}

func NewOpsHandler(ops service.OpsService) OpsHandler {
	return OpsHandler{ops: ops}
}

func (h OpsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	item, err := h.ops.Dashboard(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

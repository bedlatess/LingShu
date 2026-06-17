package admin

import (
	"net/http"

	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type ReportHandler struct {
	reports service.ReportService
}

func NewReportHandler(reports service.ReportService) ReportHandler {
	return ReportHandler{reports: reports}
}

func (h ReportHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	item, err := h.reports.AdminDashboard(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h ReportHandler) Logs(w http.ResponseWriter, r *http.Request) {
	items, err := h.reports.AdminLogs(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h ReportHandler) Ledger(w http.ResponseWriter, r *http.Request) {
	items, err := h.reports.AdminLedger(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

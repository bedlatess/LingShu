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
	page, limit := parsePagination(r)
	items, total, err := h.reports.AdminLogsPaged(r.Context(), page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h ReportHandler) Ledger(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	items, total, err := h.reports.AdminLedgerPaged(r.Context(), page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h ReportHandler) Daily(w http.ResponseWriter, r *http.Request) {
	items, err := h.reports.ReportDaily(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h ReportHandler) ByUser(w http.ResponseWriter, r *http.Request) {
	items, err := h.reports.ReportByUser(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h ReportHandler) ByModel(w http.ResponseWriter, r *http.Request) {
	items, err := h.reports.ReportByModel(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h ReportHandler) ByChannel(w http.ResponseWriter, r *http.Request) {
	items, err := h.reports.ReportByChannel(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

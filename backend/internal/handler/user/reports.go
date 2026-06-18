package user

import (
	"net/http"
	"strconv"

	"lingshu/backend/internal/dto"
	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type ReportHandler struct {
	reports service.ReportService
}

func NewReportHandler(reports service.ReportService) ReportHandler {
	return ReportHandler{reports: reports}
}

func (h ReportHandler) Logs(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	items, err := h.reports.UserLogs(r.Context(), current.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": dto.NewUserGatewayLogDTOs(items)})
}

func (h ReportHandler) Ledger(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	items, err := h.reports.UserLedger(r.Context(), current.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": dto.NewUserLedgerRecordDTOs(items)})
}

func (h ReportHandler) Daily(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	items, err := h.reports.DailyStats(r.Context(), current.ID, days)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": dto.NewUserDailyStatDTOs(items)})
}

func (h ReportHandler) Models(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	items, err := h.reports.ModelStats(r.Context(), current.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": dto.NewUserModelStatDTOs(items)})
}

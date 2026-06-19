package user

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"lingshu/backend/internal/dto"
	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
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

func (h ReportHandler) ExportUsageCSV(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	writer := userCSVResponse(w, "usage.csv")
	if err := writer.Write([]string{"request_id", "model_id", "status", "http_status", "total_tokens", "charge", "created_at"}); err != nil {
		return
	}
	err := h.reports.ExportUserLogs(r.Context(), current.ID, func(item repository.GatewayLog) error {
		return writer.Write([]string{
			item.RequestID,
			item.ModelID,
			item.Status,
			strconv.Itoa(item.HTTPStatus),
			strconv.Itoa(item.TotalTokens),
			item.Charge,
			item.CreatedAt.Format(time.RFC3339),
		})
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writer.Flush()
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

func userCSVResponse(w http.ResponseWriter, filename string) *csv.Writer {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	return csv.NewWriter(w)
}

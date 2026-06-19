package admin

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

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

func (h ReportHandler) ExportUsageCSV(w http.ResponseWriter, r *http.Request) {
	writer := csvResponse(w, "admin-usage.csv")
	if err := writer.Write([]string{"request_id", "user_id", "model_id", "status", "http_status", "total_tokens", "base_cost", "charge", "created_at"}); err != nil {
		return
	}
	err := h.reports.ExportAdminLogs(r.Context(), func(item repository.GatewayLog) error {
		return writer.Write([]string{item.RequestID, item.UserID, item.ModelID, item.Status, intString(item.HTTPStatus), intString(item.TotalTokens), item.BaseCost, item.Charge, item.CreatedAt.Format(timeFormatCSV)})
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writer.Flush()
}

func (h ReportHandler) ExportLedgerCSV(w http.ResponseWriter, r *http.Request) {
	writer := csvResponse(w, "admin-ledger.csv")
	if err := writer.Write([]string{"user_id", "type", "amount", "balance_before", "balance_after", "base_cost", "rate_multiplier", "remark", "created_at"}); err != nil {
		return
	}
	err := h.reports.ExportAdminLedger(r.Context(), func(item repository.LedgerRecord) error {
		return writer.Write([]string{item.UserID, item.Type, item.Amount, item.BalanceBefore, item.BalanceAfter, item.BaseCost, item.RateMultiplier, item.Remark, item.CreatedAt.Format(timeFormatCSV)})
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writer.Flush()
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

const timeFormatCSV = time.RFC3339

func csvResponse(w http.ResponseWriter, filename string) *csv.Writer {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	return csv.NewWriter(w)
}

func intString(value int) string {
	return strconv.Itoa(value)
}

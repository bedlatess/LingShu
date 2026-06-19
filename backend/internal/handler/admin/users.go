package admin

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type UserHandler struct {
	users   service.AdminUserService
	keys    service.APIKeyService
	reports service.ReportService
}

func NewUserHandler(users service.AdminUserService, keys service.APIKeyService, reports service.ReportService) UserHandler {
	return UserHandler{users: users, keys: keys, reports: reports}
}

func (h UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	users, total, err := h.users.ListPaged(r.Context(), page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, users, total, page, limit)
}

func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.CreateUserInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, err := h.users.Create(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, user)
}

func (h UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.UpdateUserInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, err := h.users.Update(r.Context(), current.ID, chi.URLParam(r, "id"), input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var req struct {
		Password string `json:"password"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.users.ResetPassword(r.Context(), current.ID, chi.URLParam(r, "id"), req.Password, clientIP(r), r.UserAgent()); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h UserHandler) Ban(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	user, err := h.users.Ban(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) Unban(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	user, err := h.users.Unban(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) UpdateLimits(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.UpdateUserLimitsInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, err := h.users.UpdateLimits(r.Context(), current.ID, chi.URLParam(r, "id"), input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) RevokeTokens(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	user, err := h.users.RevokeTokens(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) AdjustBalance(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.AdjustBalanceInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, err := h.users.AdjustBalance(r.Context(), current.ID, chi.URLParam(r, "id"), input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h UserHandler) UserLogs(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	query := r.URL.Query()
	filter := repository.UserLogFilter{
		Status: strings.TrimSpace(query.Get("status")),
		Model:  strings.TrimSpace(query.Get("model")),
		From:   strings.TrimSpace(query.Get("from")),
		To:     strings.TrimSpace(query.Get("to")),
	}
	items, total, err := h.reports.UserLogsFilteredPaged(r.Context(), chi.URLParam(r, "id"), filter, page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h UserHandler) ExportUserUsageCSV(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	writer := csvResponse(w, "admin-user-usage.csv")
	if err := writer.Write([]string{"request_id", "user_id", "model_id", "status", "http_status", "total_tokens", "base_cost", "charge", "created_at"}); err != nil {
		return
	}
	err := h.reports.ExportUserLogs(r.Context(), userID, func(item repository.GatewayLog) error {
		return writer.Write([]string{item.RequestID, item.UserID, item.ModelID, item.Status, intString(item.HTTPStatus), intString(item.TotalTokens), item.BaseCost, item.Charge, item.CreatedAt.Format(timeFormatCSV)})
	})
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writer.Flush()
}

func (h UserHandler) UserLedger(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	query := r.URL.Query()
	filter := repository.UserLedgerFilter{
		Type: strings.TrimSpace(query.Get("type")),
		From: strings.TrimSpace(query.Get("from")),
		To:   strings.TrimSpace(query.Get("to")),
	}
	items, total, err := h.reports.UserLedgerFilteredPaged(r.Context(), chi.URLParam(r, "id"), filter, page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h UserHandler) UserAPIKeys(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	items, total, err := h.keys.ListForUserPaged(r.Context(), chi.URLParam(r, "id"), page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h UserHandler) UserSummary(w http.ResponseWriter, r *http.Request) {
	item, err := h.reports.UserSummary(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h UserHandler) AuditCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.users.AuditCount(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]int64{"count": count})
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

package admin

import (
	"net/http"
	"strings"
	"time"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type SettingsHandler struct {
	settings service.SettingsService
	audits   repository.AuditRepository
}

func NewSettingsHandler(settings service.SettingsService, audits repository.AuditRepository) SettingsHandler {
	return SettingsHandler{settings: settings, audits: audits}
}

func (h SettingsHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	items, total, err := h.settings.ListPaged(r.Context(), page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h SettingsHandler) Patch(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var req struct {
		Items []repository.SettingUpdate `json:"items"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	items, err := h.settings.Patch(r.Context(), current.ID, req.Items, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h SettingsHandler) AuditLogs(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	filter, err := parseAuditLogFilter(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	items, total, err := h.settings.AuditLogsPaged(r.Context(), filter, page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h SettingsHandler) CleanupAuditLogs(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BeforeDays int `json:"before_days"`
	}
	_ = httpx.Decode(r, &input)
	if input.BeforeDays < 7 {
		input.BeforeDays = 90
	}
	deleted, err := h.audits.DeleteOlderThan(r.Context(), input.BeforeDays)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"deleted": deleted})
}

func parseAuditLogFilter(r *http.Request) (repository.AuditLogFilter, error) {
	query := r.URL.Query()
	filter := repository.AuditLogFilter{
		ActorID:    strings.TrimSpace(query.Get("actor_id")),
		Action:     strings.TrimSpace(query.Get("action")),
		TargetType: strings.TrimSpace(query.Get("target_type")),
	}

	from, err := parseAuditTime(query.Get("from"))
	if err != nil {
		return repository.AuditLogFilter{}, err
	}
	to, err := parseAuditTime(query.Get("to"))
	if err != nil {
		return repository.AuditLogFilter{}, err
	}
	filter.From = from
	if to != nil && len(strings.TrimSpace(query.Get("to"))) == len("2006-01-02") {
		adjusted := to.Add(24*time.Hour - time.Nanosecond)
		filter.To = &adjusted
	} else {
		filter.To = to
	}
	return filter, nil
}

func parseAuditTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		if parsed, errRFC3339 := time.Parse(time.RFC3339, value); errRFC3339 == nil {
			return &parsed, nil
		}
		return nil, err
	}
	return &parsed, nil
}

package admin

import (
	"net/http"

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
	items, err := h.settings.List(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
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
	items, err := h.audits.List(r.Context(), 100)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

package user

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"lingshu/backend/internal/dto"
	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type Handler struct {
	announcements service.AnnouncementService
	redeems       service.RedeemService
	keys          service.APIKeyService
	portal        service.UserPortalService
}

func New(announcements service.AnnouncementService, redeems service.RedeemService, keys service.APIKeyService, portal service.UserPortalService) Handler {
	return Handler{announcements: announcements, redeems: redeems, keys: keys, portal: portal}
}

func (h Handler) Announcements(w http.ResponseWriter, r *http.Request) {
	items, err := h.announcements.ListOnline(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) Redeem(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.redeems.Redeem(r.Context(), current.ID, req.Code, clientIP(r))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	item, err := h.portal.Dashboard(r.Context(), current.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, dto.NewUserDashboardDTO(item))
}

func (h Handler) Models(w http.ResponseWriter, r *http.Request) {
	items, err := h.portal.Models(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": dto.NewUserModelConfigDTOs(items)})
}

func (h Handler) APIKeys(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.keys.ListForUser(r.Context(), current.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.keys.CreateForUser(r.Context(), current.ID, req.Name)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

func (h Handler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.keys.UpdateForUser(r.Context(), current.ID, chi.URLParam(r, "id"), req.Name, req.Status)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h Handler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.keys.DeleteForUser(r.Context(), current.ID, chi.URLParam(r, "id")); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.Error(w, http.StatusNotFound, "密钥不存在或已被删除")
			return
		}
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

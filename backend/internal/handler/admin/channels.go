package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type ChannelHandler struct {
	channels service.ChannelService
}

func NewChannelHandler(channels service.ChannelService) ChannelHandler {
	return ChannelHandler{channels: channels}
}

func (h ChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.channels.List(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h ChannelHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.ChannelInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.channels.Create(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

func (h ChannelHandler) Update(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.ChannelInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.channels.Update(r.Context(), current.ID, chi.URLParam(r, "id"), input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h ChannelHandler) Disable(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	if err := h.channels.Disable(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent()); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h ChannelHandler) Test(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BaseURL string `json:"base_url"`
	}
	_ = httpx.Decode(r, &req)
	result, err := h.channels.Test(r.Context(), chi.URLParam(r, "id"), req.BaseURL)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, result)
}

func (h ChannelHandler) BindModel(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.BindChannelModelInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.channels.BindModel(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

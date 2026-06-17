package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type AnnouncementHandler struct {
	announcements service.AnnouncementService
}

func NewAnnouncementHandler(announcements service.AnnouncementService) AnnouncementHandler {
	return AnnouncementHandler{announcements: announcements}
}

func (h AnnouncementHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.announcements.ListAdmin(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h AnnouncementHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.AnnouncementInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.announcements.Create(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

func (h AnnouncementHandler) Update(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.AnnouncementInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.announcements.Update(r.Context(), current.ID, chi.URLParam(r, "id"), input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h AnnouncementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	if err := h.announcements.Delete(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent()); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

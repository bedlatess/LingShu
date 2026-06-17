package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type APIKeyHandler struct {
	keys service.APIKeyService
}

func NewAPIKeyHandler(keys service.APIKeyService) APIKeyHandler {
	return APIKeyHandler{keys: keys}
}

func (h APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.keys.ListAll(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.CreateAPIKeyInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	created, err := h.keys.Create(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h APIKeyHandler) Disable(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	if err := h.keys.Disable(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent()); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

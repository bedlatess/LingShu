package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type ModelHandler struct {
	models service.ModelService
}

func NewModelHandler(models service.ModelService) ModelHandler {
	return ModelHandler{models: models}
}

func (h ModelHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.models.List(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h ModelHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.ModelInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.models.Create(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

func (h ModelHandler) Update(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input repository.ModelInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.models.Update(r.Context(), current.ID, chi.URLParam(r, "id"), input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

func (h ModelHandler) Disable(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	if err := h.models.Disable(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent()); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type RedeemHandler struct {
	redeems service.RedeemService
}

func NewRedeemHandler(redeems service.RedeemService) RedeemHandler {
	return RedeemHandler{redeems: redeems}
}

func (h RedeemHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	items, total, err := h.redeems.ListPaged(r.Context(), page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h RedeemHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.CreateRedeemInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	items, err := h.redeems.Create(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"items": items})
}

func (h RedeemHandler) Disable(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	if err := h.redeems.Disable(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent()); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h RedeemHandler) Records(w http.ResponseWriter, r *http.Request) {
	items, err := h.redeems.Records(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

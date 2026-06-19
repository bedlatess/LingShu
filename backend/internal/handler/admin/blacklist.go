package admin

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type BlacklistHandler struct {
	blacklist service.AccessBlacklistService
}

func NewBlacklistHandler(blacklist service.AccessBlacklistService) BlacklistHandler {
	return BlacklistHandler{blacklist: blacklist}
}

func (h BlacklistHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	query := r.URL.Query()
	filter := repository.AccessBlacklistFilter{
		Kind:   strings.TrimSpace(query.Get("kind")),
		Scope:  strings.TrimSpace(query.Get("scope")),
		Active: strings.TrimSpace(query.Get("active")),
		Query:  strings.TrimSpace(query.Get("q")),
	}
	items, total, err := h.blacklist.ListPaged(r.Context(), filter, page, limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	writePagedJSON(w, items, total, page, limit)
}

func (h BlacklistHandler) Create(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	var input service.CreateAccessBlacklistRequest
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := h.blacklist.CreateManual(r.Context(), current.ID, input, clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

func (h BlacklistHandler) Release(w http.ResponseWriter, r *http.Request) {
	current, _ := middleware.CurrentUser(r.Context())
	item, err := h.blacklist.Release(r.Context(), current.ID, chi.URLParam(r, "id"), clientIP(r), r.UserAgent())
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

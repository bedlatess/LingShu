package admin

import (
	"net/http"
	"strconv"

	"lingshu/backend/internal/job"
	"lingshu/backend/internal/pkg/httpx"
)

type CleanupHandler struct {
	cleaner job.Cleaner
}

func NewCleanupHandler(cleaner job.Cleaner) CleanupHandler {
	return CleanupHandler{cleaner: cleaner}
}

func (h CleanupHandler) Run(w http.ResponseWriter, r *http.Request) {
	results := h.cleaner.Run(r.Context())
	if err := h.cleaner.SaveHistory(r.Context(), results); err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": results})
}

func (h CleanupHandler) History(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.cleaner.History(r.Context(), limit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

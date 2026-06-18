package admin

import (
	"net/http"
	"strconv"

	"lingshu/backend/internal/pkg/httpx"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

func writePagedJSON[T any](w http.ResponseWriter, items []T, total, page, limit int) {
	httpx.JSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	return page, limit
}

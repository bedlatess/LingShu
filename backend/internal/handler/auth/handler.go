package auth

import (
	"errors"
	"net/http"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type Handler struct {
	auth service.AuthService
}

func New(authService service.AuthService) Handler {
	return Handler{auth: authService}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, service.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		httpx.Error(w, status, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, result)
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterInput
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, err := h.auth.Register(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrForbidden) {
			status = http.StatusForbidden
		}
		httpx.Error(w, status, err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, user)
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.auth.Me(r.Context(), current.ID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "user not found")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	current, ok := middleware.CurrentUser(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req changePasswordRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.auth.ChangePassword(r.Context(), current.ID, req.OldPassword, req.NewPassword); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

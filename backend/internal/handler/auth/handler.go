package auth

import (
	"errors"
	"net/http"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type Handler struct {
	auth      service.AuthService
	blacklist *service.AccessBlacklistService
}

func New(authService service.AuthService, blacklist ...service.AccessBlacklistService) Handler {
	handler := Handler{auth: authService}
	if len(blacklist) > 0 {
		handler.blacklist = &blacklist[0]
	}
	return handler
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Captcha  string `json:"captcha_token"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type sendCodeRequest struct {
	Purpose string `json:"purpose"`
	Email   string `json:"email"`
	Captcha string `json:"captcha_token"`
}

type forgotRequest struct {
	Email   string `json:"email"`
	Captcha string `json:"captcha_token"`
}

type resetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	ip := clientIP(r)
	deviceID := ""
	if h.blacklist != nil {
		deviceID = httpx.VerifiedDeviceID(r, h.blacklist.DeviceSecret(r.Context()))
	}
	result, err := h.auth.Login(r.Context(), req.Login, req.Password, ip, req.Captcha)
	if err != nil {
		if h.blacklist != nil {
			h.blacklist.RecordLoginFailure(r.Context(), ip, deviceID)
		}
		status := http.StatusUnauthorized
		if errors.Is(err, service.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		if errors.Is(err, service.ErrLoginLocked) {
			status = http.StatusTooManyRequests
		}
		httpx.Error(w, status, err.Error())
		return
	}
	if h.blacklist != nil {
		h.blacklist.RecordLoginSuccess(r.Context(), ip, deviceID)
	}
	httpx.JSON(w, http.StatusOK, result)
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterInput
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.RemoteIP = clientIP(r)
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

func (h Handler) SendEmailCode(w http.ResponseWriter, r *http.Request) {
	var req sendCodeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.auth.SendEmailCode(r.Context(), req.Purpose, req.Email, req.Captcha, clientIP(r)); err != nil {
		writeAuthError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) Forgot(w http.ResponseWriter, r *http.Request) {
	var req forgotRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.auth.ForgotPassword(r.Context(), req.Email, req.Captcha, clientIP(r)); err != nil {
		writeAuthError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h Handler) Reset(w http.ResponseWriter, r *http.Request) {
	var req resetRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.auth.ResetPassword(r.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		writeAuthError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		httpx.Error(w, http.StatusForbidden, "registration is closed")
	case errors.Is(err, service.ErrEmailNotConfigured):
		httpx.Error(w, http.StatusServiceUnavailable, "smtp is not configured")
	case errors.Is(err, service.ErrEmailCodeCooldown):
		httpx.Error(w, http.StatusTooManyRequests, "email code was sent recently")
	case errors.Is(err, service.ErrInvalidEmailCode):
		httpx.Error(w, http.StatusBadRequest, "invalid email code")
	case errors.Is(err, service.ErrInvalidCredentials):
		httpx.Error(w, http.StatusBadRequest, "invalid credentials")
	case errors.Is(err, service.ErrCaptchaRequired):
		httpx.Error(w, http.StatusBadRequest, "captcha token required")
	case errors.Is(err, service.ErrCaptchaNotConfigured):
		httpx.Error(w, http.StatusServiceUnavailable, "captcha is not configured")
	case errors.Is(err, service.ErrInvalidCaptcha):
		httpx.Error(w, http.StatusBadRequest, "invalid captcha token")
	default:
		httpx.Error(w, http.StatusBadRequest, err.Error())
	}
}

func clientIP(r *http.Request) string {
	return httpx.ClientIP(r, httpx.SettingsFromContext(r.Context()))
}

package gateway

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type Handler struct {
	gateway service.GatewayService
}

func New(gateway service.GatewayService) Handler {
	return Handler{gateway: gateway}
}

func (h Handler) Models(w http.ResponseWriter, r *http.Request) {
	models, err := h.gateway.Models(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	data := make([]map[string]any, 0, len(models))
	for _, model := range models {
		data = append(data, map[string]any{"id": model.PublicName, "object": "model"})
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"object": "list", "data": data})
}

func (h Handler) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentGatewayPrincipal(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	principalDTO := service.GatewayPrincipal{
		UserID:           principal.UserID,
		APIKeyID:         principal.APIKeyID,
		Balance:          principal.Balance,
		RPMLimit:         principal.RPMLimit,
		ConcurrencyLimit: principal.ConcurrencyLimit,
	}
	var preview struct {
		Stream bool `json:"stream"`
	}
	_ = json.Unmarshal(body, &preview)
	if preview.Stream {
		start := time.Now()
		model, channel, estimate, resp, err := h.gateway.OpenChatStream(r.Context(), principalDTO, body, sessionID(r))
		if err != nil {
			httpx.Error(w, statusForGatewayError(err), err.Error())
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/event-stream")
		}
		w.WriteHeader(resp.StatusCode)
		captured, _ := service.CopyAndCapture(w, resp.Body)
		h.gateway.FinalizeStream(r.Context(), principalDTO, model, channel, body, captured, estimate, resp.StatusCode, clientIP(r), start)
		return
	}
	status, responseBody, err := h.gateway.Chat(r.Context(), principalDTO, body, clientIP(r), sessionID(r))
	if err != nil {
		httpx.Error(w, status, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if !json.Valid(responseBody) {
		_ = json.NewEncoder(w).Encode(map[string]string{"raw": string(responseBody)})
		return
	}
	_, _ = w.Write(responseBody)
}

func sessionID(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Session-Id")); value != "" {
		return value
	}
	return strings.TrimSpace(r.Header.Get("OpenAI-Conversation-ID"))
}

func statusForGatewayError(err error) int {
	switch {
	case errors.Is(err, service.ErrInsufficientBalance):
		return http.StatusPaymentRequired
	case errors.Is(err, service.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, service.ErrNoHealthyChannel):
		return http.StatusBadGateway
	default:
		return http.StatusBadGateway
	}
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

package gateway

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
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
		httpx.ErrorJSON(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key", "invalid_api_key")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			message := "request body exceeds " + formatBytes(maxBytesErr.Limit)
			httpx.ErrorJSON(w, http.StatusRequestEntityTooLarge, "request_too_large", message, "request_too_large")
			return
		}
		httpx.ErrorJSON(w, http.StatusBadRequest, "invalid_request_error", "invalid body", "invalid_body")
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
			writeGatewayError(w, statusForGatewayError(err), err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			responseBody, _ := io.ReadAll(resp.Body)
			h.gateway.FinalizeStream(r.Context(), principalDTO, model, channel, body, responseBody, estimate, resp.StatusCode, clientIP(r), start)
			writeGatewayBody(w, resp.StatusCode, responseBody)
			return
		}
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
		writeGatewayError(w, status, err)
		return
	}
	writeGatewayBody(w, status, responseBody)
}

func sessionID(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Session-Id")); value != "" {
		return value
	}
	return strings.TrimSpace(r.Header.Get("OpenAI-Conversation-ID"))
}

func statusForGatewayError(err error) int {
	switch {
	case isUpstreamError(err):
		var upstreamErr *service.UpstreamError
		if errors.As(err, &upstreamErr) {
			return upstreamErr.StatusCode
		}
		return http.StatusBadGateway
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

func isUpstreamError(err error) bool {
	var upstreamErr *service.UpstreamError
	return errors.As(err, &upstreamErr)
}

func writeGatewayError(w http.ResponseWriter, fallbackStatus int, err error) {
	var upstreamErr *service.UpstreamError
	if errors.As(err, &upstreamErr) {
		writeGatewayBody(w, upstreamErr.StatusCode, service.NormalizeUpstreamErrorBody(upstreamErr.StatusCode, upstreamErr.Body))
		return
	}
	if fallbackStatus < 400 {
		fallbackStatus = statusForGatewayError(err)
	}
	writeGatewayLocalError(w, fallbackStatus, err)
}

func writeGatewayBody(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if !json.Valid(body) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message":         strings.TrimSpace(string(body)),
				"type":            "upstream_error",
				"upstream_status": status,
			},
		})
		return
	}
	_, _ = w.Write(body)
}

func writeGatewayLocalError(w http.ResponseWriter, status int, err error) {
	switch {
	case errors.Is(err, service.ErrInsufficientBalance):
		httpx.ErrorJSON(w, http.StatusPaymentRequired, "insufficient_balance", "insufficient account balance", "insufficient_balance")
	case errors.Is(err, service.ErrRateLimited):
		httpx.ErrorJSON(w, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded", "rate_limit_exceeded")
	case errors.Is(err, service.ErrNoHealthyChannel):
		httpx.ErrorJSON(w, http.StatusBadGateway, "upstream_unavailable", "no healthy upstream channel", "no_healthy_channel")
	default:
		httpx.ErrorJSON(w, status, "gateway_error", err.Error(), "gateway_error")
	}
}

func formatBytes(limit int64) string {
	if limit > 0 && limit%(1024*1024) == 0 {
		return strconv.FormatInt(limit/(1024*1024), 10) + " MiB"
	}
	return strconv.FormatInt(limit, 10) + " bytes"
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

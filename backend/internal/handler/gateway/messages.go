package gateway

import (
	"errors"
	"io"
	"net/http"
	"time"

	"lingshu/backend/internal/middleware"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
	"lingshu/backend/internal/upstream"
)

// Messages 处理 Anthropic 原生的 POST /v1/messages 端点。
// 入站请求转成内部 OpenAI 格式，复用 Chat/OpenChatStream/FinalizeStream 全链路（含计费），
// 响应再转回 Anthropic Messages 格式。本期仅支持纯文本对话。
func (h Handler) Messages(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentGatewayPrincipal(r.Context())
	if !ok {
		writeAnthropicError(w, http.StatusUnauthorized, "authentication_error", "invalid api key")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeAnthropicError(w, http.StatusRequestEntityTooLarge, "invalid_request_error", "request body exceeds "+formatBytes(maxBytesErr.Limit))
			return
		}
		writeAnthropicError(w, http.StatusBadRequest, "invalid_request_error", "invalid body")
		return
	}

	openAIBody, isStream, err := upstream.AnthropicInboundToOpenAI(body)
	if err != nil {
		writeAnthropicError(w, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	principalDTO := service.GatewayPrincipal{
		UserID:           principal.UserID,
		APIKeyID:         principal.APIKeyID,
		Balance:          principal.Balance,
		RPMLimit:         principal.RPMLimit,
		ConcurrencyLimit: principal.ConcurrencyLimit,
	}

	if !isStream {
		status, responseBody, chatErr := h.gateway.Chat(r.Context(), principalDTO, openAIBody, clientIP(r), sessionID(r))
		if chatErr != nil {
			writeAnthropicErrorFromGateway(w, status, chatErr)
			return
		}
		if status >= 400 {
			writeAnthropicErrorBody(w, status, responseBody)
			return
		}
		converted, convErr := upstream.OpenAIToAnthropicResponse(responseBody)
		if convErr != nil {
			writeAnthropicError(w, http.StatusBadGateway, "api_error", "failed to convert upstream response")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(converted)
		return
	}

	start := time.Now()
	model, channel, estimate, resp, streamErr := h.gateway.OpenChatStream(r.Context(), principalDTO, openAIBody, sessionID(r))
	if streamErr != nil {
		writeAnthropicErrorFromGateway(w, statusForGatewayError(streamErr), streamErr)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		h.gateway.FinalizeStream(r.Context(), principalDTO, model, channel, openAIBody, responseBody, estimate, resp.StatusCode, clientIP(r), start)
		writeAnthropicErrorBody(w, resp.StatusCode, responseBody)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	flush := func() {
		if flusher != nil {
			flusher.Flush()
		}
	}
	captured, _ := upstream.StreamOpenAIToAnthropic(w, flush, resp.Body, model.PublicName)
	h.gateway.FinalizeStream(r.Context(), principalDTO, model, channel, openAIBody, captured, estimate, resp.StatusCode, clientIP(r), start)
}

func anthropicErrorTypeForStatus(status int) string {
	switch {
	case status == http.StatusUnauthorized:
		return "authentication_error"
	case status == http.StatusTooManyRequests:
		return "rate_limit_error"
	case status >= 400 && status < 500:
		return "invalid_request_error"
	default:
		return "api_error"
	}
}

func writeAnthropicError(w http.ResponseWriter, status int, errType, message string) {
	httpx.JSON(w, status, map[string]any{
		"type":  "error",
		"error": map[string]any{"type": errType, "message": message},
	})
}

// writeAnthropicErrorFromGateway 把 service 层的网关错误映射成 Anthropic 错误格式。
func writeAnthropicErrorFromGateway(w http.ResponseWriter, fallbackStatus int, err error) {
	var upstreamErr *service.UpstreamError
	if errors.As(err, &upstreamErr) {
		writeAnthropicErrorBody(w, upstreamErr.StatusCode, service.NormalizeUpstreamErrorBody(upstreamErr.StatusCode, upstreamErr.Body))
		return
	}
	switch {
	case errors.Is(err, service.ErrInsufficientBalance):
		writeAnthropicError(w, http.StatusPaymentRequired, "invalid_request_error", "insufficient account balance")
	case errors.Is(err, service.ErrRateLimited):
		writeAnthropicError(w, http.StatusTooManyRequests, "rate_limit_error", "rate limit exceeded")
	case errors.Is(err, service.ErrNoHealthyChannel):
		writeAnthropicError(w, http.StatusBadGateway, "api_error", "no healthy upstream channel")
	default:
		if fallbackStatus < 400 {
			fallbackStatus = http.StatusBadGateway
		}
		writeAnthropicError(w, fallbackStatus, anthropicErrorTypeForStatus(fallbackStatus), err.Error())
	}
}

// writeAnthropicErrorBody 把上游返回的（OpenAI 风格）错误体提取 message 后包成 Anthropic 错误格式。
func writeAnthropicErrorBody(w http.ResponseWriter, status int, body []byte) {
	message := upstream.ExtractErrorMessage(body)
	if message == "" {
		message = http.StatusText(status)
	}
	writeAnthropicError(w, status, anthropicErrorTypeForStatus(status), message)
}

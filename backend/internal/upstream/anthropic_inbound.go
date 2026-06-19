package upstream

// 入站方向（Anthropic 客户端 → 网关）的格式转换。
// 网关内部统一以 OpenAI 格式为中间格式，本文件负责：
//   1. 把 Anthropic Messages API 请求转成 OpenAI chat completions 请求
//   2. 把 OpenAI 响应/SSE 流转回 Anthropic Messages 格式
// 注意：与 anthropic_adapter.go 的 BuildAnthropicBody 等是相反方向，不可混用。
// 本期仅支持纯文本对话（system + 多轮 text），忽略 image / tool_use / tool_result 块。

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

type anthropicInboundRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      json.RawMessage    `json:"system,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Stop        []string           `json:"stop_sequences,omitempty"`
}

// AnthropicInboundToOpenAI 把 Anthropic Messages 请求体转成 OpenAI chat completions 请求体。
// 返回转换后的 body、是否流式、错误。model 保留客户端公共名（service 据此 FindEnabledModel）。
func AnthropicInboundToOpenAI(rawBody []byte) ([]byte, bool, error) {
	var req anthropicInboundRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, false, errors.New("model is required")
	}

	messages := make([]map[string]any, 0, len(req.Messages)+1)
	if system := systemText(req.System); system != "" {
		messages = append(messages, map[string]any{"role": "system", "content": system})
	}
	for _, msg := range req.Messages {
		role := msg.Role
		if role != "assistant" && role != "system" {
			role = "user"
		}
		messages = append(messages, map[string]any{"role": role, "content": contentText(msg.Content)})
	}

	payload := map[string]any{
		"model":    req.Model,
		"messages": messages,
		"stream":   req.Stream,
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		payload["top_p"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		payload["stop"] = req.Stop
	}

	out, err := json.Marshal(payload)
	if err != nil {
		return nil, false, err
	}
	return out, req.Stream, nil
}

// systemText 把 Anthropic 顶层 system 字段（string 或 text block 数组）拍平成纯文本。
func systemText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return contentText(value)
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
}

// OpenAIToAnthropicResponse 把非流式 OpenAI 响应转回 Anthropic Messages 响应格式。
func OpenAIToAnthropicResponse(openAIBody []byte) ([]byte, error) {
	var in openAIChatResponse
	if err := json.Unmarshal(openAIBody, &in); err != nil {
		return nil, err
	}
	text := ""
	finish := ""
	if len(in.Choices) > 0 {
		text = in.Choices[0].Message.Content
		finish = in.Choices[0].FinishReason
	}
	content := []map[string]any{}
	if text != "" {
		content = append(content, map[string]any{"type": "text", "text": text})
	}
	out := map[string]any{
		"id":            firstNonEmpty(in.ID, newAnthropicMessageID()),
		"type":          "message",
		"role":          "assistant",
		"model":         in.Model,
		"content":       content,
		"stop_reason":   openAIFinishToAnthropic(finish),
		"stop_sequence": nil,
		"usage": map[string]any{
			"input_tokens":            in.Usage.PromptTokens,
			"output_tokens":           in.Usage.CompletionTokens,
			"cache_read_input_tokens": in.Usage.PromptTokensDetails.CachedTokens,
		},
	}
	return json.Marshal(out)
}

func openAIFinishToAnthropic(reason string) string {
	switch reason {
	case "stop", "":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	default:
		return reason
	}
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
}

// StreamOpenAIToAnthropic 边读 OpenAI SSE 边转成 Anthropic SSE 事件写给客户端，
// 同时通过 TeeReader 把 OpenAI 原始流完整镜像到返回值，供 FinalizeStream 解析 usage 扣费。
// 事件序列：message_start → content_block_start → content_block_delta* →
//
//	content_block_stop → message_delta(stop_reason+usage) → message_stop。
func StreamOpenAIToAnthropic(w io.Writer, flush func(), openAISSE io.Reader, model string) ([]byte, error) {
	return StreamOpenAIToAnthropicWithFirstChunk(w, flush, openAISSE, model, nil)
}

func StreamOpenAIToAnthropicWithFirstChunk(w io.Writer, flush func(), openAISSE io.Reader, model string, onFirstChunk func()) ([]byte, error) {
	var captured bytes.Buffer
	tee := io.TeeReader(openAISSE, &captured)

	messageID := newAnthropicMessageID()
	started := false
	firstChunkSeen := false
	stopReason := "end_turn"
	outputTokens := 0
	inputTokens := 0
	cacheReadTokens := 0

	emitStart := func() error {
		if started {
			return nil
		}
		started = true
		if err := writeAnthropicSSEEvent(w, flush, "message_start", map[string]any{
			"type": "message_start",
			"message": map[string]any{
				"id":            messageID,
				"type":          "message",
				"role":          "assistant",
				"model":         model,
				"content":       []any{},
				"stop_reason":   nil,
				"stop_sequence": nil,
				"usage":         map[string]any{"input_tokens": 0, "output_tokens": 0},
			},
		}); err != nil {
			return err
		}
		return writeAnthropicSSEEvent(w, flush, "content_block_start", map[string]any{
			"type":          "content_block_start",
			"index":         0,
			"content_block": map[string]any{"type": "text", "text": ""},
		})
	}

	err := parseAnthropicSSE(tee, func(data string) error {
		if data == "[DONE]" {
			return io.EOF
		}
		var chunk openAIStreamChunk
		if jsonErr := json.Unmarshal([]byte(data), &chunk); jsonErr != nil {
			return nil
		}
		if chunk.Usage != nil {
			if chunk.Usage.PromptTokens > 0 {
				inputTokens = chunk.Usage.PromptTokens
			}
			if chunk.Usage.CompletionTokens > 0 {
				outputTokens = chunk.Usage.CompletionTokens
			}
			if chunk.Usage.PromptTokensDetails.CachedTokens > 0 {
				cacheReadTokens = chunk.Usage.PromptTokensDetails.CachedTokens
			}
		}
		if len(chunk.Choices) == 0 {
			return nil
		}
		choice := chunk.Choices[0]
		if choice.FinishReason != "" {
			stopReason = openAIFinishToAnthropic(choice.FinishReason)
		}
		if choice.Delta.Content != "" {
			if !firstChunkSeen {
				firstChunkSeen = true
				if onFirstChunk != nil {
					onFirstChunk()
				}
			}
			if emitErr := emitStart(); emitErr != nil {
				return emitErr
			}
			return writeAnthropicSSEEvent(w, flush, "content_block_delta", map[string]any{
				"type":  "content_block_delta",
				"index": 0,
				"delta": map[string]any{"type": "text_delta", "text": choice.Delta.Content},
			})
		}
		return nil
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return captured.Bytes(), err
	}

	// 保证即使没有任何 delta（空响应）也产出合法事件序列。
	if startErr := emitStart(); startErr != nil {
		return captured.Bytes(), startErr
	}
	if err := writeAnthropicSSEEvent(w, flush, "content_block_stop", map[string]any{
		"type":  "content_block_stop",
		"index": 0,
	}); err != nil {
		return captured.Bytes(), err
	}
	if err := writeAnthropicSSEEvent(w, flush, "message_delta", map[string]any{
		"type":  "message_delta",
		"delta": map[string]any{"stop_reason": stopReason, "stop_sequence": nil},
		"usage": map[string]any{"input_tokens": inputTokens, "output_tokens": outputTokens, "cache_read_input_tokens": cacheReadTokens},
	}); err != nil {
		return captured.Bytes(), err
	}
	if err := writeAnthropicSSEEvent(w, flush, "message_stop", map[string]any{
		"type": "message_stop",
	}); err != nil {
		return captured.Bytes(), err
	}
	return captured.Bytes(), nil
}

func writeAnthropicSSEEvent(w io.Writer, flush func(), event string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, "event: "+event+"\ndata: "); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\n\n"); err != nil {
		return err
	}
	if flush != nil {
		flush()
	}
	return nil
}

// parseAnthropicSSE 复用 anthropic_adapter.go 中的 SSE 逐帧解析器，它按 data: 前缀提取，
// 对 OpenAI SSE 同样适用。

func newAnthropicMessageID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "msg_" + hex.EncodeToString([]byte(time.Now().Format("150405.000")))
	}
	return "msg_" + hex.EncodeToString(buf)
}

// ExtractErrorMessage 从上游错误体（OpenAI 风格 {error:{message}} 或裸 {message}）提取人类可读消息。
func ExtractErrorMessage(raw []byte) string {
	var parsed struct {
		Message string `json:"message"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err == nil {
		if msg := firstNonEmpty(parsed.Error.Message, parsed.Message); msg != "" {
			return msg
		}
	}
	return strings.TrimSpace(string(raw))
}

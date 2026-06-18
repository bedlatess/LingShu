package upstream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type AnthropicAdapter struct{}

type openAIChatPayload struct {
	Model       string                     `json:"model"`
	Messages    []openAIMessage            `json:"messages"`
	MaxTokens   int                        `json:"max_tokens,omitempty"`
	Temperature *float64                   `json:"temperature,omitempty"`
	TopP        *float64                   `json:"top_p,omitempty"`
	Stream      bool                       `json:"stream,omitempty"`
	Metadata    map[string]any             `json:"metadata,omitempty"`
	Extra       map[string]json.RawMessage `json:"-"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (AnthropicAdapter) ForwardChat(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (ChatResponse, error) {
	body, err := BuildAnthropicBody(rawBody, upstreamModelName, false)
	if err != nil {
		return ChatResponse{}, err
	}
	resp, err := doAnthropic(ctx, baseURL, apiKey, timeoutSeconds, body)
	if err != nil {
		return ChatResponse{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, err
	}
	if resp.StatusCode >= 400 {
		return ChatResponse{StatusCode: resp.StatusCode, Body: respBody}, nil
	}
	bodyOut, usage := AnthropicResponseToOpenAI(respBody)
	return ChatResponse{StatusCode: resp.StatusCode, Body: bodyOut, Usage: usage}, nil
}

func (AnthropicAdapter) OpenChatStream(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (*http.Response, error) {
	body, err := BuildAnthropicBody(rawBody, upstreamModelName, true)
	if err != nil {
		return nil, err
	}
	resp, err := doAnthropic(ctx, baseURL, apiKey, timeoutSeconds, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return resp, nil
	}
	resp.Body = StreamAnthropicToOpenAI(resp.Body)
	resp.Header.Set("Content-Type", "text/event-stream")
	resp.Header.Del("Content-Length")
	return resp, nil
}

func (AnthropicAdapter) ListModels(ctx context.Context, baseURL, apiKey string) ([]ProviderModel, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, anthropicURL(baseURL, "/models"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := client.Do(req)
	if err != nil {
		return anthropicPresetModels(), nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode >= 400 {
		return anthropicPresetModels(), nil
	}
	if err := ensureJSONResponse(resp, body); err != nil {
		return nil, err
	}
	var parsed struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || len(parsed.Data) == 0 {
		return anthropicPresetModels(), nil
	}
	models := make([]ProviderModel, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		models = append(models, ProviderModel{ID: id, Type: "chat", Owned: firstNonEmpty(item.DisplayName, "anthropic")})
	}
	if len(models) == 0 {
		return anthropicPresetModels(), nil
	}
	return models, nil
}

func doAnthropic(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, body []byte) (*http.Response, error) {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicURL(baseURL, "/messages"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return client.Do(req)
}

func anthropicURL(baseURL, path string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + path
	}
	return trimmed + "/v1" + path
}

func anthropicPresetModels() []ProviderModel {
	names := []string{
		"claude-opus-4-1-20250805",
		"claude-opus-4-20250514",
		"claude-sonnet-4-20250514",
		"claude-3-7-sonnet-20250219",
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-haiku-20240307",
	}
	out := make([]ProviderModel, 0, len(names))
	for _, name := range names {
		out = append(out, ProviderModel{ID: name, Type: "chat", Owned: "anthropic"})
	}
	return out
}

func BuildAnthropicBody(rawBody []byte, upstreamModelName string, forceStream bool) ([]byte, error) {
	var input openAIChatPayload
	if err := json.Unmarshal(rawBody, &input); err != nil {
		return nil, err
	}
	modelName := strings.TrimSpace(upstreamModelName)
	if modelName == "" {
		modelName = input.Model
	}
	maxTokens := input.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	out := anthropicRequest{
		Model:       modelName,
		MaxTokens:   maxTokens,
		Temperature: input.Temperature,
		TopP:        input.TopP,
		Stream:      forceStream || input.Stream,
		Metadata:    input.Metadata,
	}
	systemParts := []string{}
	for _, msg := range input.Messages {
		switch msg.Role {
		case "system":
			if text := contentText(msg.Content); text != "" {
				systemParts = append(systemParts, text)
			}
		case "assistant":
			out.Messages = append(out.Messages, anthropicMessage{Role: "assistant", Content: anthropicContent(msg.Content)})
		default:
			out.Messages = append(out.Messages, anthropicMessage{Role: "user", Content: anthropicContent(msg.Content)})
		}
	}
	out.System = strings.Join(systemParts, "\n\n")
	return json.Marshal(out)
}

func AnthropicResponseToOpenAI(raw []byte) ([]byte, Usage) {
	var input anthropicResponse
	if err := json.Unmarshal(raw, &input); err != nil {
		return raw, Usage{}
	}
	textParts := []string{}
	for _, item := range input.Content {
		if item.Type == "text" {
			textParts = append(textParts, item.Text)
		}
	}
	usage := Usage{
		PromptTokens:     input.Usage.InputTokens,
		CompletionTokens: input.Usage.OutputTokens,
		TotalTokens:      input.Usage.InputTokens + input.Usage.OutputTokens,
	}
	out := map[string]any{
		"id":      firstNonEmpty(input.ID, "chatcmpl-anthropic"),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   input.Model,
		"choices": []map[string]any{{
			"index": 0,
			"message": map[string]any{
				"role":    "assistant",
				"content": strings.Join(textParts, ""),
			},
			"finish_reason": anthropicFinishReason(input.StopReason),
		}},
		"usage": usage,
	}
	body, err := json.Marshal(out)
	if err != nil {
		return raw, usage
	}
	return body, usage
}

func ConvertAnthropicStream(reader io.Reader) ([]byte, error) {
	items, err := collectAnthropicStreamEvents(reader)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	for _, item := range items {
		out.WriteString(item)
	}
	return out.Bytes(), nil
}

func writeOpenAIStreamChunk(out *bytes.Buffer, model, content string, usage Usage) {
	chunk := map[string]any{
		"id":      "chatcmpl-anthropic-stream",
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]any{{
			"index": 0,
			"delta": map[string]any{"content": content},
		}},
	}
	if usage.TotalTokens > 0 {
		chunk["usage"] = usage
	}
	body, _ := json.Marshal(chunk)
	out.WriteString("data: ")
	out.Write(body)
	out.WriteString("\n\n")
}

func StreamAnthropicToOpenAI(upstream io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		defer upstream.Close()
		defer pw.Close()
		scanner := bufio.NewScanner(upstream)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var model string
		var usage Usage
		writeLine := func(content string, chunkUsage Usage) error {
			var out bytes.Buffer
			writeOpenAIStreamChunk(&out, model, content, chunkUsage)
			_, err := pw.Write(out.Bytes())
			return err
		}
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "" || data == "[DONE]" {
				continue
			}
			var event struct {
				Type  string `json:"type"`
				Model string `json:"model"`
				Delta struct {
					Type         string `json:"type"`
					Text         string `json:"text"`
					StopReason   string `json:"stop_reason"`
					OutputTokens int    `json:"output_tokens"`
				} `json:"delta"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
				Message struct {
					Model string `json:"model"`
					Usage struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			if event.Model != "" {
				model = event.Model
			}
			if event.Message.Model != "" {
				model = event.Message.Model
			}
			if event.Message.Usage.InputTokens > 0 {
				usage.PromptTokens = event.Message.Usage.InputTokens
			}
			if event.Usage.OutputTokens > 0 {
				usage.CompletionTokens = event.Usage.OutputTokens
			}
			if event.Delta.OutputTokens > 0 {
				usage.CompletionTokens = event.Delta.OutputTokens
			}
			if event.Type == "content_block_delta" && event.Delta.Text != "" {
				if err := writeLine(event.Delta.Text, Usage{}); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
		if err := writeLine("", usage); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_, _ = pw.Write([]byte("data: [DONE]\n\n"))
	}()
	return pr
}

func collectAnthropicStreamEvents(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	parts := []string{}
	var model string
	var usage Usage
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var event struct {
			Type  string `json:"type"`
			Model string `json:"model"`
			Delta struct {
				Type         string `json:"type"`
				Text         string `json:"text"`
				StopReason   string `json:"stop_reason"`
				OutputTokens int    `json:"output_tokens"`
			} `json:"delta"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
			Message struct {
				Model string `json:"model"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Model != "" {
			model = event.Model
		}
		if event.Message.Model != "" {
			model = event.Message.Model
		}
		if event.Message.Usage.InputTokens > 0 {
			usage.PromptTokens = event.Message.Usage.InputTokens
		}
		if event.Usage.OutputTokens > 0 {
			usage.CompletionTokens = event.Usage.OutputTokens
		}
		if event.Delta.OutputTokens > 0 {
			usage.CompletionTokens = event.Delta.OutputTokens
		}
		if event.Type == "content_block_delta" && event.Delta.Text != "" {
			var out bytes.Buffer
			writeOpenAIStreamChunk(&out, model, event.Delta.Text, Usage{})
			parts = append(parts, out.String())
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	var out bytes.Buffer
	writeOpenAIStreamChunk(&out, model, "", usage)
	parts = append(parts, out.String(), "data: [DONE]\n\n")
	return parts, nil
}

func contentText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		parts := []string{}
		for _, item := range typed {
			if object, ok := item.(map[string]any); ok {
				if object["type"] == "text" {
					if text, ok := object["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		return strings.Join(parts, "")
	default:
		return ""
	}
}

func anthropicContent(value any) any {
	if text := contentText(value); text != "" {
		return text
	}
	return value
}

func anthropicFinishReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "":
		return "stop"
	default:
		return reason
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

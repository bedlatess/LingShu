package upstream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ChatResponse struct {
	StatusCode int
	Body       []byte
	Usage      Usage
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIAdapter struct{}

func (OpenAIAdapter) ForwardChat(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (ChatResponse, error) {
	return ForwardChat(ctx, baseURL, apiKey, timeoutSeconds, PrepareOpenAIBody(rawBody, upstreamModelName))
}

func (OpenAIAdapter) OpenChatStream(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (*http.Response, error) {
	return OpenChatStream(ctx, baseURL, apiKey, timeoutSeconds, PrepareOpenAIBody(rawBody, upstreamModelName))
}

func (OpenAIAdapter) ListModels(ctx context.Context, baseURL, apiKey string) ([]ProviderModel, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := ensureJSONResponse(resp, body); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, &ProviderError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	var parsed struct {
		Data []struct {
			ID      string `json:"id"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	models := make([]ProviderModel, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		models = append(models, ProviderModel{ID: id, Type: inferModelType(id), Owned: item.OwnedBy})
	}
	return models, nil
}

func ForwardChat(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, body []byte) (ChatResponse, error) {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return ChatResponse{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, err
	}
	return ChatResponse{StatusCode: resp.StatusCode, Body: respBody, Usage: extractUsage(respBody)}, nil
}

type ProviderError struct {
	StatusCode int
	Body       string
}

func (e *ProviderError) Error() string {
	if strings.TrimSpace(e.Body) == "" {
		return http.StatusText(e.StatusCode)
	}
	return strings.TrimSpace(e.Body)
}

type ProviderContentTypeError struct {
	StatusCode  int
	ContentType string
	Body        string
}

func (e *ProviderContentTypeError) Error() string {
	return "上游返回 " + firstNonEmpty(e.ContentType, "空") + " 类型而非 JSON，请检查 base_url 是否正确；HTTP " + strconv.Itoa(e.StatusCode) + " " + http.StatusText(e.StatusCode) + "，body: " + truncateForError(e.Body, 200)
}

func ensureJSONResponse(resp *http.Response, body []byte) error {
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		return nil
	}
	return &ProviderContentTypeError{StatusCode: resp.StatusCode, ContentType: contentType, Body: string(body)}
}

func truncateForError(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func inferModelType(id string) string {
	name := strings.ToLower(id)
	switch {
	case strings.Contains(name, "embedding") || strings.Contains(name, "embed"):
		return "embedding"
	case strings.Contains(name, "image") || strings.Contains(name, "dall-e") || strings.Contains(name, "gpt-image"):
		return "image"
	case strings.Contains(name, "video") || strings.Contains(name, "sora"):
		return "video"
	default:
		return "chat"
	}
}

func OpenChatStream(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, body []byte) (*http.Response, error) {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	return client.Do(req)
}

func PrepareOpenAIBody(rawBody []byte, upstreamModelName string) []byte {
	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return rawBody
	}
	changed := false
	if name := strings.TrimSpace(upstreamModelName); name != "" {
		payload["model"] = name
		changed = true
	}
	// Streaming billing depends on real upstream usage. OpenAI-compatible
	// providers usually emit it only when include_usage is explicitly set.
	if stream, _ := payload["stream"].(bool); stream {
		opts, ok := payload["stream_options"].(map[string]any)
		if !ok {
			opts = map[string]any{}
		}
		if _, exists := opts["include_usage"]; !exists {
			opts["include_usage"] = true
			payload["stream_options"] = opts
			changed = true
		}
	}
	if !changed {
		return rawBody
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return rawBody
	}
	return out
}

func extractUsage(body []byte) Usage {
	var parsed struct {
		Usage Usage `json:"usage"`
	}
	_ = json.Unmarshal(body, &parsed)
	return parsed.Usage
}

// ExtractStreamUsage 从已捕获的 SSE 响应里取上游回灌的真实 usage。
// OpenAI 兼容上游在流式时通常把 usage 放在最后一个非 [DONE] 的 data 帧里
// （需要客户端带 stream_options.include_usage，多数中转站默认就回灌）。
// 取最后一个出现 total_tokens>0 的 usage 为准；取不到返回零值，调用方退回估算。
func ExtractStreamUsage(raw string) Usage {
	var found Usage
	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var parsed struct {
			Usage Usage `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			continue
		}
		if parsed.Usage.TotalTokens > 0 {
			found = parsed.Usage
		}
	}
	return found
}

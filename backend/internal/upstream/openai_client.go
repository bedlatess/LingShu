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

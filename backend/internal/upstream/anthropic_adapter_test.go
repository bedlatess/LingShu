package upstream

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildAnthropicBody(t *testing.T) {
	raw := []byte(`{
		"model":"claude-public",
		"messages":[
			{"role":"system","content":"You are helpful."},
			{"role":"user","content":"Hello"}
		],
		"max_tokens":128,
		"stream":true
	}`)
	body, err := BuildAnthropicBody(raw, "claude-3-5-sonnet-latest", true)
	if err != nil {
		t.Fatalf("BuildAnthropicBody failed: %v", err)
	}
	text := string(body)
	for _, want := range []string{
		`"model":"claude-3-5-sonnet-latest"`,
		`"system":"You are helpful."`,
		`"role":"user"`,
		`"max_tokens":128`,
		`"stream":true`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("body missing %s: %s", want, text)
		}
	}
}

func TestBuildAnthropicBodyUsesDefaultMaxTokens(t *testing.T) {
	raw := []byte(`{"model":"claude-public","messages":[{"role":"user","content":"Hello"}]}`)
	body, err := BuildAnthropicBody(raw, "", false)
	if err != nil {
		t.Fatalf("BuildAnthropicBody failed: %v", err)
	}
	if !strings.Contains(string(body), `"max_tokens":4096`) {
		t.Fatalf("default max_tokens missing: %s", body)
	}
}

func TestAnthropicResponseToOpenAI(t *testing.T) {
	body, usage := AnthropicResponseToOpenAI([]byte(`{
		"id":"msg_1",
		"model":"claude-3-5-sonnet-latest",
		"content":[{"type":"text","text":"你好"}],
		"stop_reason":"end_turn",
		"usage":{"input_tokens":11,"output_tokens":7}
	}`))
	if usage.PromptTokens != 11 || usage.CompletionTokens != 7 || usage.TotalTokens != 18 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	text := string(body)
	for _, want := range []string{`"object":"chat.completion"`, `"content":"你好"`, `"finish_reason":"stop"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("converted response missing %s: %s", want, text)
		}
	}
}

func TestConvertAnthropicStreamKeepsUsageExtractable(t *testing.T) {
	raw := strings.NewReader(strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"model":"claude-3-5-sonnet-latest","usage":{"input_tokens":13}}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"hi"}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`,
		``,
	}, "\n"))
	converted, err := ConvertAnthropicStream(raw)
	if err != nil {
		t.Fatalf("ConvertAnthropicStream failed: %v", err)
	}
	text := string(converted)
	if !strings.Contains(text, `"content":"hi"`) {
		t.Fatalf("converted stream missing content chunk: %s", text)
	}
	usage := ExtractStreamUsage(text)
	if usage.PromptTokens != 13 || usage.CompletionTokens != 5 || usage.TotalTokens != 18 {
		t.Fatalf("usage not extractable by existing OpenAI parser: %+v\n%s", usage, text)
	}
}

func TestAnthropicResponseToOpenAIMapsCacheUsage(t *testing.T) {
	body, usage := AnthropicResponseToOpenAI([]byte(`{
		"id":"msg_cache",
		"model":"claude-cache",
		"content":[{"type":"text","text":"hello"}],
		"stop_reason":"end_turn",
		"usage":{"input_tokens":11,"output_tokens":7,"cache_creation_input_tokens":13,"cache_read_input_tokens":17}
	}`))
	if usage.PromptTokens != 11 || usage.CompletionTokens != 7 || usage.TotalTokens != 18 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	if usage.CacheCreationTokens != 13 || usage.CacheReadTokens != 17 {
		t.Fatalf("unexpected cache usage: %+v", usage)
	}
	text := string(body)
	for _, want := range []string{`"cache_creation_tokens":13`, `"cache_read_tokens":17`} {
		if !strings.Contains(text, want) {
			t.Fatalf("converted response missing %s: %s", want, text)
		}
	}
}

func TestAnthropicStreamMapsCacheUsage(t *testing.T) {
	raw := strings.NewReader(strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"model":"claude-cache","usage":{"input_tokens":13,"cache_creation_input_tokens":3,"cache_read_input_tokens":5}}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"hi"}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}`,
		``,
	}, "\n"))
	converted, err := ConvertAnthropicStream(raw)
	if err != nil {
		t.Fatalf("ConvertAnthropicStream failed: %v", err)
	}
	usage := ExtractStreamUsage(string(converted))
	if usage.CacheCreationTokens != 3 || usage.CacheReadTokens != 5 {
		t.Fatalf("cache usage not extractable: %+v\n%s", usage, string(converted))
	}
}

func TestStreamAnthropicToOpenAIPreservesRealSSEContent(t *testing.T) {
	raw := strings.Join([]string{
		"event: message_start\r",
		`data: {"type":"message_start","message":{"model":"claude-3-5-sonnet-latest","usage":{"input_tokens":452}}}` + "\r",
		"\r",
		"event: content_block_delta\r",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"你好，"}}` + "\r",
		"\r",
		"event: content_block_delta\r",
		`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"这是流式内容。"}}` + "\r",
		"\r",
		"event: message_delta\r",
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":161}}` + "\r",
		"\r",
		"event: message_stop\r",
		`data: {"type":"message_stop"}` + "\r",
		"\r",
	}, "\n")

	converted, err := io.ReadAll(StreamAnthropicToOpenAI(io.NopCloser(strings.NewReader(raw))))
	if err != nil {
		t.Fatalf("read converted stream: %v", err)
	}
	text := string(converted)
	for _, want := range []string{`"content":"你好，"`, `"content":"这是流式内容。"`, `data: [DONE]`} {
		if !strings.Contains(text, want) {
			t.Fatalf("converted stream missing %s: %s", want, text)
		}
	}
	usage := ExtractStreamUsage(text)
	if usage.PromptTokens != 452 || usage.CompletionTokens != 161 || usage.TotalTokens != 613 {
		t.Fatalf("usage mismatch: %+v\n%s", usage, text)
	}
}

func TestAnthropicErrorToOpenAIPreservesCodeAndMessage(t *testing.T) {
	got := AnthropicErrorToOpenAI([]byte(`{"code":"INSUFFICIENT_BALANCE","message":"Insufficient account balance"}`))
	var parsed struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if parsed.Error.Message != "Insufficient account balance" {
		t.Fatalf("message = %q, want preserved upstream message", parsed.Error.Message)
	}
	if parsed.Error.Code != "INSUFFICIENT_BALANCE" {
		t.Fatalf("code = %q, want preserved upstream code", parsed.Error.Code)
	}
	if parsed.Error.Type != "upstream_error" {
		t.Fatalf("type = %q, want upstream_error", parsed.Error.Type)
	}
}

func TestAnthropicOpenChatStreamReturnsJSONErrorBeforeSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("path = %s, want /v1/messages", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"INSUFFICIENT_BALANCE","message":"Insufficient account balance"}`))
	}))
	defer server.Close()

	raw := []byte(`{"model":"claude-public","messages":[{"role":"user","content":"hi"}],"stream":true}`)
	resp, err := (AnthropicAdapter{}).OpenChatStream(t.Context(), server.URL, "test-key", 5, raw, "claude-upstream")
	if err != nil {
		t.Fatalf("OpenChatStream failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("content-type = %q, want application/json", resp.Header.Get("Content-Type"))
	}
	text := string(body)
	for _, want := range []string{`"type":"upstream_error"`, `"code":"INSUFFICIENT_BALANCE"`, `"message":"Insufficient account balance"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("body missing %s: %s", want, text)
		}
	}
}

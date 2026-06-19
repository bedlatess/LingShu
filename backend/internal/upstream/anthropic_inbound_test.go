package upstream

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAnthropicInboundToOpenAI(t *testing.T) {
	raw := []byte(`{
		"model":"claude-opus-4-7",
		"max_tokens":64,
		"system":"You are concise.",
		"messages":[{"role":"user","content":"Hello"}],
		"stream":true
	}`)
	body, isStream, err := AnthropicInboundToOpenAI(raw)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if !isStream {
		t.Fatalf("expected stream=true")
	}
	text := string(body)
	for _, want := range []string{
		`"model":"claude-opus-4-7"`,
		`"role":"system"`,
		`"content":"You are concise."`,
		`"role":"user"`,
		`"max_tokens":64`,
		`"stream":true`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("body missing %s: %s", want, text)
		}
	}
}

func TestAnthropicInboundSystemAndContentBlocks(t *testing.T) {
	// system 与 content 均为 block 数组，验证拍平。
	raw := []byte(`{
		"model":"claude-opus-4-7",
		"max_tokens":32,
		"system":[{"type":"text","text":"Be brief."}],
		"messages":[{"role":"user","content":[{"type":"text","text":"Hi there"}]}]
	}`)
	body, _, err := AnthropicInboundToOpenAI(raw)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, `"content":"Be brief."`) {
		t.Fatalf("system block not flattened: %s", text)
	}
	if !strings.Contains(text, `"content":"Hi there"`) {
		t.Fatalf("user content block not flattened: %s", text)
	}
}

func TestAnthropicInboundRequiresModel(t *testing.T) {
	if _, _, err := AnthropicInboundToOpenAI([]byte(`{"max_tokens":8,"messages":[]}`)); err == nil {
		t.Fatalf("expected error for missing model")
	}
}

func TestOpenAIToAnthropicResponse(t *testing.T) {
	openAI := []byte(`{
		"id":"chatcmpl-1",
		"model":"claude-opus-4-7",
		"choices":[{"message":{"role":"assistant","content":"Hi."},"finish_reason":"stop"}],
		"usage":{"prompt_tokens":10,"completion_tokens":3,"total_tokens":13,"prompt_tokens_details":{"cached_tokens":4}}
	}`)
	out, err := OpenAIToAnthropicResponse(openAI)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	var parsed struct {
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens          int `json:"input_tokens"`
			OutputTokens         int `json:"output_tokens"`
			CacheReadInputTokens int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid output json: %v", err)
	}
	if parsed.Type != "message" || parsed.Role != "assistant" {
		t.Fatalf("unexpected envelope: %s", out)
	}
	if len(parsed.Content) != 1 || parsed.Content[0].Text != "Hi." {
		t.Fatalf("unexpected content: %s", out)
	}
	if parsed.StopReason != "end_turn" {
		t.Fatalf("stop_reason want end_turn got %s", parsed.StopReason)
	}
	if parsed.Usage.InputTokens != 10 || parsed.Usage.OutputTokens != 3 || parsed.Usage.CacheReadInputTokens != 4 {
		t.Fatalf("usage mismatch: %s", out)
	}
}

func TestStreamOpenAIToAnthropic(t *testing.T) {
	// 模拟 OpenAI SSE：两个文本 delta + 带 usage 的末帧 + [DONE]。
	openAISSE := strings.Join([]string{
		`data: {"choices":[{"delta":{"content":"Hel"},"finish_reason":null}]}`,
		``,
		`data: {"choices":[{"delta":{"content":"lo"},"finish_reason":"stop"}]}`,
		``,
		`data: {"choices":[],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7,"prompt_tokens_details":{"cached_tokens":4}}}`,
		``,
		`data: [DONE]`,
		``,
	}, "\n")

	var out bytes.Buffer
	captured, err := StreamOpenAIToAnthropic(&out, nil, strings.NewReader(openAISSE), "claude-opus-4-7")
	if err != nil {
		t.Fatalf("stream convert failed: %v", err)
	}

	anthropic := out.String()
	for _, want := range []string{
		"event: message_start",
		"event: content_block_start",
		"event: content_block_delta",
		`"text":"Hel"`,
		`"text":"lo"`,
		"event: content_block_stop",
		"event: message_delta",
		`"stop_reason":"end_turn"`,
		`"output_tokens":2`,
		`"cache_read_input_tokens":4`,
		"event: message_stop",
	} {
		if !strings.Contains(anthropic, want) {
			t.Fatalf("anthropic stream missing %q:\n%s", want, anthropic)
		}
	}

	// captured 必须是完整 OpenAI 原始流（含 usage 末帧），供 FinalizeStream 扣费。
	if got := ExtractStreamUsage(string(captured)); got.TotalTokens != 7 || got.CacheReadTokens != 4 {
		t.Fatalf("captured OpenAI usage mismatch: got total=%d cache_read=%d\ncaptured:\n%s", got.TotalTokens, got.CacheReadTokens, captured)
	}
}

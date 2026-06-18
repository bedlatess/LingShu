package upstream

import (
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

package upstream

import "testing"

func TestExtractUsageMapsOpenAICachedPromptTokens(t *testing.T) {
	usage := extractUsage([]byte(`{
		"usage":{
			"prompt_tokens":100,
			"completion_tokens":20,
			"total_tokens":120,
			"prompt_tokens_details":{"cached_tokens":80}
		}
	}`))

	if usage.PromptTokens != 100 || usage.CompletionTokens != 20 || usage.TotalTokens != 120 {
		t.Fatalf("unexpected token usage: %+v", usage)
	}
	if usage.CacheReadTokens != 80 {
		t.Fatalf("cache read tokens = %d, want 80", usage.CacheReadTokens)
	}
}

func TestExtractStreamUsageMapsOpenAICachedPromptTokens(t *testing.T) {
	raw := "data: {\"usage\":{\"prompt_tokens\":100,\"completion_tokens\":20,\"total_tokens\":120,\"prompt_tokens_details\":{\"cached_tokens\":80}}}\n\n" +
		"data: [DONE]\n\n"

	usage := ExtractStreamUsage(raw)
	if usage.CacheReadTokens != 80 {
		t.Fatalf("cache read tokens = %d, want 80; usage=%+v", usage.CacheReadTokens, usage)
	}
}

package billing

import "testing"

func TestEstimateTokensUsesTokenizerForUnicode(t *testing.T) {
	tokens := EstimateTokens("你好，灵枢")
	if tokens <= 0 {
		t.Fatalf("expected positive token count, got %d", tokens)
	}
	if tokens > 20 {
		t.Fatalf("unexpectedly high token count: %d", tokens)
	}
}

func TestExtractSSEText(t *testing.T) {
	raw := "data: {\"choices\":[{\"delta\":{\"content\":\"你\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"好\"}}]}\n\n" +
		"data: [DONE]\n\n"
	got := ExtractSSEText(raw)
	if got != "你好" {
		t.Fatalf("expected extracted content, got %q", got)
	}
	if EstimateStreamTokens(raw) <= 0 {
		t.Fatal("expected positive stream token estimate")
	}
}

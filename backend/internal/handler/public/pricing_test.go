package public

import (
	"strings"
	"testing"
)

func TestPublicModelsResponseDoesNotExposeSensitiveBillingFields(t *testing.T) {
	body, err := encodePublicModelsForTest([]PublicModelDTO{{
		ID:               "model-1",
		PublicName:       "gpt-public",
		Type:             "chat",
		Group:            "通用",
		BillingMode:      "token",
		InputPricePer1M:  "1.440000",
		OutputPricePer1M: "2.880000",
		Currency:         "USD",
	}})
	if err != nil {
		t.Fatalf("marshal public models: %v", err)
	}
	text := string(body)
	for _, forbidden := range []string{"base_cost", "rate_multiplier", "gross_profit"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("public response leaked %q: %s", forbidden, text)
		}
	}
}

func TestPublicPricePer1MUsesMultiplier(t *testing.T) {
	got := publicPricePer1M("0.001000", "1.200")
	if got != "1.200000" {
		t.Fatalf("price = %s, want 1.200000", got)
	}
}

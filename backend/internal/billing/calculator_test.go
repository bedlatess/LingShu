package billing

import "testing"

func TestCalculateTokenChargeUsesBaseCostTimesMultiplier(t *testing.T) {
	got := CalculateTokenCharge(
		TokenPricing{
			InputPricePer1K:  2_000,
			OutputPricePer1K: 6_000,
			RateMultiplier:   1_300_000,
		},
		TokenUsage{InputTokens: 1000, OutputTokens: 500},
	)

	if got.BaseCost != 5_000 {
		t.Fatalf("base cost = %d, want 5000", got.BaseCost)
	}
	if got.Charge != 6_500 {
		t.Fatalf("charge = %d, want 6500", got.Charge)
	}
}

func TestCalculatePerCallChargeUsesMultiplier(t *testing.T) {
	got := CalculatePerCallCharge(10_000, 3, 1_200_000)
	if got.BaseCost != 30_000 {
		t.Fatalf("base cost = %d, want 30000", got.BaseCost)
	}
	if got.Charge != 36_000 {
		t.Fatalf("charge = %d, want 36000", got.Charge)
	}
}

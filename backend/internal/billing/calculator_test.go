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

func TestCalculateTokenChargeIncludesCacheTokensInBaseCost(t *testing.T) {
	got := CalculateTokenCharge(
		TokenPricing{
			InputPricePer1K:         100,
			OutputPricePer1K:        200,
			CacheCreationPricePer1K: 300,
			CacheReadPricePer1K:     50,
			RateMultiplier:          1_250_000,
		},
		TokenUsage{
			InputTokens:         1000,
			OutputTokens:        500,
			CacheCreationTokens: 2000,
			CacheReadTokens:     3000,
		},
	)

	if got.BaseCost != 950 {
		t.Fatalf("base cost = %d, want 950", got.BaseCost)
	}
	if got.Charge != 1_188 {
		t.Fatalf("charge = %d, want 1188", got.Charge)
	}
}

func TestCalculateTokenChargeCachePricesDefaultZeroKeepsOldCost(t *testing.T) {
	pricing := TokenPricing{
		InputPricePer1K:  123,
		OutputPricePer1K: 456,
		RateMultiplier:   1_200_000,
	}
	withoutCache := CalculateTokenCharge(pricing, TokenUsage{InputTokens: 777, OutputTokens: 333})
	withZeroPricedCache := CalculateTokenCharge(pricing, TokenUsage{
		InputTokens:         777,
		OutputTokens:        333,
		CacheCreationTokens: 999_999,
		CacheReadTokens:     888_888,
	})

	if withZeroPricedCache != withoutCache {
		t.Fatalf("zero-priced cache changed charge: with=%+v without=%+v", withZeroPricedCache, withoutCache)
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

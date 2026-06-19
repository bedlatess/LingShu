package billing

import "math"

type TokenPricing struct {
	InputPricePer1K         int64
	OutputPricePer1K        int64
	CacheCreationPricePer1K int64
	CacheReadPricePer1K     int64
	RateMultiplier          int64
}

type TokenUsage struct {
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
}

type Charge struct {
	BaseCost       int64
	RateMultiplier int64
	Charge         int64
}

const Scale int64 = 1_000_000

func CalculateTokenCharge(pricing TokenPricing, usage TokenUsage) Charge {
	inputCost := divCeil(usage.InputTokens*pricing.InputPricePer1K, 1000)
	outputCost := divCeil(usage.OutputTokens*pricing.OutputPricePer1K, 1000)
	cacheCreationCost := divCeil(usage.CacheCreationTokens*pricing.CacheCreationPricePer1K, 1000)
	cacheReadCost := divCeil(usage.CacheReadTokens*pricing.CacheReadPricePer1K, 1000)
	baseCost := inputCost + outputCost + cacheCreationCost + cacheReadCost
	charge := divCeil(baseCost*pricing.RateMultiplier, Scale)
	return Charge{
		BaseCost:       baseCost,
		RateMultiplier: pricing.RateMultiplier,
		Charge:         charge,
	}
}

func CalculatePerCallCharge(pricePerCall, calls, rateMultiplier int64) Charge {
	baseCost := pricePerCall * int64(math.Max(float64(calls), 0))
	return Charge{
		BaseCost:       baseCost,
		RateMultiplier: rateMultiplier,
		Charge:         divCeil(baseCost*rateMultiplier, Scale),
	}
}

func divCeil(numerator, denominator int64) int64 {
	if numerator <= 0 {
		return 0
	}
	return (numerator + denominator - 1) / denominator
}

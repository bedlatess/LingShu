package dto

import (
	"encoding/json"
	"strings"
	"testing"

	"lingshu/backend/internal/service"
)

func TestUserModelConfigDTOsDoNotExposeSensitiveBillingFields(t *testing.T) {
	body, err := json.Marshal(NewUserModelConfigDTOs([]service.UserModelPrice{{
		ID:               "model-1",
		PublicName:       "gpt-user",
		Type:             "chat",
		Group:            "通用",
		BillingMode:      "token",
		InputPricePer1K:  "0.001000",
		OutputPricePer1K: "0.002000",
		PricePerCall:     "0",
		RateMultiplier:   "1.200",
		InputUnitPrice:   "0.001200",
		OutputUnitPrice:  "0.002400",
		CallUnitPrice:    "0",
		Status:           "enabled",
		SortOrder:        1,
	}}))
	if err != nil {
		t.Fatalf("marshal user model dto: %v", err)
	}
	text := string(body)
	for _, forbidden := range []string{"base_cost", "rate_multiplier", "gross_profit", "upstream_model_name"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("user model response leaked %q: %s", forbidden, text)
		}
	}
}

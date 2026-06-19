package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"lingshu/backend/internal/repository"
)

func TestUserGatewayLogDTOsDoNotExposeSensitiveBillingFields(t *testing.T) {
	body, err := json.Marshal(NewUserGatewayLogDTOs([]repository.GatewayLog{{
		RequestID:   "req-1",
		UserID:      "user-1",
		ModelID:     "model-1",
		Status:      "success",
		HTTPStatus:  httpStatusOK,
		TotalTokens: 100,
		BaseCost:    "0.010000",
		Charge:      "0.012000",
		CreatedAt:   time.Now(),
	}}))
	if err != nil {
		t.Fatalf("marshal user gateway log dto: %v", err)
	}
	text := string(body)
	for _, forbidden := range []string{"base_cost", "rate_multiplier", "gross_profit", "upstream_model_name"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("user gateway log response leaked %q: %s", forbidden, text)
		}
	}
}

const httpStatusOK = 200

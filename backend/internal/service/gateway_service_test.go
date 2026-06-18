package service

import (
	"encoding/json"
	"net/http"
	"testing"

	"lingshu/backend/internal/billing"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/upstream"
)

func TestShouldRetryStatus(t *testing.T) {
	retryStatuses := []int{
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	}
	for _, status := range retryStatuses {
		if !shouldRetryStatus(status) {
			t.Fatalf("status %d should retry", status)
		}
	}

	nonRetryStatuses := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusPaymentRequired,
		http.StatusNotFound,
		http.StatusTooManyRequests,
	}
	for _, status := range nonRetryStatuses {
		if shouldRetryStatus(status) {
			t.Fatalf("status %d should not retry", status)
		}
	}
}

func TestNormalizeUpstreamErrorBodyKeepsJSON(t *testing.T) {
	body := []byte(`{"code":"INSUFFICIENT_BALANCE","message":"Insufficient account balance"}`)
	got := NormalizeUpstreamErrorBody(http.StatusForbidden, body)
	if string(got) != string(body) {
		t.Fatalf("body = %s, want original JSON", got)
	}
}

func TestNormalizeUpstreamErrorBodyWrapsText(t *testing.T) {
	got := NormalizeUpstreamErrorBody(http.StatusForbidden, []byte("Forbidden"))
	var parsed struct {
		Error struct {
			Message        string `json:"message"`
			Type           string `json:"type"`
			UpstreamStatus int    `json:"upstream_status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("unmarshal normalized body: %v", err)
	}
	if parsed.Error.Message != "Forbidden" || parsed.Error.Type != "upstream_error" || parsed.Error.UpstreamStatus != http.StatusForbidden {
		t.Fatalf("unexpected normalized body: %s", got)
	}
}

func TestBodyForUpstreamRewritesModel(t *testing.T) {
	raw := []byte(`{"model":"public-chat","messages":[{"role":"user","content":"hi"}]}`)
	out := upstream.PrepareOpenAIBody(raw, "upstream-chat")

	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("unmarshal rewritten body: %v", err)
	}
	if payload["model"] != "upstream-chat" {
		t.Fatalf("model = %v, want upstream-chat", payload["model"])
	}

	unchanged := string(upstream.PrepareOpenAIBody(raw, ""))
	if unchanged != string(raw) {
		t.Fatalf("empty upstream name should keep raw body")
	}
}

func TestBodyWithMaxTokensInjectsDefault(t *testing.T) {
	raw := []byte(`{"model":"public-chat","messages":[{"role":"user","content":"hi"}]}`)
	out := bodyWithMaxTokens(raw, 4096)

	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("unmarshal rewritten body: %v", err)
	}
	if got := int(payload["max_tokens"].(float64)); got != 4096 {
		t.Fatalf("max_tokens = %d, want 4096", got)
	}
}

func TestPerCallGatewayCharge(t *testing.T) {
	model := repository.GatewayModel{
		BillingMode:    "per_call",
		PricePerCall:   "0.020000",
		RateMultiplier: "1.300",
	}
	multiplier, _ := billing.DecimalStringToUnits(model.RateMultiplier)
	got := actualChargeForModel(model, ChatRequest{N: 2}, upstream.Usage{}, multiplier)

	if billing.UnitsToDecimalString(got.BaseCost) != "0.040000" {
		t.Fatalf("base cost = %s, want 0.040000", billing.UnitsToDecimalString(got.BaseCost))
	}
	if billing.UnitsToDecimalString(got.Charge) != "0.052000" {
		t.Fatalf("charge = %s, want 0.052000", billing.UnitsToDecimalString(got.Charge))
	}
}

func TestWeightedRandomOrderKeepsAllChannels(t *testing.T) {
	channels := []repository.GatewayChannel{
		{ID: "a", Weight: 1},
		{ID: "b", Weight: 3},
		{ID: "c", Weight: 6},
	}
	ordered := weightedRandomOrder(channels)
	if len(ordered) != len(channels) {
		t.Fatalf("expected %d channels, got %d", len(channels), len(ordered))
	}
	seen := map[string]bool{}
	for _, channel := range ordered {
		seen[channel.ID] = true
	}
	for _, channel := range channels {
		if !seen[channel.ID] {
			t.Fatalf("missing channel %s", channel.ID)
		}
	}
}

func TestStickyKey(t *testing.T) {
	key := stickyKey(GatewayPrincipal{APIKeyID: "key-1"}, ChatRequest{Model: "gpt-test", User: "end-user"}, "")
	if key != "key-1:gpt-test:end-user" {
		t.Fatalf("unexpected sticky key %q", key)
	}
	headerKey := stickyKey(GatewayPrincipal{APIKeyID: "key-1"}, ChatRequest{Model: "gpt-test", User: "end-user"}, "header-session")
	if headerKey != "key-1:gpt-test:header-session" {
		t.Fatalf("unexpected header sticky key %q", headerKey)
	}
}

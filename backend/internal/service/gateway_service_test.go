package service

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/billing"
	redisstore "lingshu/backend/internal/redis"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/upstream"
)

func TestShouldRetryStatus(t *testing.T) {
	retryStatuses := []int{
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusTooManyRequests,
		529,
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

func TestOrderChannelsSkipsExcludedAndCooling(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := redisstore.NewFrozenStore(client)
	service := GatewayService{frozen: store}

	if err := store.SetStickyChannel(ctx, "sticky", "cooling", time.Minute); err != nil {
		t.Fatalf("set sticky: %v", err)
	}
	if err := store.SetChannelCooldown(ctx, "cooling", "upstream 429", time.Minute); err != nil {
		t.Fatalf("set cooldown: %v", err)
	}

	ordered := service.orderChannels(ctx, "sticky", []repository.GatewayChannel{
		{ID: "excluded", Weight: 100},
		{ID: "cooling", Weight: 100},
		{ID: "healthy", Weight: 1},
	}, map[string]struct{}{"excluded": struct{}{}})

	if len(ordered) != 1 || ordered[0].ID != "healthy" {
		t.Fatalf("ordered channels = %+v, want only healthy", ordered)
	}
}

func TestOrderChannelsSkipsRateLimitedAndOverloaded(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := redisstore.NewFrozenStore(client)
	service := GatewayService{frozen: store}

	if err := store.SetChannelRateLimited(ctx, "rate-limited", "upstream 429", time.Minute); err != nil {
		t.Fatalf("set rate limit memory: %v", err)
	}
	if err := store.SetChannelOverloaded(ctx, "overloaded", "upstream 529", time.Minute); err != nil {
		t.Fatalf("set overload memory: %v", err)
	}

	ordered := service.orderChannels(ctx, "", []repository.GatewayChannel{
		{ID: "rate-limited", Weight: 100},
		{ID: "overloaded", Weight: 100},
		{ID: "healthy", Weight: 1},
	}, nil)

	if len(ordered) != 1 || ordered[0].ID != "healthy" {
		t.Fatalf("ordered channels = %+v, want only healthy", ordered)
	}
}

func TestChannelCooldownForStatusUsesDifferentWindows(t *testing.T) {
	if got := channelCooldownForStatus(http.StatusBadGateway); got != channelFailureCooldown {
		t.Fatalf("5xx cooldown = %s, want %s", got, channelFailureCooldown)
	}
	if got := channelCooldownForStatus(http.StatusTooManyRequests); got != channelRateLimitCooldown {
		t.Fatalf("429 cooldown = %s, want %s", got, channelRateLimitCooldown)
	}
	if got := channelCooldownForStatus(529); got != channelOverloadCooldown {
		t.Fatalf("529 cooldown = %s, want %s", got, channelOverloadCooldown)
	}
	if channelCooldownForStatus(http.StatusTooManyRequests) == channelCooldownForStatus(http.StatusBadGateway) {
		t.Fatalf("429 cooldown should differ from 5xx cooldown")
	}
}

func TestChannelPenaltyTTLsDiffer(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := redisstore.NewFrozenStore(client)

	if err := store.SetChannelCooldown(ctx, "5xx", "bad gateway", channelCooldownForStatus(http.StatusBadGateway)); err != nil {
		t.Fatalf("set 5xx cooldown: %v", err)
	}
	if err := store.SetChannelRateLimited(ctx, "429", "too many requests", channelCooldownForStatus(http.StatusTooManyRequests)); err != nil {
		t.Fatalf("set 429 cooldown: %v", err)
	}
	if err := store.SetChannelOverloaded(ctx, "529", "overloaded", channelCooldownForStatus(529)); err != nil {
		t.Fatalf("set 529 cooldown: %v", err)
	}

	fiveXXTTL := mr.TTL("channel_cooldown:5xx")
	rateLimitTTL := mr.TTL("channel_rate_limited:429")
	overloadTTL := mr.TTL("channel_overload:529")
	if fiveXXTTL != channelFailureCooldown {
		t.Fatalf("5xx TTL = %s, want %s", fiveXXTTL, channelFailureCooldown)
	}
	if rateLimitTTL != channelRateLimitCooldown {
		t.Fatalf("429 TTL = %s, want %s", rateLimitTTL, channelRateLimitCooldown)
	}
	if overloadTTL != channelOverloadCooldown {
		t.Fatalf("529 TTL = %s, want %s", overloadTTL, channelOverloadCooldown)
	}
	if !(fiveXXTTL < rateLimitTTL && rateLimitTTL < overloadTTL) {
		t.Fatalf("expected 5xx < 429 < 529 TTLs, got %s, %s, %s", fiveXXTTL, rateLimitTTL, overloadTTL)
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

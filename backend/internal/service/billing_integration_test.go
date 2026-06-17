package service_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/bootstrap"
	"lingshu/backend/internal/config"
	"lingshu/backend/internal/pkg/apikey"
	"lingshu/backend/internal/pkg/password"
	"lingshu/backend/internal/pkg/token"
	"lingshu/backend/internal/server"
)

type billingHarness struct {
	t        *testing.T
	ctx      context.Context
	db       *pgxpool.Pool
	redis    *redis.Client
	handler  http.Handler
	upstream *httptest.Server
	cfg      config.Config
	userID   string
	adminID  string
	apiKey   string
}

func TestBillingMoneyPathsIntegration(t *testing.T) {
	h := newBillingHarness(t, 200.000000, 20)

	t.Run("token model charges balance ledger and request consistently", func(t *testing.T) {
		status, body := h.chat(t, h.apiKey, "public-token", 1)
		if status != http.StatusOK {
			t.Fatalf("status=%d body=%s", status, body)
		}
		assertMoneyPath(t, h.db, h.userID, "public-token", "0.050000", "1.300", "0.065000")
		assertBalance(t, h.db, h.userID, "199.935000")
	})

	t.Run("per_call model charges price per call times n times multiplier", func(t *testing.T) {
		status, body := h.chat(t, h.apiKey, "public-image", 3)
		if status != http.StatusOK {
			t.Fatalf("status=%d body=%s", status, body)
		}
		assertMoneyPath(t, h.db, h.userID, "public-image", "0.060000", "1.300", "0.078000")
		assertBalance(t, h.db, h.userID, "199.857000")
	})

	t.Run("admin manual grant increases balance and writes ledger", func(t *testing.T) {
		adminToken := mustSignAdmin(t, h.cfg, h.adminID)
		body := []byte(`{"amount":"10.000000","remark":"integration top up"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+h.userID+"/balance", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		assertBalance(t, h.db, h.userID, "209.857000")
		assertLedgerType(t, h.db, h.userID, "admin_grant", "10.000000", "199.857000", "209.857000")
	})

	t.Run("reserve phase insufficient balance returns 402 without charge", func(t *testing.T) {
		low := newBillingHarness(t, 0.010000, 20)
		beforeLedger, beforeRequests := countBillingRows(t, low.db, low.userID)
		status, _ := low.chat(t, low.apiKey, "public-token", 1)
		if status != http.StatusPaymentRequired {
			t.Fatalf("status=%d want 402", status)
		}
		assertBalance(t, low.db, low.userID, "0.010000")
		afterLedger, afterRequests := countBillingRows(t, low.db, low.userID)
		if afterLedger != beforeLedger || afterRequests != beforeRequests {
			t.Fatalf("unexpected billing rows after reserve failure: ledger %d->%d requests %d->%d", beforeLedger, afterLedger, beforeRequests, afterRequests)
		}
	})
}

func TestBillingConcurrentReserveDoesNotOverdraw(t *testing.T) {
	const allowed = 7
	h := newBillingHarness(t, float64(allowed)*0.065000, 100)

	var wg sync.WaitGroup
	results := make(chan int, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, _ := h.chat(t, h.apiKey, "public-token", 1)
			results <- status
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	paymentRequired := 0
	for status := range results {
		switch status {
		case http.StatusOK:
			successes++
		case http.StatusPaymentRequired:
			paymentRequired++
		default:
			t.Fatalf("unexpected status %d", status)
		}
	}
	if successes != allowed {
		t.Fatalf("successes=%d want %d; paymentRequired=%d", successes, allowed, paymentRequired)
	}
	assertBalance(t, h.db, h.userID, "0.000000")
	if negativeBalance(t, h.db, h.userID) {
		t.Fatal("balance went negative")
	}
	t.Logf("concurrency assertion: successes=%d payment_required=%d final_balance=0.000000", successes, paymentRequired)
}

func newBillingHarness(t *testing.T, initialBalance float64, concurrencyLimit int) *billingHarness {
	t.Helper()
	ctx := context.Background()
	port := freePostgresPort(t)
	dataDir := t.TempDir()
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(uint32(port)).
		Database("postgres").
		Username("postgres").
		Password("postgres").
		RuntimePath(filepath.Join(dataDir, "runtime")).
		DataPath(filepath.Join(dataDir, "data")).
		BinariesPath(filepath.Join(dataDir, "bin")))
	if err := pg.Start(); err != nil {
		t.Fatalf("start embedded postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Stop() })

	dsn := fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect embedded postgres: %v", err)
	}
	t.Cleanup(db.Close)

	if err := bootstrap.Migrate(ctx, db, "../../migrations"); err != nil {
		t.Fatalf("migrate embedded postgres: %v", err)
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"choices": []map[string]any{{"message": map[string]any{"role": "assistant", "content": "ok"}, "finish_reason": "stop"}},
			"usage":   map[string]any{"prompt_tokens": 100, "completion_tokens": 200, "total_tokens": 300},
		})
	}))
	t.Cleanup(upstream.Close)

	cfg := config.Config{
		AppEnv:                       "test",
		AppPort:                      "0",
		AppPublicURL:                 "http://localhost",
		DatabaseURL:                  dsn,
		RedisURL:                     "redis://" + mr.Addr() + "/0",
		JWTSecret:                    "integration-secret",
		KeyEncryptionSecret:          "integration-key",
		DefaultRateMultiplier:        "1.3",
		APIKeyPrefix:                 "lsk_test_",
		DefaultUserRPMLimit:          1000,
		DefaultUserConcurrencyLimit:  concurrencyLimit,
		DefaultGatewayTimeoutSeconds: 30,
		AdminUser:                    "admin",
		AdminPass:                    "password123",
	}
	h := &billingHarness{t: t, ctx: ctx, db: db, redis: redisClient, upstream: upstream, cfg: cfg}
	h.seed(t, initialBalance, concurrencyLimit)
	h.handler = server.New(cfg, db, redisClient)
	return h
}

func (h *billingHarness) seed(t *testing.T, initialBalance float64, concurrencyLimit int) {
	t.Helper()
	ctx := h.ctx
	userHash, err := password.Hash("password123")
	if err != nil {
		t.Fatal(err)
	}
	adminHash, err := password.Hash("password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, status, balance)
		VALUES ('money-user', 'money-user@local', $1, 'user', 'active', $2::numeric)
		RETURNING id::text
	`, userHash, fixed(initialBalance)).Scan(&h.userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := h.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, status, balance)
		VALUES ('money-admin', 'money-admin@local', $1, 'admin', 'active', 0)
		RETURNING id::text
	`, adminHash).Scan(&h.adminID); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	h.apiKey = "lsk_test_plaintext_money_key"
	if _, err := h.db.Exec(ctx, `
		INSERT INTO api_keys (user_id, key_prefix, key_hash, name, status, rpm_limit, concurrency_limit)
		VALUES ($1, $2, $3, 'integration', 'active', 1000, $4)
	`, h.userID, apikey.Mask(h.apiKey), apikey.Hash(h.apiKey), concurrencyLimit); err != nil {
		t.Fatalf("seed api key: %v", err)
	}
	var tokenModelID, imageModelID, channelID string
	if err := h.db.QueryRow(ctx, `
		INSERT INTO models (public_name, type, model_group, billing_mode, input_price_per_1k, output_price_per_1k, price_per_call, rate_multiplier, status)
		VALUES ('public-token', 'chat', 'test', 'token', 0.100000, 0.200000, 0, 1.300, 'enabled')
		RETURNING id::text
	`).Scan(&tokenModelID); err != nil {
		t.Fatalf("seed token model: %v", err)
	}
	if err := h.db.QueryRow(ctx, `
		INSERT INTO models (public_name, type, model_group, billing_mode, input_price_per_1k, output_price_per_1k, price_per_call, rate_multiplier, status)
		VALUES ('public-image', 'image', 'test', 'per_call', 0, 0, 0.020000, 1.300, 'enabled')
		RETURNING id::text
	`).Scan(&imageModelID); err != nil {
		t.Fatalf("seed image model: %v", err)
	}
	if err := h.db.QueryRow(ctx, `
		INSERT INTO upstream_channels (name, provider_type, base_url, api_key_encrypted, status, weight, timeout_seconds, rpm_limit, concurrency_limit, fail_threshold)
		VALUES ('local-upstream', 'openai', $1, $2, 'enabled', 1, 10, 1000, 100, 5)
		RETURNING id::text
	`, h.upstream.URL, base64.StdEncoding.EncodeToString([]byte("upstream-key"))).Scan(&channelID); err != nil {
		t.Fatalf("seed channel: %v", err)
	}
	for _, pair := range []struct{ modelID, upstream string }{{tokenModelID, "upstream-token"}, {imageModelID, "upstream-image"}} {
		if _, err := h.db.Exec(ctx, `
			INSERT INTO channel_models (channel_id, model_id, upstream_model_name, status)
			VALUES ($1, $2, $3, 'enabled')
		`, channelID, pair.modelID, pair.upstream); err != nil {
			t.Fatalf("seed channel model: %v", err)
		}
	}
}

func (h *billingHarness) chat(t *testing.T, key, model string, n int) (int, string) {
	t.Helper()
	body := fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":"hello"}],"max_tokens":200,"n":%d}`, model, n)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func assertMoneyPath(t *testing.T, db *pgxpool.Pool, userID, publicName, baseCost, multiplier, charge string) {
	t.Helper()
	var requestBase, requestMultiplier, requestCharge, ledgerBase, ledgerMultiplier, ledgerCharge string
	err := db.QueryRow(context.Background(), `
		SELECT gr.base_cost::numeric(18,6)::text, gr.rate_multiplier::numeric(6,3)::text, gr.charge::numeric(18,6)::text,
		       bl.base_cost::numeric(18,6)::text, bl.rate_multiplier::numeric(6,3)::text, (-bl.amount)::numeric(18,6)::text
		FROM gateway_requests gr
		JOIN models m ON m.id = gr.model_id
		JOIN balance_ledger bl ON bl.related_id = gr.id
		WHERE gr.user_id=$1 AND m.public_name=$2
		ORDER BY gr.created_at DESC
		LIMIT 1
	`, userID, publicName).Scan(&requestBase, &requestMultiplier, &requestCharge, &ledgerBase, &ledgerMultiplier, &ledgerCharge)
	if err != nil {
		t.Fatalf("query money path: %v", err)
	}
	if requestBase != baseCost || ledgerBase != baseCost || requestMultiplier != multiplier || ledgerMultiplier != multiplier || requestCharge != charge || ledgerCharge != charge {
		t.Fatalf("money path mismatch: request=(%s,%s,%s) ledger=(%s,%s,%s) want=(%s,%s,%s)", requestBase, requestMultiplier, requestCharge, ledgerBase, ledgerMultiplier, ledgerCharge, baseCost, multiplier, charge)
	}
	t.Logf("money path %s: base_cost=%s rate_multiplier=%s charge=%s", publicName, baseCost, multiplier, charge)
}

func assertBalance(t *testing.T, db *pgxpool.Pool, userID, want string) {
	t.Helper()
	var got string
	if err := db.QueryRow(context.Background(), "SELECT balance::numeric(18,6)::text FROM users WHERE id=$1", userID).Scan(&got); err != nil {
		t.Fatalf("query balance: %v", err)
	}
	if got != want {
		t.Fatalf("balance=%s want %s", got, want)
	}
	t.Logf("balance assertion: user=%s balance=%s", userID, got)
}

func assertLedgerType(t *testing.T, db *pgxpool.Pool, userID, ledgerType, amount, before, after string) {
	t.Helper()
	var gotAmount, gotBefore, gotAfter string
	if err := db.QueryRow(context.Background(), `
		SELECT amount::numeric(18,6)::text, balance_before::numeric(18,6)::text, balance_after::numeric(18,6)::text
		FROM balance_ledger
		WHERE user_id=$1 AND type=$2
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, ledgerType).Scan(&gotAmount, &gotBefore, &gotAfter); err != nil {
		t.Fatalf("query ledger %s: %v", ledgerType, err)
	}
	if gotAmount != amount || gotBefore != before || gotAfter != after {
		t.Fatalf("ledger %s=(%s,%s,%s) want=(%s,%s,%s)", ledgerType, gotAmount, gotBefore, gotAfter, amount, before, after)
	}
	t.Logf("ledger assertion: type=%s amount=%s before=%s after=%s", ledgerType, amount, before, after)
}

func countBillingRows(t *testing.T, db *pgxpool.Pool, userID string) (int, int) {
	t.Helper()
	var ledger, requests int
	if err := db.QueryRow(context.Background(), "SELECT count(*) FROM balance_ledger WHERE user_id=$1", userID).Scan(&ledger); err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if err := db.QueryRow(context.Background(), "SELECT count(*) FROM gateway_requests WHERE user_id=$1", userID).Scan(&requests); err != nil {
		t.Fatalf("count requests: %v", err)
	}
	return ledger, requests
}

func negativeBalance(t *testing.T, db *pgxpool.Pool, userID string) bool {
	t.Helper()
	var negative bool
	if err := db.QueryRow(context.Background(), "SELECT balance < 0 FROM users WHERE id=$1", userID).Scan(&negative); err != nil {
		t.Fatalf("query negative balance: %v", err)
	}
	return negative
}

func mustSignAdmin(t *testing.T, cfg config.Config, adminID string) string {
	t.Helper()
	raw, err := token.Sign(cfg.JWTSecret, adminID, "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func freePostgresPort(t *testing.T) int {
	t.Helper()
	listener := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer listener.Close()
	parts := strings.Split(listener.Listener.Addr().String(), ":")
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		t.Fatalf("parse free port: %v", err)
	}
	return port
}

func fixed(value float64) string {
	return fmt.Sprintf("%.6f", value)
}

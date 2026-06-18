package job

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"lingshu/backend/internal/bootstrap"
	"lingshu/backend/internal/pkg/password"
)

func TestCleanerDeletesExpiredGatewayRequestsButKeepsLedger(t *testing.T) {
	ctx := context.Background()
	db := newCleanerTestDB(t, ctx)
	userID, apiKeyID, modelID, channelID := seedCleanupRefs(t, ctx, db)

	var relatedRequestID string
	for i := 0; i < 100; i++ {
		var requestID string
		if err := db.QueryRow(ctx, `
			INSERT INTO gateway_requests (
				request_id, user_id, api_key_id, model_id, channel_id, endpoint, status, http_status,
				prompt_tokens, completion_tokens, total_tokens, base_cost, rate_multiplier, charge, created_at
			)
			VALUES ($1, $2, $3, $4, $5, '/v1/chat/completions', 'success', 200, 1, 1, 2, 0.010000, 1.200, 0.012000, now() - interval '31 days')
			RETURNING id::text
		`, fmt.Sprintf("old-%03d", i), userID, apiKeyID, modelID, channelID).Scan(&requestID); err != nil {
			t.Fatalf("insert old gateway request: %v", err)
		}
		if i == 0 {
			relatedRequestID = requestID
		}
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO gateway_requests (
			request_id, user_id, api_key_id, model_id, channel_id, endpoint, status, http_status,
			prompt_tokens, completion_tokens, total_tokens, base_cost, rate_multiplier, charge, created_at
		)
		VALUES ('recent', $1, $2, $3, $4, '/v1/chat/completions', 'success', 200, 1, 1, 2, 0.010000, 1.200, 0.012000, now())
	`, userID, apiKeyID, modelID, channelID); err != nil {
		t.Fatalf("insert recent gateway request: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO balance_ledger (
			user_id, type, amount, balance_before, balance_after, base_cost, rate_multiplier, related_type, related_id, remark
		)
		VALUES ($1, 'usage_charge', -0.012000, 1.000000, 0.988000, 0.010000, 1.200, 'gateway_request', $2::uuid, 'must stay')
	`, userID, relatedRequestID); err != nil {
		t.Fatalf("insert ledger: %v", err)
	}

	results := NewCleaner(db, nil, CleanerConfig{LogRetentionDays: 30}).Run(ctx)
	for _, result := range results {
		if result.Err != "" {
			t.Fatalf("cleanup %s returned error: %s", result.Table, result.Err)
		}
	}
	if got := resultDeleted(results, "gateway_requests"); got != 100 {
		t.Fatalf("gateway_requests deleted=%d, want 100", got)
	}
	assertCount(t, ctx, db, "gateway_requests", 1)
	assertCount(t, ctx, db, "balance_ledger", 1)
}

func newCleanerTestDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	port := freeCleanerPostgresPort(t)
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
	return db
}

func seedCleanupRefs(t *testing.T, ctx context.Context, db *pgxpool.Pool) (string, string, string, string) {
	t.Helper()
	hash, err := password.Hash("password123")
	if err != nil {
		t.Fatal(err)
	}
	var userID, apiKeyID, modelID, channelID string
	if err := db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, status, balance)
		VALUES ('cleanup-user', 'cleanup@local', $1, 'user', 'active', 1.000000)
		RETURNING id::text
	`, hash).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.QueryRow(ctx, `
		INSERT INTO api_keys (user_id, key_prefix, key_hash, name, status)
		VALUES ($1, 'lsk_test_', 'hash', 'cleanup', 'active')
		RETURNING id::text
	`, userID).Scan(&apiKeyID); err != nil {
		t.Fatalf("insert api key: %v", err)
	}
	if err := db.QueryRow(ctx, `
		INSERT INTO models (public_name, type, billing_mode, rate_multiplier)
		VALUES ('cleanup-model', 'chat', 'token', 1.200)
		RETURNING id::text
	`).Scan(&modelID); err != nil {
		t.Fatalf("insert model: %v", err)
	}
	if err := db.QueryRow(ctx, `
		INSERT INTO upstream_channels (name, provider_type, base_url, api_key_encrypted, status)
		VALUES ('cleanup-channel', 'openai', 'http://127.0.0.1', 'secret', 'enabled')
		RETURNING id::text
	`).Scan(&channelID); err != nil {
		t.Fatalf("insert channel: %v", err)
	}
	return userID, apiKeyID, modelID, channelID
}

func resultDeleted(results []CleanupResult, table string) int64 {
	for _, item := range results {
		if item.Table == table {
			return item.Deleted
		}
	}
	return -1
}

func assertCount(t *testing.T, ctx context.Context, db *pgxpool.Pool, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(ctx, "SELECT count(*)::int FROM "+table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s count=%d, want %d", table, got, want)
	}
}

func freeCleanerPostgresPort(t *testing.T) int {
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

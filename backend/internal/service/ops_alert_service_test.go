package service

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
	"lingshu/backend/internal/repository"
)

func TestOpsAlertDisablesChannelAfterConsecutiveFailures(t *testing.T) {
	ctx := context.Background()
	db := newServiceTestDB(t, ctx)
	settings := repository.NewSettingsRepository(db)
	audits := repository.NewAuditRepository(db)
	notification := NewNotificationService(settings)
	alerts := NewOpsAlertService(db, settings, audits, notification)
	channelID := seedOpsAlertChannel(t, ctx, db, 5)

	if _, err := db.Exec(ctx, `
		UPDATE system_settings
		SET value='true'
		WHERE key='alert_enabled'
	`); err != nil {
		t.Fatalf("enable alerts: %v", err)
	}

	if err := alerts.Evaluate(ctx); err != nil {
		t.Fatalf("evaluate alerts: %v", err)
	}

	var status string
	if err := db.QueryRow(ctx, "SELECT status FROM upstream_channels WHERE id=$1", channelID).Scan(&status); err != nil {
		t.Fatalf("query channel status: %v", err)
	}
	if status != "disabled" {
		t.Fatalf("channel status=%q want disabled", status)
	}
	var count int
	if err := db.QueryRow(ctx, "SELECT count(*)::int FROM ops_alerts WHERE rule_key='channel_consecutive_failures' AND status='active'").Scan(&count); err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if count != 1 {
		t.Fatalf("active channel alerts=%d want 1", count)
	}
}

func newServiceTestDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	port := freeOpsAlertPostgresPort(t)
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

func seedOpsAlertChannel(t *testing.T, ctx context.Context, db *pgxpool.Pool, failCount int) string {
	t.Helper()
	hash, err := password.Hash("password123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO users (username, email, password_hash, role, status, balance)
		VALUES ('ops-alert-user', 'ops-alert@local', $1, 'admin', 'active', 10)
	`, hash); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var channelID string
	if err := db.QueryRow(ctx, `
		INSERT INTO upstream_channels (name, provider_type, base_url, api_key_encrypted, status, fail_count, health)
		VALUES ('ops-alert-channel', 'openai', 'http://127.0.0.1', 'secret', 'enabled', $1, 'unhealthy')
		RETURNING id::text
	`, failCount).Scan(&channelID); err != nil {
		t.Fatalf("insert channel: %v", err)
	}
	return channelID
}

func freeOpsAlertPostgresPort(t *testing.T) int {
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

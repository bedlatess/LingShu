package repository_test

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"lingshu/backend/internal/bootstrap"
	"lingshu/backend/internal/repository"
)

func TestChannelListPagedCountsEnabledBindings(t *testing.T) {
	ctx := context.Background()
	db := newRepositoryTestDB(t)
	repo := repository.NewChannelRepository(db)

	channel, err := repo.Create(ctx, repository.ChannelInput{
		Name:             "count-channel",
		ProviderType:     "openai",
		BaseURL:          "http://upstream.test",
		Status:           "enabled",
		Weight:           1,
		TimeoutSeconds:   120,
		RPMLimit:         60,
		ConcurrencyLimit: 5,
		FailThreshold:    5,
	}, "secret")
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	modelIDs := []string{}
	for _, name := range []string{"count-a", "count-b", "count-disabled"} {
		var modelID string
		if err := db.QueryRow(ctx, `
			INSERT INTO models (public_name, type, model_group, billing_mode, input_price_per_1k, output_price_per_1k, price_per_call, rate_multiplier, status)
			VALUES ($1, 'chat', '', 'token', 0, 0, 0, 1.200, 'enabled')
			RETURNING id::text
		`, name).Scan(&modelID); err != nil {
			t.Fatalf("insert model %s: %v", name, err)
		}
		modelIDs = append(modelIDs, modelID)
	}
	for i, modelID := range modelIDs {
		status := "enabled"
		if i == 2 {
			status = "disabled"
		}
		if _, err := db.Exec(ctx, `
			INSERT INTO channel_models (channel_id, model_id, upstream_model_name, status)
			VALUES ($1, $2, $3, $4)
		`, channel.ID, modelID, "upstream-"+modelID, status); err != nil {
			t.Fatalf("insert binding: %v", err)
		}
	}

	items, _, err := repo.ListPaged(ctx, 20, 0)
	if err != nil {
		t.Fatalf("list channels: %v", err)
	}
	for _, item := range items {
		if item.ID == channel.ID {
			if item.BoundCount != 2 {
				t.Fatalf("bound_count=%d want 2", item.BoundCount)
			}
			return
		}
	}
	t.Fatalf("created channel %s not found", channel.ID)
}

func TestChannelDeleteRemovesModelBindings(t *testing.T) {
	ctx := context.Background()
	db := newRepositoryTestDB(t)
	repo := repository.NewChannelRepository(db)

	channel, err := repo.Create(ctx, repository.ChannelInput{
		Name:             "delete-channel",
		ProviderType:     "openai",
		BaseURL:          "http://upstream.test",
		Status:           "enabled",
		Weight:           1,
		TimeoutSeconds:   120,
		RPMLimit:         60,
		ConcurrencyLimit: 5,
		FailThreshold:    5,
	}, "secret")
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	var modelID string
	if err := db.QueryRow(ctx, `
		INSERT INTO models (public_name, type, model_group, billing_mode, input_price_per_1k, output_price_per_1k, price_per_call, rate_multiplier, status)
		VALUES ('delete-bound-model', 'chat', '', 'token', 0, 0, 0, 1.200, 'enabled')
		RETURNING id::text
	`).Scan(&modelID); err != nil {
		t.Fatalf("insert model: %v", err)
	}
	if _, err := db.Exec(ctx, `
		INSERT INTO channel_models (channel_id, model_id, upstream_model_name, status)
		VALUES ($1, $2, 'upstream-delete-bound-model', 'enabled')
	`, channel.ID, modelID); err != nil {
		t.Fatalf("insert binding: %v", err)
	}

	if err := repo.Delete(ctx, channel.ID); err != nil {
		t.Fatalf("delete channel: %v", err)
	}
	var bindings int
	if err := db.QueryRow(ctx, `SELECT count(*)::int FROM channel_models WHERE channel_id=$1`, channel.ID).Scan(&bindings); err != nil {
		t.Fatalf("count bindings: %v", err)
	}
	if bindings != 0 {
		t.Fatalf("bindings=%d want 0", bindings)
	}
}

func TestModelListPagedSortsHealthyBoundModelsFirst(t *testing.T) {
	ctx := context.Background()
	db := newRepositoryTestDB(t)
	channelRepo := repository.NewChannelRepository(db)
	modelRepo := repository.NewModelRepository(db)

	healthyChannel, err := channelRepo.Create(ctx, repository.ChannelInput{
		Name:             "healthy-channel",
		ProviderType:     "openai",
		BaseURL:          "http://upstream.test",
		Status:           "enabled",
		Weight:           1,
		TimeoutSeconds:   120,
		RPMLimit:         60,
		ConcurrencyLimit: 5,
		FailThreshold:    5,
	}, "secret")
	if err != nil {
		t.Fatalf("create healthy channel: %v", err)
	}
	unhealthyChannel, err := channelRepo.Create(ctx, repository.ChannelInput{
		Name:             "unhealthy-channel",
		ProviderType:     "openai",
		BaseURL:          "http://upstream.test",
		Status:           "enabled",
		Weight:           1,
		TimeoutSeconds:   120,
		RPMLimit:         60,
		ConcurrencyLimit: 5,
		FailThreshold:    5,
	}, "secret")
	if err != nil {
		t.Fatalf("create unhealthy channel: %v", err)
	}
	if _, err := db.Exec(ctx, `UPDATE upstream_channels SET health='unhealthy' WHERE id=$1`, unhealthyChannel.ID); err != nil {
		t.Fatalf("mark unhealthy: %v", err)
	}

	healthyModel := insertRepositoryModel(t, db, "healthy-bound-model", 20)
	unhealthyModel := insertRepositoryModel(t, db, "unhealthy-bound-model", 1)
	if _, err := db.Exec(ctx, `
		INSERT INTO channel_models (channel_id, model_id, upstream_model_name, status)
		VALUES ($1, $2, 'healthy-upstream', 'enabled'), ($3, $4, 'unhealthy-upstream', 'enabled')
	`, healthyChannel.ID, healthyModel, unhealthyChannel.ID, unhealthyModel); err != nil {
		t.Fatalf("insert bindings: %v", err)
	}

	items, _, err := modelRepo.ListPaged(ctx, 20, 0)
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("items len=%d want at least 2", len(items))
	}
	if items[0].ID != healthyModel {
		t.Fatalf("first model=%s want healthy model %s", items[0].ID, healthyModel)
	}
	if !items[0].HasHealthyChannel {
		t.Fatalf("healthy model HasHealthyChannel=false")
	}
	for _, item := range items {
		if item.ID == unhealthyModel && item.HasHealthyChannel {
			t.Fatalf("unhealthy-bound model HasHealthyChannel=true")
		}
	}
}

func insertRepositoryModel(t *testing.T, db *pgxpool.Pool, name string, sortOrder int) string {
	t.Helper()
	var modelID string
	if err := db.QueryRow(context.Background(), `
		INSERT INTO models (public_name, type, model_group, billing_mode, input_price_per_1k, output_price_per_1k, price_per_call, rate_multiplier, status, sort_order)
		VALUES ($1, 'chat', '', 'token', 0, 0, 0, 1.200, 'enabled', $2)
		RETURNING id::text
	`, name, sortOrder).Scan(&modelID); err != nil {
		t.Fatalf("insert model %s: %v", name, err)
	}
	return modelID
}

func newRepositoryTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	port := freeRepositoryPostgresPort(t)
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

func freeRepositoryPostgresPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate postgres port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

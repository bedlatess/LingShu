package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Channel struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	ProviderType     string     `json:"provider_type"`
	BaseURL          string     `json:"base_url"`
	Status           string     `json:"status"`
	Weight           int        `json:"weight"`
	TimeoutSeconds   int        `json:"timeout_seconds"`
	RPMLimit         int        `json:"rpm_limit"`
	ConcurrencyLimit int        `json:"concurrency_limit"`
	FailThreshold    int        `json:"fail_threshold"`
	FailCount        int        `json:"fail_count"`
	Health           string     `json:"health"`
	LastSuccessAt    *time.Time `json:"last_success_at,omitempty"`
	LastErrorAt      *time.Time `json:"last_error_at,omitempty"`
	LastErrorMessage string     `json:"last_error_message"`
	LastLatencyMS    int        `json:"last_latency_ms"`
	BoundCount       int        `json:"bound_count"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ChannelSecret struct {
	ID              string
	ProviderType    string
	BaseURL         string
	APIKeyEncrypted string
	TimeoutSeconds  int
}

type ChannelInput struct {
	Name             string `json:"name"`
	ProviderType     string `json:"provider_type"`
	BaseURL          string `json:"base_url"`
	APIKey           string `json:"api_key"`
	Status           string `json:"status"`
	Weight           int    `json:"weight"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	RPMLimit         int    `json:"rpm_limit"`
	ConcurrencyLimit int    `json:"concurrency_limit"`
	FailThreshold    int    `json:"fail_threshold"`
}

type ChannelRepository struct {
	db *pgxpool.Pool
}

type ChannelDetailStats struct {
	Requests       int    `json:"requests"`
	Successes      int    `json:"successes"`
	Failures       int    `json:"failures"`
	AverageLatency string `json:"average_latency"`
}

type ChannelDetailBinding struct {
	ID                string    `json:"id"`
	ModelID           string    `json:"model_id"`
	ModelName         string    `json:"model_name"`
	UpstreamModelName string    `json:"upstream_model_name"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type ChannelDetail struct {
	Channel Channel                `json:"channel"`
	Models  []ChannelDetailBinding `json:"models"`
	Stats   ChannelDetailStats     `json:"stats"`
}

type ChannelModelBinding struct {
	ID                string    `json:"id"`
	ChannelID         string    `json:"channel_id"`
	ModelID           string    `json:"model_id"`
	UpstreamModelName string    `json:"upstream_model_name"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type BindChannelModelInput struct {
	ChannelID         string `json:"channel_id"`
	ModelID           string `json:"model_id"`
	UpstreamModelName string `json:"upstream_model_name"`
}

type ImportChannelModelInput struct {
	UpstreamName     string `json:"upstream_name"`
	PublicName       string `json:"public_name"`
	Type             string `json:"type"`
	BillingMode      string `json:"billing_mode"`
	InputPricePer1K  string `json:"input_price_per_1k"`
	OutputPricePer1K string `json:"output_price_per_1k"`
	PricePerCall     string `json:"price_per_call"`
	RateMultiplier   string `json:"rate_multiplier"`
	Status           string `json:"status"`
	SortOrder        int    `json:"sort_order"`
	BindExistingOnly bool   `json:"-"`
}

type ImportChannelModelResult struct {
	ModelID           string `json:"model_id"`
	PublicName        string `json:"public_name"`
	UpstreamModelName string `json:"upstream_model_name"`
	BindingID         string `json:"binding_id"`
	Created           bool   `json:"created"`
	Bound             bool   `json:"bound"`
}

func NewChannelRepository(db *pgxpool.Pool) ChannelRepository {
	return ChannelRepository{db: db}
}

func (r ChannelRepository) List(ctx context.Context) ([]Channel, error) {
	items, _, err := r.ListPaged(ctx, 100, 0)
	return items, err
}

func (r ChannelRepository) ListPaged(ctx context.Context, limit, offset int) ([]Channel, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM upstream_channels WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT c.id::text, c.name, c.provider_type, c.base_url, c.status, c.weight, c.timeout_seconds,
		       c.rpm_limit, c.concurrency_limit, c.fail_threshold, c.fail_count, c.health,
		       c.last_success_at, c.last_error_at, COALESCE(c.last_error_message, ''), c.last_latency_ms,
		       COALESCE(b.bound_count, 0)::int, c.created_at, c.updated_at
		FROM upstream_channels c
		LEFT JOIN (
			SELECT channel_id, COUNT(*)::int AS bound_count
			FROM channel_models
			WHERE status='enabled'
			GROUP BY channel_id
		) AS b ON b.channel_id = c.id
		WHERE c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []Channel{}
	for rows.Next() {
		item, err := scanChannel(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r ChannelRepository) FindByID(ctx context.Context, id string) (Channel, error) {
	var item Channel
	err := r.db.QueryRow(ctx, `
		SELECT id::text, name, provider_type, base_url, status, weight, timeout_seconds,
		       rpm_limit, concurrency_limit, fail_threshold, fail_count, health,
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), last_latency_ms, 0::int, created_at, updated_at
		FROM upstream_channels
		WHERE id=$1 AND deleted_at IS NULL
	`, id).Scan(
		&item.ID, &item.Name, &item.ProviderType, &item.BaseURL, &item.Status,
		&item.Weight, &item.TimeoutSeconds, &item.RPMLimit, &item.ConcurrencyLimit,
		&item.FailThreshold, &item.FailCount, &item.Health, &item.LastSuccessAt,
		&item.LastErrorAt, &item.LastErrorMessage, &item.LastLatencyMS, &item.BoundCount, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func (r ChannelRepository) FindSecretByID(ctx context.Context, id string) (ChannelSecret, error) {
	var item ChannelSecret
	err := r.db.QueryRow(ctx, `
		SELECT id::text, provider_type, base_url, api_key_encrypted, timeout_seconds
		FROM upstream_channels
		WHERE id=$1 AND deleted_at IS NULL
	`, id).Scan(&item.ID, &item.ProviderType, &item.BaseURL, &item.APIKeyEncrypted, &item.TimeoutSeconds)
	return item, err
}

func (r ChannelRepository) Detail(ctx context.Context, id string) (ChannelDetail, error) {
	channel, err := r.FindByID(ctx, id)
	if err != nil {
		return ChannelDetail{}, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT cm.id::text, cm.model_id::text, m.public_name, cm.upstream_model_name, cm.status, cm.created_at
		FROM channel_models cm
		JOIN models m ON m.id = cm.model_id AND m.deleted_at IS NULL
		WHERE cm.channel_id=$1 AND cm.status='enabled'
		ORDER BY cm.created_at DESC
	`, id)
	if err != nil {
		return ChannelDetail{}, err
	}
	defer rows.Close()
	models := []ChannelDetailBinding{}
	for rows.Next() {
		var item ChannelDetailBinding
		if err := rows.Scan(&item.ID, &item.ModelID, &item.ModelName, &item.UpstreamModelName, &item.Status, &item.CreatedAt); err != nil {
			return ChannelDetail{}, err
		}
		models = append(models, item)
	}
	if err := rows.Err(); err != nil {
		return ChannelDetail{}, err
	}
	var stats ChannelDetailStats
	if err := r.db.QueryRow(ctx, `
		SELECT count(*)::int,
		       count(*) FILTER (WHERE status='success')::int,
		       count(*) FILTER (WHERE status!='success')::int,
		       COALESCE(avg(latency_ms),0)::text
		FROM gateway_requests
		WHERE channel_id=$1
	`, id).Scan(&stats.Requests, &stats.Successes, &stats.Failures, &stats.AverageLatency); err != nil {
		return ChannelDetail{}, err
	}
	return ChannelDetail{Channel: channel, Models: models, Stats: stats}, nil
}

func (r ChannelRepository) Create(ctx context.Context, input ChannelInput, encryptedKey string) (Channel, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO upstream_channels (
			name, provider_type, base_url, api_key_encrypted, status, weight,
			timeout_seconds, rpm_limit, concurrency_limit, fail_threshold
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id::text, name, provider_type, base_url, status, weight, timeout_seconds,
		       rpm_limit, concurrency_limit, fail_threshold, fail_count, health,
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), last_latency_ms, 0::int, created_at, updated_at
	`, input.Name, input.ProviderType, input.BaseURL, encryptedKey, input.Status, input.Weight, input.TimeoutSeconds, input.RPMLimit, input.ConcurrencyLimit, input.FailThreshold)
	return scanChannel(row)
}

func (r ChannelRepository) Update(ctx context.Context, id string, input ChannelInput, encryptedKey string) (Channel, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE upstream_channels
		SET name=$2, provider_type=$3, base_url=$4,
		    api_key_encrypted=CASE WHEN $5='' THEN api_key_encrypted ELSE $5 END,
		    status=$6, weight=$7, timeout_seconds=$8, rpm_limit=$9,
		    concurrency_limit=$10, fail_threshold=$11, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING id::text, name, provider_type, base_url, status, weight, timeout_seconds,
		       rpm_limit, concurrency_limit, fail_threshold, fail_count, health,
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), last_latency_ms, 0::int, created_at, updated_at
	`, id, input.Name, input.ProviderType, input.BaseURL, encryptedKey, input.Status, input.Weight, input.TimeoutSeconds, input.RPMLimit, input.ConcurrencyLimit, input.FailThreshold)
	return scanChannel(row)
}

func (r ChannelRepository) Disable(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE upstream_channels SET status='disabled', updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r ChannelRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE upstream_channels SET deleted_at=now(), status='disabled', updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("channel not found")
	}
	return nil
}

func (r ChannelRepository) MarkTest(ctx context.Context, id string, ok bool, message string, latencyMS int64) error {
	if ok {
		_, err := r.db.Exec(ctx, "UPDATE upstream_channels SET health='healthy', fail_count=0, last_success_at=now(), last_error_message=NULL, last_latency_ms=$2 WHERE id=$1 AND deleted_at IS NULL", id, latencyMS)
		return err
	}
	_, err := r.db.Exec(ctx, "UPDATE upstream_channels SET health='unhealthy', fail_count=fail_count+1, last_error_at=now(), last_error_message=$2, last_latency_ms=$3 WHERE id=$1 AND deleted_at IS NULL", id, message, latencyMS)
	return err
}

func (r ChannelRepository) BindModel(ctx context.Context, input BindChannelModelInput) (ChannelModelBinding, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO channel_models (channel_id, model_id, upstream_model_name, status)
		VALUES ($1, $2, $3, 'enabled')
		ON CONFLICT (channel_id, model_id)
		DO UPDATE SET upstream_model_name=EXCLUDED.upstream_model_name, status='enabled', updated_at=now()
		RETURNING id::text, channel_id::text, model_id::text, upstream_model_name, status, created_at
	`, input.ChannelID, input.ModelID, input.UpstreamModelName)
	var item ChannelModelBinding
	err := row.Scan(&item.ID, &item.ChannelID, &item.ModelID, &item.UpstreamModelName, &item.Status, &item.CreatedAt)
	return item, err
}

func (r ChannelRepository) ImportModels(ctx context.Context, channelID string, models []ImportChannelModelInput) ([]ImportChannelModelResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	results := make([]ImportChannelModelResult, 0, len(models))
	for _, input := range models {
		var modelID string
		var created bool
		err := tx.QueryRow(ctx, `
			SELECT id::text
			FROM models
			WHERE public_name=$1 AND deleted_at IS NULL
		`, input.PublicName).Scan(&modelID)
		if errors.Is(err, pgx.ErrNoRows) && input.BindExistingOnly {
			return nil, errors.New("model not found: " + input.PublicName)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			created = true
			err = tx.QueryRow(ctx, `
				INSERT INTO models (
					public_name, type, model_group, billing_mode,
					input_price_per_1k, output_price_per_1k, price_per_call,
					rate_multiplier, status, sort_order
				)
				VALUES ($1,$2,'',$3,$4::numeric,$5::numeric,$6::numeric,$7::numeric,$8,$9)
				RETURNING id::text
			`, input.PublicName, input.Type, input.BillingMode, input.InputPricePer1K, input.OutputPricePer1K, input.PricePerCall, input.RateMultiplier, input.Status, input.SortOrder).Scan(&modelID)
		}
		if err != nil {
			return nil, err
		}

		var bindingID string
		var bindingStatus string
		if err := tx.QueryRow(ctx, `
			INSERT INTO channel_models (channel_id, model_id, upstream_model_name, status)
			VALUES ($1, $2, $3, 'enabled')
			ON CONFLICT (channel_id, model_id)
			DO UPDATE SET upstream_model_name=EXCLUDED.upstream_model_name, status='enabled', updated_at=now()
			RETURNING id::text, status
		`, channelID, modelID, input.UpstreamName).Scan(&bindingID, &bindingStatus); err != nil {
			return nil, err
		}
		results = append(results, ImportChannelModelResult{
			ModelID:           modelID,
			PublicName:        input.PublicName,
			UpstreamModelName: input.UpstreamName,
			BindingID:         bindingID,
			Created:           created,
			Bound:             bindingStatus == "enabled",
		})
	}
	return results, tx.Commit(ctx)
}

func (r ChannelRepository) UnbindModel(ctx context.Context, channelID, modelID string) error {
	tag, err := r.db.Exec(ctx, "UPDATE channel_models SET status='disabled', updated_at=now() WHERE channel_id=$1 AND model_id=$2", channelID, modelID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type channelScanner interface {
	Scan(dest ...any) error
}

func scanChannel(row channelScanner) (Channel, error) {
	var item Channel
	err := row.Scan(
		&item.ID, &item.Name, &item.ProviderType, &item.BaseURL, &item.Status,
		&item.Weight, &item.TimeoutSeconds, &item.RPMLimit, &item.ConcurrencyLimit,
		&item.FailThreshold, &item.FailCount, &item.Health, &item.LastSuccessAt,
		&item.LastErrorAt, &item.LastErrorMessage, &item.LastLatencyMS, &item.BoundCount, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

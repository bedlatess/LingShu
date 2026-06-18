package repository

import (
	"context"
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
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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
	Requests      int    `json:"requests"`
	Successes     int    `json:"successes"`
	Failures      int    `json:"failures"`
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
	Channel  Channel                `json:"channel"`
	Models   []ChannelDetailBinding `json:"models"`
	Stats    ChannelDetailStats     `json:"stats"`
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

func NewChannelRepository(db *pgxpool.Pool) ChannelRepository {
	return ChannelRepository{db: db}
}

func (r ChannelRepository) List(ctx context.Context) ([]Channel, error) {
	items, _, err := r.ListPaged(ctx, 100, 0)
	return items, err
}

func (r ChannelRepository) ListPaged(ctx context.Context, limit, offset int) ([]Channel, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM upstream_channels`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, name, provider_type, base_url, status, weight, timeout_seconds,
		       rpm_limit, concurrency_limit, fail_threshold, fail_count, health,
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), created_at, updated_at
		FROM upstream_channels
		ORDER BY created_at DESC
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
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), created_at, updated_at
		FROM upstream_channels
		WHERE id=$1
	`, id).Scan(
		&item.ID, &item.Name, &item.ProviderType, &item.BaseURL, &item.Status,
		&item.Weight, &item.TimeoutSeconds, &item.RPMLimit, &item.ConcurrencyLimit,
		&item.FailThreshold, &item.FailCount, &item.Health, &item.LastSuccessAt,
		&item.LastErrorAt, &item.LastErrorMessage, &item.CreatedAt, &item.UpdatedAt,
	)
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
		JOIN models m ON m.id = cm.model_id
		WHERE cm.channel_id=$1
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
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), created_at, updated_at
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
		WHERE id=$1
		RETURNING id::text, name, provider_type, base_url, status, weight, timeout_seconds,
		       rpm_limit, concurrency_limit, fail_threshold, fail_count, health,
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), created_at, updated_at
	`, id, input.Name, input.ProviderType, input.BaseURL, encryptedKey, input.Status, input.Weight, input.TimeoutSeconds, input.RPMLimit, input.ConcurrencyLimit, input.FailThreshold)
	return scanChannel(row)
}

func (r ChannelRepository) Disable(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE upstream_channels SET status='disabled', updated_at=now() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r ChannelRepository) Delete(ctx context.Context, id string) error {
	return r.Disable(ctx, id)
}

func (r ChannelRepository) MarkTest(ctx context.Context, id string, ok bool, message string) error {
	if ok {
		_, err := r.db.Exec(ctx, "UPDATE upstream_channels SET health='healthy', fail_count=0, last_success_at=now(), last_error_message=NULL WHERE id=$1", id)
		return err
	}
	_, err := r.db.Exec(ctx, "UPDATE upstream_channels SET health='unhealthy', fail_count=fail_count+1, last_error_at=now(), last_error_message=$2 WHERE id=$1", id, message)
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
		&item.LastErrorAt, &item.LastErrorMessage, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

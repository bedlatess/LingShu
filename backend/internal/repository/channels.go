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
	rows, err := r.db.Query(ctx, `
		SELECT id::text, name, provider_type, base_url, status, weight, timeout_seconds,
		       rpm_limit, concurrency_limit, fail_threshold, fail_count, health,
		       last_success_at, last_error_at, COALESCE(last_error_message, ''), created_at, updated_at
		FROM upstream_channels
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Channel{}
	for rows.Next() {
		item, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

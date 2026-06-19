package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Model struct {
	ID                      string    `json:"id"`
	PublicName              string    `json:"public_name"`
	Type                    string    `json:"type"`
	Group                   string    `json:"group"`
	BillingMode             string    `json:"billing_mode"`
	InputPricePer1K         string    `json:"input_price_per_1k"`
	OutputPricePer1K        string    `json:"output_price_per_1k"`
	CacheCreationPricePer1K string    `json:"cache_creation_price_per_1k"`
	CacheReadPricePer1K     string    `json:"cache_read_price_per_1k"`
	PricePerCall            string    `json:"price_per_call"`
	RateMultiplier          string    `json:"rate_multiplier"`
	SupportsStream          bool      `json:"supports_stream"`
	SupportsTools           bool      `json:"supports_tools"`
	SupportsVision          bool      `json:"supports_vision"`
	Status                  string    `json:"status"`
	SortOrder               int       `json:"sort_order"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type ModelInput struct {
	PublicName              string `json:"public_name"`
	Type                    string `json:"type"`
	Group                   string `json:"group"`
	BillingMode             string `json:"billing_mode"`
	InputPricePer1K         string `json:"input_price_per_1k"`
	OutputPricePer1K        string `json:"output_price_per_1k"`
	CacheCreationPricePer1K string `json:"cache_creation_price_per_1k"`
	CacheReadPricePer1K     string `json:"cache_read_price_per_1k"`
	PricePerCall            string `json:"price_per_call"`
	RateMultiplier          string `json:"rate_multiplier"`
	SupportsStream          bool   `json:"supports_stream"`
	SupportsTools           bool   `json:"supports_tools"`
	SupportsVision          bool   `json:"supports_vision"`
	Status                  string `json:"status"`
	SortOrder               int    `json:"sort_order"`
}

type ModelRepository struct {
	db *pgxpool.Pool
}

type ModelChannelBinding struct {
	ID                string    `json:"id"`
	ChannelID         string    `json:"channel_id"`
	ChannelName       string    `json:"channel_name"`
	ProviderType      string    `json:"provider_type"`
	BaseURL           string    `json:"base_url"`
	UpstreamModelName string    `json:"upstream_model_name"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type ModelDetailStats struct {
	Requests    int    `json:"requests"`
	Successes   int    `json:"successes"`
	Failures    int    `json:"failures"`
	BaseCost    string `json:"base_cost"`
	Charge      string `json:"charge"`
	GrossProfit string `json:"gross_profit"`
}

type ModelDetail struct {
	Model    Model                 `json:"model"`
	Channels []ModelChannelBinding `json:"channels"`
	Stats    ModelDetailStats      `json:"stats"`
}

func NewModelRepository(db *pgxpool.Pool) ModelRepository {
	return ModelRepository{db: db}
}

func (r ModelRepository) List(ctx context.Context) ([]Model, error) {
	items, _, err := r.ListPaged(ctx, 100, 0)
	return items, err
}

func (r ModelRepository) ListVisible(ctx context.Context) ([]Model, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       rate_multiplier::text, supports_stream, supports_tools, supports_vision,
		       status, sort_order, created_at, updated_at
		FROM models
		WHERE status='enabled'
		  AND deleted_at IS NULL
		  AND EXISTS (
		  	SELECT 1
		  	FROM channel_models cm
		  	JOIN upstream_channels c ON c.id = cm.channel_id AND c.deleted_at IS NULL
		  	WHERE cm.model_id=models.id AND cm.status='enabled' AND c.status='enabled'
		  )
		ORDER BY sort_order ASC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Model{}
	for rows.Next() {
		item, err := scanModel(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r ModelRepository) ListPaged(ctx context.Context, limit, offset int) ([]Model, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM models WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       rate_multiplier::text, supports_stream, supports_tools, supports_vision,
		       status, sort_order, created_at, updated_at
		FROM models
		WHERE deleted_at IS NULL
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []Model{}
	for rows.Next() {
		item, err := scanModel(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r ModelRepository) FindByID(ctx context.Context, id string) (Model, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       rate_multiplier::text, supports_stream, supports_tools, supports_vision,
		       status, sort_order, created_at, updated_at
		FROM models
		WHERE id=$1 AND deleted_at IS NULL
	`, id)
	return scanModel(row)
}

func (r ModelRepository) Detail(ctx context.Context, id string) (ModelDetail, error) {
	model, err := r.FindByID(ctx, id)
	if err != nil {
		return ModelDetail{}, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT cm.id::text, cm.channel_id::text, c.name, c.provider_type, c.base_url,
		       cm.upstream_model_name, cm.status, cm.created_at
		FROM channel_models cm
		JOIN upstream_channels c ON c.id = cm.channel_id AND c.deleted_at IS NULL
		WHERE cm.model_id=$1 AND cm.status='enabled'
		ORDER BY cm.created_at DESC
	`, id)
	if err != nil {
		return ModelDetail{}, err
	}
	defer rows.Close()
	channels := []ModelChannelBinding{}
	for rows.Next() {
		var item ModelChannelBinding
		if err := rows.Scan(&item.ID, &item.ChannelID, &item.ChannelName, &item.ProviderType, &item.BaseURL, &item.UpstreamModelName, &item.Status, &item.CreatedAt); err != nil {
			return ModelDetail{}, err
		}
		channels = append(channels, item)
	}
	if err := rows.Err(); err != nil {
		return ModelDetail{}, err
	}

	var stats ModelDetailStats
	if err := r.db.QueryRow(ctx, `
		SELECT count(*)::int,
		       count(*) FILTER (WHERE status='success')::int,
		       count(*) FILTER (WHERE status!='success')::int,
		       COALESCE(sum(base_cost),0)::text,
		       COALESCE(sum(charge),0)::text,
		       COALESCE(sum(charge - base_cost),0)::text
		FROM gateway_requests
		WHERE model_id=$1
	`, id).Scan(&stats.Requests, &stats.Successes, &stats.Failures, &stats.BaseCost, &stats.Charge, &stats.GrossProfit); err != nil {
		return ModelDetail{}, err
	}

	return ModelDetail{Model: model, Channels: channels, Stats: stats}, nil
}

func (r ModelRepository) Create(ctx context.Context, input ModelInput) (Model, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO models (
			public_name, type, model_group, billing_mode,
			input_price_per_1k, output_price_per_1k, price_per_call,
			cache_creation_price_per_1k, cache_read_price_per_1k,
			rate_multiplier, supports_stream, supports_tools, supports_vision,
			status, sort_order
		)
		VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11,$12,$13,$14,$15)
		RETURNING id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       rate_multiplier::text, supports_stream, supports_tools, supports_vision,
		       status, sort_order, created_at, updated_at
	`, input.PublicName, input.Type, input.Group, input.BillingMode, input.InputPricePer1K, input.OutputPricePer1K, input.PricePerCall, input.CacheCreationPricePer1K, input.CacheReadPricePer1K, input.RateMultiplier, input.SupportsStream, input.SupportsTools, input.SupportsVision, input.Status, input.SortOrder)
	return scanModel(row)
}

func (r ModelRepository) Update(ctx context.Context, id string, input ModelInput) (Model, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE models
		SET public_name=$2, type=$3, model_group=$4, billing_mode=$5,
		    input_price_per_1k=$6::numeric, output_price_per_1k=$7::numeric, price_per_call=$8::numeric,
		    cache_creation_price_per_1k=$9::numeric, cache_read_price_per_1k=$10::numeric,
		    rate_multiplier=$11::numeric, supports_stream=$12, supports_tools=$13, supports_vision=$14,
		    status=$15, sort_order=$16, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       rate_multiplier::text, supports_stream, supports_tools, supports_vision,
		       status, sort_order, created_at, updated_at
	`, id, input.PublicName, input.Type, input.Group, input.BillingMode, input.InputPricePer1K, input.OutputPricePer1K, input.PricePerCall, input.CacheCreationPricePer1K, input.CacheReadPricePer1K, input.RateMultiplier, input.SupportsStream, input.SupportsTools, input.SupportsVision, input.Status, input.SortOrder)
	return scanModel(row)
}

func (r ModelRepository) Disable(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE models SET status='disabled', updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r ModelRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE models SET deleted_at=now(), status='disabled', updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("model not found")
	}
	return nil
}

type modelScanner interface {
	Scan(dest ...any) error
}

func scanModel(row modelScanner) (Model, error) {
	var item Model
	err := row.Scan(
		&item.ID,
		&item.PublicName,
		&item.Type,
		&item.Group,
		&item.BillingMode,
		&item.InputPricePer1K,
		&item.OutputPricePer1K,
		&item.PricePerCall,
		&item.CacheCreationPricePer1K,
		&item.CacheReadPricePer1K,
		&item.RateMultiplier,
		&item.SupportsStream,
		&item.SupportsTools,
		&item.SupportsVision,
		&item.Status,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

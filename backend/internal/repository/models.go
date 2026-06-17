package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Model struct {
	ID               string    `json:"id"`
	PublicName       string    `json:"public_name"`
	Type             string    `json:"type"`
	Group            string    `json:"group"`
	BillingMode      string    `json:"billing_mode"`
	InputPricePer1K  string    `json:"input_price_per_1k"`
	OutputPricePer1K string    `json:"output_price_per_1k"`
	PricePerCall     string    `json:"price_per_call"`
	RateMultiplier   string    `json:"rate_multiplier"`
	Status           string    `json:"status"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ModelInput struct {
	PublicName       string `json:"public_name"`
	Type             string `json:"type"`
	Group            string `json:"group"`
	BillingMode      string `json:"billing_mode"`
	InputPricePer1K  string `json:"input_price_per_1k"`
	OutputPricePer1K string `json:"output_price_per_1k"`
	PricePerCall     string `json:"price_per_call"`
	RateMultiplier   string `json:"rate_multiplier"`
	Status           string `json:"status"`
	SortOrder        int    `json:"sort_order"`
}

type ModelRepository struct {
	db *pgxpool.Pool
}

func NewModelRepository(db *pgxpool.Pool) ModelRepository {
	return ModelRepository{db: db}
}

func (r ModelRepository) List(ctx context.Context) ([]Model, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       rate_multiplier::text, status, sort_order, created_at, updated_at
		FROM models
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

func (r ModelRepository) Create(ctx context.Context, input ModelInput) (Model, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO models (
			public_name, type, model_group, billing_mode,
			input_price_per_1k, output_price_per_1k, price_per_call,
			rate_multiplier, status, sort_order
		)
		VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7::numeric,$8::numeric,$9,$10)
		RETURNING id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       rate_multiplier::text, status, sort_order, created_at, updated_at
	`, input.PublicName, input.Type, input.Group, input.BillingMode, input.InputPricePer1K, input.OutputPricePer1K, input.PricePerCall, input.RateMultiplier, input.Status, input.SortOrder)
	return scanModel(row)
}

func (r ModelRepository) Update(ctx context.Context, id string, input ModelInput) (Model, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE models
		SET public_name=$2, type=$3, model_group=$4, billing_mode=$5,
		    input_price_per_1k=$6::numeric, output_price_per_1k=$7::numeric, price_per_call=$8::numeric,
		    rate_multiplier=$9::numeric, status=$10, sort_order=$11, updated_at=now()
		WHERE id=$1
		RETURNING id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       rate_multiplier::text, status, sort_order, created_at, updated_at
	`, id, input.PublicName, input.Type, input.Group, input.BillingMode, input.InputPricePer1K, input.OutputPricePer1K, input.PricePerCall, input.RateMultiplier, input.Status, input.SortOrder)
	return scanModel(row)
}

func (r ModelRepository) Disable(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE models SET status='disabled', updated_at=now() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
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
		&item.RateMultiplier,
		&item.Status,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

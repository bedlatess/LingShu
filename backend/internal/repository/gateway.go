package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSettlementInsufficientBalance = errors.New("settlement insufficient balance")

type GatewayModel struct {
	ID                      string `json:"-"`
	PublicName              string `json:"id"`
	Type                    string `json:"-"`
	BillingMode             string `json:"-"`
	InputPricePer1K         string `json:"-"`
	OutputPricePer1K        string `json:"-"`
	CacheCreationPricePer1K string `json:"-"`
	CacheReadPricePer1K     string `json:"-"`
	PricePerCall            string `json:"-"`
	RateMultiplier          string `json:"-"`
}

type GatewayChannel struct {
	ID                string
	ProviderType      string
	BaseURL           string
	APIKeyEncrypted   string
	UpstreamModelName string
	Weight            int
	TimeoutSeconds    int
	RPMLimit          int
	ConcurrencyLimit  int
	FailThreshold     int
}

type GatewayRequestRecord struct {
	RequestID           string
	UserID              string
	APIKeyID            string
	ModelID             string
	ChannelID           string
	Endpoint            string
	Status              string
	HTTPStatus          int
	PromptTokens        int
	CompletionTokens    int
	TotalTokens         int
	CacheCreationTokens int
	CacheReadTokens     int
	ImageOutputTokens   int
	BaseCost            string
	RateMultiplier      string
	Charge              string
	IsStream            bool
	IsEstimated         bool
	LatencyMS           int
	FirstTokenMS        int
	UpstreamModelName   string
	ErrorCode           string
	ErrorMessage        string
	ClientIP            string
}

type GatewayRepository struct {
	db *pgxpool.Pool
}

func NewGatewayRepository(db *pgxpool.Pool) GatewayRepository {
	return GatewayRepository{db: db}
}

func (r GatewayRepository) ListEnabledModels(ctx context.Context) ([]GatewayModel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, public_name, type, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       price_per_call::text, rate_multiplier::text
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
	items := []GatewayModel{}
	for rows.Next() {
		var item GatewayModel
		if err := rows.Scan(&item.ID, &item.PublicName, &item.Type, &item.BillingMode, &item.InputPricePer1K, &item.OutputPricePer1K, &item.CacheCreationPricePer1K, &item.CacheReadPricePer1K, &item.PricePerCall, &item.RateMultiplier); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r GatewayRepository) FindEnabledModel(ctx context.Context, publicName string) (GatewayModel, error) {
	var item GatewayModel
	err := r.db.QueryRow(ctx, `
		SELECT id::text, public_name, type, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text,
		       cache_creation_price_per_1k::text, cache_read_price_per_1k::text,
		       price_per_call::text, rate_multiplier::text
		FROM models
		WHERE public_name=$1 AND status='enabled' AND deleted_at IS NULL
	`, publicName).Scan(&item.ID, &item.PublicName, &item.Type, &item.BillingMode, &item.InputPricePer1K, &item.OutputPricePer1K, &item.CacheCreationPricePer1K, &item.CacheReadPricePer1K, &item.PricePerCall, &item.RateMultiplier)
	return item, err
}

func (r GatewayRepository) ListCandidateChannels(ctx context.Context, modelID string) ([]GatewayChannel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id::text, c.provider_type, c.base_url, c.api_key_encrypted, COALESCE(cm.upstream_model_name, ''),
		       c.weight, c.timeout_seconds, c.rpm_limit, c.concurrency_limit, c.fail_threshold
		FROM upstream_channels c
		JOIN channel_models cm ON cm.channel_id = c.id AND cm.model_id=$1 AND cm.status='enabled'
		WHERE c.status='enabled'
		  AND c.deleted_at IS NULL
		  AND (c.health='healthy' OR (c.health='unhealthy' AND c.last_error_at < now() - interval '5 minutes'))
		ORDER BY CASE WHEN c.health='healthy' THEN 0 ELSE 1 END, c.created_at ASC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GatewayChannel{}
	for rows.Next() {
		var item GatewayChannel
		if err := rows.Scan(&item.ID, &item.ProviderType, &item.BaseURL, &item.APIKeyEncrypted, &item.UpstreamModelName, &item.Weight, &item.TimeoutSeconds, &item.RPMLimit, &item.ConcurrencyLimit, &item.FailThreshold); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r GatewayRepository) MarkChannelSuccess(ctx context.Context, channelID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE upstream_channels
		SET health='healthy', fail_count=0, last_success_at=now(), last_error_message=NULL, updated_at=now()
		WHERE id=$1
	`, channelID)
	return err
}

func (r GatewayRepository) MarkChannelFailure(ctx context.Context, channelID, message string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE upstream_channels
		SET fail_count=fail_count+1,
		    health=CASE WHEN fail_count + 1 >= fail_threshold THEN 'unhealthy' ELSE health END,
		    last_error_at=now(),
		    last_error_message=$2,
		    updated_at=now()
		WHERE id=$1
	`, channelID, message)
	return err
}

func (r GatewayRepository) RecordAndCharge(ctx context.Context, record GatewayRequestRecord) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var before string
	if err := tx.QueryRow(ctx, "SELECT balance::text FROM users WHERE id=$1 FOR UPDATE", record.UserID).Scan(&before); err != nil {
		return err
	}

	var requestUUID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO gateway_requests (
			request_id, user_id, api_key_id, model_id, channel_id, endpoint, status, http_status,
			prompt_tokens, completion_tokens, total_tokens, cache_creation_tokens, cache_read_tokens, image_output_tokens,
			base_cost, rate_multiplier, charge, is_stream, is_estimated, latency_ms, first_token_ms,
			upstream_model_name, error_code, error_message, client_ip
		)
		VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::numeric,$16::numeric,$17::numeric,$18,$19,$20,$21,NULLIF($22,''),NULLIF($23,''),NULLIF($24,''),NULLIF($25,'')::inet)
		RETURNING id::text
	`, record.RequestID, record.UserID, record.APIKeyID, record.ModelID, record.ChannelID, record.Endpoint, record.Status, record.HTTPStatus,
		record.PromptTokens, record.CompletionTokens, record.TotalTokens, record.CacheCreationTokens, record.CacheReadTokens, record.ImageOutputTokens,
		record.BaseCost, record.RateMultiplier, record.Charge, record.IsStream, record.IsEstimated, record.LatencyMS, record.FirstTokenMS,
		record.UpstreamModelName, record.ErrorCode, record.ErrorMessage, record.ClientIP).Scan(&requestUUID); err != nil {
		return err
	}

	if record.Status == "success" || record.Status == "partial" {
		var after string
		if err := tx.QueryRow(ctx, `
			UPDATE users
			SET balance = balance - $2::numeric, updated_at=now()
			WHERE id=$1 AND balance >= $2::numeric
			RETURNING balance::text
		`, record.UserID, record.Charge).Scan(&after); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_ = tx.Rollback(ctx)
				record.Status = "failed"
				record.ErrorCode = "settlement_insufficient_balance"
				record.ErrorMessage = "insufficient balance during final settlement"
				if recordErr := r.RecordOnly(ctx, record); recordErr != nil {
					return recordErr
				}
				return ErrSettlementInsufficientBalance
			}
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO balance_ledger (
				user_id, type, amount, balance_before, balance_after, base_cost, rate_multiplier,
				related_type, related_id, remark
			)
			VALUES ($1, 'usage_charge', -($2::numeric), $3::numeric, $4::numeric, $5::numeric, $6::numeric, 'gateway_request', $7, 'gateway usage charge')
		`, record.UserID, record.Charge, before, after, record.BaseCost, record.RateMultiplier, requestUUID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r GatewayRepository) RecordOnly(ctx context.Context, record GatewayRequestRecord) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO gateway_requests (
			request_id, user_id, api_key_id, model_id, channel_id, endpoint, status, http_status,
			prompt_tokens, completion_tokens, total_tokens, cache_creation_tokens, cache_read_tokens, image_output_tokens,
			base_cost, rate_multiplier, charge, is_stream, is_estimated, latency_ms, first_token_ms,
			upstream_model_name, error_code, error_message, client_ip
		)
		VALUES ($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15::numeric,$16::numeric,$17::numeric,$18,$19,$20,$21,NULLIF($22,''),NULLIF($23,''),NULLIF($24,''),NULLIF($25,'')::inet)
		ON CONFLICT (request_id) DO NOTHING
	`, record.RequestID, record.UserID, record.APIKeyID, record.ModelID, record.ChannelID, record.Endpoint, record.Status, record.HTTPStatus,
		record.PromptTokens, record.CompletionTokens, record.TotalTokens, record.CacheCreationTokens, record.CacheReadTokens, record.ImageOutputTokens,
		record.BaseCost, record.RateMultiplier, record.Charge, record.IsStream, record.IsEstimated, record.LatencyMS, record.FirstTokenMS,
		record.UpstreamModelName, record.ErrorCode, record.ErrorMessage, record.ClientIP)
	return err
}

func NowMS(start time.Time) int {
	return int(time.Since(start).Milliseconds())
}

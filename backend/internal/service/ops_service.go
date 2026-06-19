package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OpsService struct {
	db *pgxpool.Pool
}

type OpsSummary struct {
	RPM             int    `json:"rpm"`
	TPM             int    `json:"tpm"`
	Requests24h     int    `json:"requests_24h"`
	ErrorRate24h    string `json:"error_rate_24h"`
	P50LatencyMS    int    `json:"p50_latency_ms"`
	P95LatencyMS    int    `json:"p95_latency_ms"`
	AvgFirstTokenMS int    `json:"avg_first_token_ms"`
	ChannelSwitches int    `json:"channel_switches"`
}

type OpsTrendPoint struct {
	Bucket       string `json:"bucket"`
	Requests     int    `json:"requests"`
	Failures     int    `json:"failures"`
	TotalTokens  int    `json:"total_tokens"`
	Charge       string `json:"charge"`
	AvgLatencyMS int    `json:"avg_latency_ms"`
	P95LatencyMS int    `json:"p95_latency_ms"`
}

type OpsChannelHealth struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	ProviderType     string     `json:"provider_type"`
	Status           string     `json:"status"`
	Health           string     `json:"health"`
	FailCount        int        `json:"fail_count"`
	LastLatencyMS    int        `json:"last_latency_ms"`
	LastSuccessAt    *time.Time `json:"last_success_at,omitempty"`
	LastErrorAt      *time.Time `json:"last_error_at,omitempty"`
	LastErrorMessage string     `json:"last_error_message"`
	Requests24h      int        `json:"requests_24h"`
	Failures24h      int        `json:"failures_24h"`
	ErrorRate24h     string     `json:"error_rate_24h"`
	AvgLatencyMS     int        `json:"avg_latency_ms"`
}

type OpsStatusBucket struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type OpsDashboard struct {
	Summary  OpsSummary         `json:"summary"`
	Trends   []OpsTrendPoint    `json:"trends"`
	Channels []OpsChannelHealth `json:"channels"`
	Statuses []OpsStatusBucket  `json:"statuses"`
	Alerts   []OpsAlert         `json:"alerts"`
}

func NewOpsService(db *pgxpool.Pool) OpsService {
	return OpsService{db: db}
}

func (s OpsService) Dashboard(ctx context.Context, alerts ...OpsAlertService) (OpsDashboard, error) {
	summary, err := s.summary(ctx)
	if err != nil {
		return OpsDashboard{}, err
	}
	trends, err := s.trends(ctx)
	if err != nil {
		return OpsDashboard{}, err
	}
	channels, err := s.channels(ctx)
	if err != nil {
		return OpsDashboard{}, err
	}
	statuses, err := s.statuses(ctx)
	if err != nil {
		return OpsDashboard{}, err
	}
	activeAlerts := []OpsAlert{}
	if len(alerts) > 0 {
		var err error
		activeAlerts, err = alerts[0].Active(ctx)
		if err != nil {
			return OpsDashboard{}, err
		}
	}
	return OpsDashboard{Summary: summary, Trends: trends, Channels: channels, Statuses: statuses, Alerts: activeAlerts}, nil
}

func (s OpsService) summary(ctx context.Context) (OpsSummary, error) {
	var item OpsSummary
	err := s.db.QueryRow(ctx, `
		WITH recent AS (
			SELECT *
			FROM gateway_requests
			WHERE created_at >= now() - interval '24 hours'
		),
		switches AS (
			SELECT count(*)::int AS count
			FROM (
				SELECT channel_id, lag(channel_id) OVER (ORDER BY created_at) AS prev_channel_id
				FROM recent
				WHERE channel_id IS NOT NULL
				ORDER BY created_at
			) t
			WHERE prev_channel_id IS NOT NULL AND channel_id IS DISTINCT FROM prev_channel_id
		)
		SELECT
			COALESCE((SELECT count(*) FROM gateway_requests WHERE created_at >= now() - interval '1 minute'),0)::int,
			COALESCE((SELECT sum(total_tokens) FROM gateway_requests WHERE created_at >= now() - interval '1 minute'),0)::int,
			COALESCE(count(*),0)::int,
			COALESCE(((count(*) FILTER (WHERE status!='success'))::numeric / NULLIF(count(*),0)) * 100,0)::numeric(8,2)::text,
			COALESCE(percentile_cont(0.50) WITHIN GROUP (ORDER BY latency_ms),0)::int,
			COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms),0)::int,
			COALESCE(avg(NULLIF(first_token_ms,0)),0)::int,
			COALESCE((SELECT count FROM switches),0)::int
		FROM recent
	`).Scan(&item.RPM, &item.TPM, &item.Requests24h, &item.ErrorRate24h, &item.P50LatencyMS, &item.P95LatencyMS, &item.AvgFirstTokenMS, &item.ChannelSwitches)
	return item, err
}

func (s OpsService) trends(ctx context.Context) ([]OpsTrendPoint, error) {
	rows, err := s.db.Query(ctx, `
		SELECT to_char(date_trunc('hour', created_at), 'MM-DD HH24:00') AS bucket,
		       count(*)::int,
		       count(*) FILTER (WHERE status!='success')::int,
		       COALESCE(sum(total_tokens),0)::int,
		       COALESCE(sum(charge),0)::text,
		       COALESCE(avg(latency_ms),0)::int,
		       COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms),0)::int
		FROM gateway_requests
		WHERE created_at >= now() - interval '24 hours'
		GROUP BY date_trunc('hour', created_at)
		ORDER BY date_trunc('hour', created_at)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []OpsTrendPoint{}
	for rows.Next() {
		var item OpsTrendPoint
		if err := rows.Scan(&item.Bucket, &item.Requests, &item.Failures, &item.TotalTokens, &item.Charge, &item.AvgLatencyMS, &item.P95LatencyMS); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s OpsService) channels(ctx context.Context) ([]OpsChannelHealth, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.id::text, c.name, c.provider_type, c.status, c.health, c.fail_count,
		       c.last_latency_ms, c.last_success_at, c.last_error_at, COALESCE(c.last_error_message,''),
		       count(gr.id)::int,
		       count(gr.id) FILTER (WHERE gr.status!='success')::int,
		       COALESCE(((count(gr.id) FILTER (WHERE gr.status!='success'))::numeric / NULLIF(count(gr.id),0)) * 100,0)::numeric(8,2)::text,
		       COALESCE(avg(gr.latency_ms),0)::int
		FROM upstream_channels c
		LEFT JOIN gateway_requests gr ON gr.channel_id = c.id AND gr.created_at >= now() - interval '24 hours'
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.provider_type, c.status, c.health, c.fail_count,
		         c.last_latency_ms, c.last_success_at, c.last_error_at, c.last_error_message
		ORDER BY CASE WHEN c.health='healthy' AND c.status='enabled' THEN 0 ELSE 1 END,
		         count(gr.id) DESC, c.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []OpsChannelHealth{}
	for rows.Next() {
		var item OpsChannelHealth
		if err := rows.Scan(&item.ID, &item.Name, &item.ProviderType, &item.Status, &item.Health, &item.FailCount, &item.LastLatencyMS, &item.LastSuccessAt, &item.LastErrorAt, &item.LastErrorMessage, &item.Requests24h, &item.Failures24h, &item.ErrorRate24h, &item.AvgLatencyMS); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s OpsService) statuses(ctx context.Context) ([]OpsStatusBucket, error) {
	rows, err := s.db.Query(ctx, `
		SELECT http_status::text, count(*)::int
		FROM gateway_requests
		WHERE created_at >= now() - interval '24 hours'
		GROUP BY http_status
		ORDER BY count(*) DESC, http_status ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []OpsStatusBucket{}
	for rows.Next() {
		var item OpsStatusBucket
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

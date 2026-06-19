package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"lingshu/backend/internal/repository"
)

type OpsAlertService struct {
	db           *pgxpool.Pool
	settings     repository.SettingsRepository
	audits       repository.AuditRepository
	notification NotificationService
}

type OpsAlert struct {
	ID             string     `json:"id"`
	RuleKey        string     `json:"rule_key"`
	Severity       string     `json:"severity"`
	TargetType     string     `json:"target_type"`
	TargetID       string     `json:"target_id"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Status         string     `json:"status"`
	LastNotifiedAt *time.Time `json:"last_notified_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type candidateAlert struct {
	RuleKey     string
	Severity    string
	TargetType  string
	TargetID    string
	Title       string
	Message     string
	Fingerprint string
}

func NewOpsAlertService(db *pgxpool.Pool, settings repository.SettingsRepository, audits repository.AuditRepository, notification NotificationService) OpsAlertService {
	return OpsAlertService{db: db, settings: settings, audits: audits, notification: notification}
}

func (s OpsAlertService) Active(ctx context.Context) ([]OpsAlert, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, rule_key, severity, target_type, COALESCE(target_id::text,''), title, message, status,
		       last_notified_at, created_at, updated_at
		FROM ops_alerts
		WHERE status='active'
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []OpsAlert{}
	for rows.Next() {
		var item OpsAlert
		if err := rows.Scan(&item.ID, &item.RuleKey, &item.Severity, &item.TargetType, &item.TargetID, &item.Title, &item.Message, &item.Status, &item.LastNotifiedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s OpsAlertService) Evaluate(ctx context.Context) error {
	cfg, err := s.config(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	candidates, err := s.candidates(ctx, cfg)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		created, err := s.upsertAlert(ctx, candidate)
		if err != nil {
			return err
		}
		if candidate.RuleKey == "channel_consecutive_failures" {
			if err := s.disableChannel(ctx, candidate.TargetID, candidate); err != nil {
				return err
			}
		}
		if created {
			_ = s.notification.SendAlert(ctx, AlertNotification{
				RuleKey:    candidate.RuleKey,
				Severity:   candidate.Severity,
				Title:      candidate.Title,
				Message:    candidate.Message,
				TargetType: candidate.TargetType,
				TargetID:   candidate.TargetID,
			})
			_ = s.markNotified(ctx, candidate.Fingerprint)
		}
	}
	return nil
}

type alertConfig struct {
	Enabled                    bool
	ChannelFailureThreshold    int
	Gateway5xxRateThreshold    float64
	UpstreamErrorRateThreshold float64
	LowBalanceThreshold        float64
}

func (s OpsAlertService) config(ctx context.Context) (alertConfig, error) {
	settings, err := s.settings.GetMap(ctx,
		"alert_enabled",
		"alert_channel_failure_threshold",
		"alert_gateway_5xx_rate_threshold",
		"alert_upstream_error_rate_threshold",
		"alert_low_balance_threshold",
	)
	if err != nil {
		return alertConfig{}, err
	}
	return alertConfig{
		Enabled:                    strings.EqualFold(settings["alert_enabled"], "true"),
		ChannelFailureThreshold:    intSetting(settings["alert_channel_failure_threshold"], 5),
		Gateway5xxRateThreshold:    floatSetting(settings["alert_gateway_5xx_rate_threshold"], 0.20),
		UpstreamErrorRateThreshold: floatSetting(settings["alert_upstream_error_rate_threshold"], 0.20),
		LowBalanceThreshold:        floatSetting(settings["alert_low_balance_threshold"], 5),
	}, nil
}

func (s OpsAlertService) candidates(ctx context.Context, cfg alertConfig) ([]candidateAlert, error) {
	out := []candidateAlert{}
	channelAlerts, err := s.channelFailureAlerts(ctx, cfg.ChannelFailureThreshold)
	if err != nil {
		return nil, err
	}
	out = append(out, channelAlerts...)
	if item, ok, err := s.gateway5xxAlert(ctx, cfg.Gateway5xxRateThreshold); err != nil {
		return nil, err
	} else if ok {
		out = append(out, item)
	}
	if item, ok, err := s.upstreamErrorAlert(ctx, cfg.UpstreamErrorRateThreshold); err != nil {
		return nil, err
	} else if ok {
		out = append(out, item)
	}
	lowBalanceAlerts, err := s.lowBalanceAlerts(ctx, cfg.LowBalanceThreshold)
	if err != nil {
		return nil, err
	}
	out = append(out, lowBalanceAlerts...)
	return out, nil
}

func (s OpsAlertService) channelFailureAlerts(ctx context.Context, threshold int) ([]candidateAlert, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, name, fail_count
		FROM upstream_channels
		WHERE deleted_at IS NULL AND status='enabled' AND fail_count >= $1
	`, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []candidateAlert{}
	for rows.Next() {
		var id, name string
		var failCount int
		if err := rows.Scan(&id, &name, &failCount); err != nil {
			return nil, err
		}
		items = append(items, candidateAlert{
			RuleKey:     "channel_consecutive_failures",
			Severity:    "critical",
			TargetType:  "channel",
			TargetID:    id,
			Title:       "渠道连续失败",
			Message:     fmt.Sprintf("渠道 %s 连续失败 %d 次，已达到自动禁用阈值。", name, failCount),
			Fingerprint: "channel_consecutive_failures:" + id,
		})
	}
	return items, rows.Err()
}

func (s OpsAlertService) gateway5xxAlert(ctx context.Context, threshold float64) (candidateAlert, bool, error) {
	var total, failures int
	if err := s.db.QueryRow(ctx, `
		SELECT count(*)::int, count(*) FILTER (WHERE http_status >= 500)::int
		FROM gateway_requests
		WHERE created_at >= now() - interval '15 minutes'
	`).Scan(&total, &failures); err != nil {
		return candidateAlert{}, false, err
	}
	if total < 20 || float64(failures)/float64(total) < threshold {
		return candidateAlert{}, false, nil
	}
	return candidateAlert{
		RuleKey:     "gateway_5xx_rate",
		Severity:    "warning",
		TargetType:  "gateway",
		Title:       "网关 5xx 比例过高",
		Message:     fmt.Sprintf("最近 15 分钟 %d/%d 个请求返回 5xx。", failures, total),
		Fingerprint: "gateway_5xx_rate",
	}, true, nil
}

func (s OpsAlertService) upstreamErrorAlert(ctx context.Context, threshold float64) (candidateAlert, bool, error) {
	var total, failures int
	if err := s.db.QueryRow(ctx, `
		SELECT count(*)::int, count(*) FILTER (WHERE http_status IN (401, 429))::int
		FROM gateway_requests
		WHERE created_at >= now() - interval '15 minutes'
	`).Scan(&total, &failures); err != nil {
		return candidateAlert{}, false, err
	}
	if total < 20 || float64(failures)/float64(total) < threshold {
		return candidateAlert{}, false, nil
	}
	return candidateAlert{
		RuleKey:     "upstream_auth_or_limit_rate",
		Severity:    "warning",
		TargetType:  "upstream",
		Title:       "上游认证或限流错误过高",
		Message:     fmt.Sprintf("最近 15 分钟 %d/%d 个请求返回 401/429。", failures, total),
		Fingerprint: "upstream_auth_or_limit_rate",
	}, true, nil
}

func (s OpsAlertService) lowBalanceAlerts(ctx context.Context, threshold float64) ([]candidateAlert, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id::text, username, balance::text
		FROM users
		WHERE status='active' AND balance > 0 AND balance < $1::numeric
		ORDER BY balance ASC
		LIMIT 50
	`, fmt.Sprintf("%.6f", threshold))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []candidateAlert{}
	for rows.Next() {
		var id, username, balance string
		if err := rows.Scan(&id, &username, &balance); err != nil {
			return nil, err
		}
		items = append(items, candidateAlert{
			RuleKey:     "user_low_balance",
			Severity:    "info",
			TargetType:  "user",
			TargetID:    id,
			Title:       "用户余额偏低",
			Message:     fmt.Sprintf("用户 %s 当前余额 %s，低于告警阈值。", username, balance),
			Fingerprint: "user_low_balance:" + id,
		})
	}
	return items, rows.Err()
}

func (s OpsAlertService) upsertAlert(ctx context.Context, item candidateAlert) (bool, error) {
	var created bool
	err := s.db.QueryRow(ctx, `
		INSERT INTO ops_alerts (rule_key, severity, target_type, target_id, title, message, fingerprint)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7)
		ON CONFLICT (fingerprint) WHERE status='active'
		DO UPDATE SET message=EXCLUDED.message, updated_at=now()
		RETURNING xmax = 0
	`, item.RuleKey, item.Severity, item.TargetType, item.TargetID, item.Title, item.Message, item.Fingerprint).Scan(&created)
	return created, err
}

func (s OpsAlertService) markNotified(ctx context.Context, fingerprint string) error {
	_, err := s.db.Exec(ctx, "UPDATE ops_alerts SET last_notified_at=now(), updated_at=now() WHERE fingerprint=$1 AND status='active'", fingerprint)
	return err
}

func (s OpsAlertService) disableChannel(ctx context.Context, id string, item candidateAlert) error {
	tag, err := s.db.Exec(ctx, "UPDATE upstream_channels SET status='disabled', updated_at=now() WHERE id=$1 AND status='enabled' AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return s.audits.Write(ctx, repository.AuditEntry{
		Action:     "ops.alert.channel_auto_disabled",
		TargetType: "channel",
		TargetID:   id,
		After:      item,
		UserAgent:  "ops-alert-service",
	})
}

func intSetting(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func floatSetting(value string, fallback float64) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

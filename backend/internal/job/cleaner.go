package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Cleaner struct {
	db    *pgxpool.Pool
	redis *redis.Client
	cfg   CleanerConfig
}

type CleanerConfig struct {
	LogRetentionDays      int
	AuditRetentionDays    int
	AnnouncementGraceDays int
	RedeemGraceDays       int
}

type CleanupResult struct {
	Table     string    `json:"table"`
	Deleted   int64     `json:"deleted"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Err       string    `json:"err,omitempty"`
}

type CleanupHistory struct {
	ID        string          `json:"id"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   time.Time       `json:"ended_at"`
	Results   []CleanupResult `json:"results"`
}

func NewCleaner(db *pgxpool.Pool, redisClient *redis.Client, cfg CleanerConfig) Cleaner {
	if cfg.LogRetentionDays <= 0 {
		cfg.LogRetentionDays = 30
	}
	if cfg.AuditRetentionDays <= 0 {
		cfg.AuditRetentionDays = 90
	}
	if cfg.AnnouncementGraceDays <= 0 {
		cfg.AnnouncementGraceDays = 30
	}
	if cfg.RedeemGraceDays <= 0 {
		cfg.RedeemGraceDays = 90
	}
	return Cleaner{db: db, redis: redisClient, cfg: cfg}
}

func (c Cleaner) Run(ctx context.Context) []CleanupResult {
	results := []CleanupResult{
		c.runDelete(ctx, "gateway_requests", "DELETE FROM gateway_requests WHERE ctid IN (SELECT ctid FROM gateway_requests WHERE created_at < now() - ($1::int * interval '1 day') LIMIT 10000)", c.cfg.LogRetentionDays),
		c.runDelete(ctx, "audit_logs", "DELETE FROM audit_logs WHERE ctid IN (SELECT ctid FROM audit_logs WHERE created_at < now() - ($1::int * interval '1 day') LIMIT 10000)", c.cfg.AuditRetentionDays),
		c.runDelete(ctx, "announcements", "DELETE FROM announcements WHERE ctid IN (SELECT ctid FROM announcements WHERE status='offline' AND expire_at IS NOT NULL AND expire_at < now() - ($1::int * interval '1 day') LIMIT 10000)", c.cfg.AnnouncementGraceDays),
		c.runDelete(ctx, "redeem_codes", "DELETE FROM redeem_codes WHERE ctid IN (SELECT ctid FROM redeem_codes WHERE used_count=0 AND expires_at IS NOT NULL AND expires_at < now() - ($1::int * interval '1 day') LIMIT 10000)", c.cfg.RedeemGraceDays),
		c.cleanupRedisFrozen(ctx),
	}
	return results
}

func (c Cleaner) SaveHistory(ctx context.Context, results []CleanupResult) error {
	startedAt, endedAt := cleanupWindow(results)
	payload, err := json.Marshal(results)
	if err != nil {
		return err
	}
	_, err = c.db.Exec(ctx, "INSERT INTO cleanup_history (started_at, ended_at, result_json) VALUES ($1, $2, $3::jsonb)", startedAt, endedAt, string(payload))
	return err
}

func (c Cleaner) History(ctx context.Context, limit int) ([]CleanupHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := c.db.Query(ctx, `
		SELECT id::text, started_at, ended_at, result_json
		FROM cleanup_history
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []CleanupHistory{}
	for rows.Next() {
		var item CleanupHistory
		var raw []byte
		if err := rows.Scan(&item.ID, &item.StartedAt, &item.EndedAt, &raw); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &item.Results)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (c Cleaner) runDelete(ctx context.Context, table, query string, days int) CleanupResult {
	result := CleanupResult{Table: table, StartedAt: time.Now()}
	for {
		tag, err := c.db.Exec(ctx, query, days)
		if err != nil {
			result.Err = err.Error()
			break
		}
		deleted := tag.RowsAffected()
		result.Deleted += deleted
		if deleted < 10000 {
			break
		}
	}
	result.EndedAt = time.Now()
	return result
}

func (c Cleaner) cleanupRedisFrozen(ctx context.Context) CleanupResult {
	result := CleanupResult{Table: "redis:frozen", StartedAt: time.Now()}
	if c.redis == nil {
		result.EndedAt = time.Now()
		return result
	}
	var cursor uint64
	for {
		keys, next, err := c.redis.Scan(ctx, cursor, "frozen:*", 100).Result()
		if err != nil {
			result.Err = err.Error()
			break
		}
		cursor = next
		for _, key := range keys {
			ttl, err := c.redis.TTL(ctx, key).Result()
			if err != nil {
				continue
			}
			if ttl < 0 {
				if deleted, err := c.redis.Del(ctx, key).Result(); err == nil {
					result.Deleted += deleted
				}
			}
		}
		if cursor == 0 {
			break
		}
	}
	result.EndedAt = time.Now()
	return result
}

func cleanupWindow(results []CleanupResult) (time.Time, time.Time) {
	if len(results) == 0 {
		now := time.Now()
		return now, now
	}
	startedAt := results[0].StartedAt
	endedAt := results[0].EndedAt
	for _, item := range results[1:] {
		if item.StartedAt.Before(startedAt) {
			startedAt = item.StartedAt
		}
		if item.EndedAt.After(endedAt) {
			endedAt = item.EndedAt
		}
	}
	return startedAt, endedAt
}

func (r CleanupResult) String() string {
	if r.Err != "" {
		return fmt.Sprintf("%s deleted=%d err=%s", r.Table, r.Deleted, r.Err)
	}
	return fmt.Sprintf("%s deleted=%d", r.Table, r.Deleted)
}

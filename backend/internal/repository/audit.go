package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

type AuditEntry struct {
	ActorID    string
	Action     string
	TargetType string
	TargetID   string
	Before     any
	After      any
	IP         string
	UserAgent  string
}

type AuditLog struct {
	ID         string          `json:"id"`
	ActorID    string          `json:"actor_id"`
	Action     string          `json:"action"`
	TargetType string          `json:"target_type"`
	TargetID   string          `json:"target_id"`
	Before     json.RawMessage `json:"before_snapshot,omitempty"`
	After      json.RawMessage `json:"after_snapshot,omitempty"`
	IP         string          `json:"ip"`
	UserAgent  string          `json:"user_agent"`
	CreatedAt  time.Time       `json:"created_at"`
}

func NewAuditRepository(db *pgxpool.Pool) AuditRepository {
	return AuditRepository{db: db}
}

func (r AuditRepository) Write(ctx context.Context, entry AuditEntry) error {
	before, _ := json.Marshal(entry.Before)
	after, _ := json.Marshal(entry.After)
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_logs (actor_id, action, target_type, target_id, before_snapshot, after_snapshot, ip, user_agent)
		VALUES (NULLIF($1, '')::uuid, $2, $3, NULLIF($4, '')::uuid, NULLIF($5, 'null')::jsonb, NULLIF($6, 'null')::jsonb, NULLIF($7, '')::inet, NULLIF($8, ''))
	`, entry.ActorID, entry.Action, entry.TargetType, entry.TargetID, string(before), string(after), entry.IP, entry.UserAgent)
	return err
}

func (r AuditRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT count(*) FROM audit_logs").Scan(&count)
	return count, err
}

func (r AuditRepository) List(ctx context.Context, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, COALESCE(actor_id::text, ''), action, target_type, COALESCE(target_id::text, ''),
		       COALESCE(before_snapshot, '{}'::jsonb), COALESCE(after_snapshot, '{}'::jsonb),
		       COALESCE(ip::text, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AuditLog{}
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(&item.ID, &item.ActorID, &item.Action, &item.TargetType, &item.TargetID, &item.Before, &item.After, &item.IP, &item.UserAgent, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

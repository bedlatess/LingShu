package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SettingUpdate struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SettingsRepository struct {
	db *pgxpool.Pool
}

func NewSettingsRepository(db *pgxpool.Pool) SettingsRepository {
	return SettingsRepository{db: db}
}

func (r SettingsRepository) List(ctx context.Context) ([]Setting, error) {
	items, _, err := r.ListPaged(ctx, 100, 0)
	return items, err
}

func (r SettingsRepository) GetMap(ctx context.Context, keys ...string) (map[string]string, error) {
	rows, err := r.db.Query(ctx, `SELECT key, value FROM system_settings WHERE key = ANY($1)`, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, rows.Err()
}

func (r SettingsRepository) ListPaged(ctx context.Context, limit, offset int) ([]Setting, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM system_settings`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT key, value, description, COALESCE(updated_by::text, ''), updated_at
		FROM system_settings
		ORDER BY key ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []Setting{}
	for rows.Next() {
		var item Setting
		if err := rows.Scan(&item.Key, &item.Value, &item.Description, &item.UpdatedBy, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r SettingsRepository) Patch(ctx context.Context, actorID string, updates []SettingUpdate) ([]Setting, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	for _, update := range updates {
		if _, err := tx.Exec(ctx, `
			UPDATE system_settings
			SET value=$2, updated_by=$3::uuid, updated_at=now()
			WHERE key=$1
		`, update.Key, update.Value, actorID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.List(ctx)
}

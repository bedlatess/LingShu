package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type APIKey struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Mask      string     `json:"mask"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyPrincipal struct {
	APIKeyID         string
	UserID           string
	UserRole         string
	UserStatus       string
	KeyStatus        string
	Balance          string
	RPMLimit         int
	ConcurrencyLimit int
}

type CreateAPIKeyParams struct {
	UserID    string
	KeyPrefix string
	KeyHash   string
	Mask      string
	Name      string
}

type UpdateAPIKeyParams struct {
	ID     string
	UserID string
	Name   string
	Status string
}

type APIKeyRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewAPIKeyRepository(db *pgxpool.Pool) APIKeyRepository {
	return APIKeyRepository{db: db}
}

func NewAPIKeyRepositoryWithCache(db *pgxpool.Pool, redisClient *redis.Client) APIKeyRepository {
	return APIKeyRepository{db: db, redis: redisClient}
}

func (r APIKeyRepository) HasStore() bool {
	return r.db != nil
}

func (r APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]APIKey, error) {
	items, _, err := r.ListByUserPaged(ctx, userID, 100, 0)
	return items, err
}

func (r APIKeyRepository) ListByUserPaged(ctx context.Context, userID string, limit, offset int) ([]APIKey, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM api_keys WHERE user_id=$1 AND deleted_at IS NULL`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, key_prefix, name, status, expires_at, created_at
		FROM api_keys
		WHERE user_id=$1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []APIKey{}
	for rows.Next() {
		item, err := scanAPIKey(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r APIKeyRepository) ListAll(ctx context.Context) ([]APIKey, error) {
	items, _, err := r.ListAllPaged(ctx, 100, 0)
	return items, err
}

func (r APIKeyRepository) ListAllPaged(ctx context.Context, limit, offset int) ([]APIKey, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM api_keys WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, key_prefix, name, status, expires_at, created_at
		FROM api_keys
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []APIKey{}
	for rows.Next() {
		item, err := scanAPIKey(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r APIKeyRepository) Create(ctx context.Context, params CreateAPIKeyParams) (APIKey, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO api_keys (user_id, key_prefix, key_hash, name, status)
		VALUES ($1, $2, $3, $4, 'active')
		RETURNING id::text, user_id::text, key_prefix, name, status, expires_at, created_at
	`, params.UserID, params.Mask, params.KeyHash, params.Name)
	return scanAPIKey(row)
}

func (r APIKeyRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_ = r.InvalidateByID(ctx, id)
	tag, err := r.db.Exec(ctx, "UPDATE api_keys SET status=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r APIKeyRepository) UpdateForUser(ctx context.Context, params UpdateAPIKeyParams) (APIKey, error) {
	_ = r.InvalidateByID(ctx, params.ID)
	row := r.db.QueryRow(ctx, `
		UPDATE api_keys
		SET name=CASE WHEN $3='' THEN name ELSE $3 END,
		    status=CASE WHEN $4='' THEN status ELSE $4 END,
		    updated_at=now()
		WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL
		RETURNING id::text, user_id::text, key_prefix, name, status, expires_at, created_at
	`, params.ID, params.UserID, params.Name, params.Status)
	return scanAPIKey(row)
}

func (r APIKeyRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	_ = r.InvalidateByID(ctx, id)
	tag, err := r.db.Exec(ctx, "DELETE FROM api_keys WHERE id=$1 AND user_id=$2", id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r APIKeyRepository) Delete(ctx context.Context, id string) error {
	_ = r.InvalidateByID(ctx, id)
	tag, err := r.db.Exec(ctx, "UPDATE api_keys SET deleted_at=now(), status='disabled', updated_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("api key not found")
	}
	return nil
}

func (r APIKeyRepository) FindPrincipalByHash(ctx context.Context, hash string) (APIKeyPrincipal, error) {
	var principal APIKeyPrincipal
	if r.redis != nil {
		payload, err := r.redis.Get(ctx, apiKeyCacheKey(hash)).Bytes()
		if err == nil {
			if jsonErr := json.Unmarshal(payload, &principal); jsonErr == nil {
				if balanceErr := r.refreshPrincipalBalance(ctx, &principal); balanceErr != nil {
					return APIKeyPrincipal{}, balanceErr
				}
				return principal, nil
			}
		} else if err != redis.Nil {
			// Cache failures should not break gateway authentication.
		}
	}
	err := r.db.QueryRow(ctx, `
		SELECT k.id::text, u.id::text, u.role, u.status, k.status, u.balance::text,
		       k.rpm_limit, k.concurrency_limit
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.key_hash=$1 AND k.deleted_at IS NULL
	`, hash).Scan(&principal.APIKeyID, &principal.UserID, &principal.UserRole, &principal.UserStatus, &principal.KeyStatus, &principal.Balance, &principal.RPMLimit, &principal.ConcurrencyLimit)
	if err == nil && r.redis != nil {
		if payload, jsonErr := json.Marshal(principal); jsonErr == nil {
			_ = r.redis.Set(ctx, apiKeyCacheKey(hash), payload, time.Minute).Err()
		}
	}
	return principal, err
}

func (r APIKeyRepository) DisableByUser(ctx context.Context, userID string) error {
	hashes, err := r.hashesByUser(ctx, userID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, "UPDATE api_keys SET status='disabled', updated_at=now() WHERE user_id=$1 AND deleted_at IS NULL", userID)
	if err != nil {
		return err
	}
	r.invalidateHashes(ctx, hashes)
	return nil
}

func (r APIKeyRepository) refreshPrincipalBalance(ctx context.Context, principal *APIKeyPrincipal) error {
	return r.db.QueryRow(ctx, `
		SELECT u.balance::text
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.id=$1 AND k.deleted_at IS NULL
	`, principal.APIKeyID).Scan(&principal.Balance)
}

func (r APIKeyRepository) InvalidateByID(ctx context.Context, id string) error {
	if r.redis == nil {
		return nil
	}
	var hash string
	if err := r.db.QueryRow(ctx, "SELECT key_hash FROM api_keys WHERE id=$1", id).Scan(&hash); err != nil {
		return err
	}
	return r.redis.Del(ctx, apiKeyCacheKey(hash)).Err()
}

func (r APIKeyRepository) hashesByUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx, "SELECT key_hash FROM api_keys WHERE user_id=$1 AND deleted_at IS NULL", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hashes := []string{}
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, rows.Err()
}

func (r APIKeyRepository) invalidateHashes(ctx context.Context, hashes []string) {
	if r.redis == nil || len(hashes) == 0 {
		return
	}
	keys := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		keys = append(keys, apiKeyCacheKey(hash))
	}
	_ = r.redis.Del(ctx, keys...).Err()
}

func apiKeyCacheKey(hash string) string {
	return "apikey:" + hash
}

type apiKeyScanner interface {
	Scan(dest ...any) error
}

func scanAPIKey(row apiKeyScanner) (APIKey, error) {
	var item APIKey
	err := row.Scan(&item.ID, &item.UserID, &item.Mask, &item.Name, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}

package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db *pgxpool.Pool
}

func NewAPIKeyRepository(db *pgxpool.Pool) APIKeyRepository {
	return APIKeyRepository{db: db}
}

func (r APIKeyRepository) ListByUser(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, key_prefix, name, status, expires_at, created_at
		FROM api_keys
		WHERE user_id=$1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []APIKey{}
	for rows.Next() {
		item, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r APIKeyRepository) ListAll(ctx context.Context) ([]APIKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, key_prefix, name, status, expires_at, created_at
		FROM api_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []APIKey{}
	for rows.Next() {
		item, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
	tag, err := r.db.Exec(ctx, "UPDATE api_keys SET status=$2, updated_at=now() WHERE id=$1", id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r APIKeyRepository) UpdateForUser(ctx context.Context, params UpdateAPIKeyParams) (APIKey, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE api_keys
		SET name=CASE WHEN $3='' THEN name ELSE $3 END,
		    status=CASE WHEN $4='' THEN status ELSE $4 END,
		    updated_at=now()
		WHERE id=$1 AND user_id=$2
		RETURNING id::text, user_id::text, key_prefix, name, status, expires_at, created_at
	`, params.ID, params.UserID, params.Name, params.Status)
	return scanAPIKey(row)
}

func (r APIKeyRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM api_keys WHERE id=$1 AND user_id=$2", id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r APIKeyRepository) FindPrincipalByHash(ctx context.Context, hash string) (APIKeyPrincipal, error) {
	var principal APIKeyPrincipal
	err := r.db.QueryRow(ctx, `
		SELECT k.id::text, u.id::text, u.role, u.status, k.status, u.balance::text,
		       k.rpm_limit, k.concurrency_limit
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.key_hash=$1
	`, hash).Scan(&principal.APIKeyID, &principal.UserID, &principal.UserRole, &principal.UserStatus, &principal.KeyStatus, &principal.Balance, &principal.RPMLimit, &principal.ConcurrencyLimit)
	return principal, err
}

type apiKeyScanner interface {
	Scan(dest ...any) error
}

func scanAPIKey(row apiKeyScanner) (APIKey, error) {
	var item APIKey
	err := row.Scan(&item.ID, &item.UserID, &item.Mask, &item.Name, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}

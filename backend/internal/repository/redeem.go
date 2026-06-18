package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRedeemUnavailable = errors.New("redeem code unavailable")

type RedeemCode struct {
	ID        string     `json:"id"`
	Code      string     `json:"code,omitempty"`
	Prefix    string     `json:"code_prefix"`
	BatchName string     `json:"batch_name"`
	Amount    string     `json:"amount"`
	Status    string     `json:"status"`
	MaxUses   int        `json:"max_uses"`
	UsedCount int        `json:"used_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateRedeemCodeInput struct {
	CodeHash   string
	CodePrefix string
	BatchName  string
	Amount     string
	MaxUses    int
	ExpiresAt  *time.Time
	CreatedBy  string
}

type RedeemRepository struct {
	db *pgxpool.Pool
}

func NewRedeemRepository(db *pgxpool.Pool) RedeemRepository {
	return RedeemRepository{db: db}
}

func (r RedeemRepository) List(ctx context.Context) ([]RedeemCode, error) {
	items, _, err := r.ListPaged(ctx, 100, 0)
	return items, err
}

func (r RedeemRepository) ListPaged(ctx context.Context, limit, offset int) ([]RedeemCode, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM redeem_codes`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, code_prefix, batch_name, amount::text, status, max_uses, used_count, expires_at, created_at
		FROM redeem_codes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []RedeemCode{}
	for rows.Next() {
		item, err := scanRedeemCode(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r RedeemRepository) Create(ctx context.Context, input CreateRedeemCodeInput) (RedeemCode, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO redeem_codes (code_hash, code_prefix, batch_name, amount, max_uses, expires_at, created_by)
		VALUES ($1,$2,$3,$4::numeric,$5,$6,NULLIF($7,'')::uuid)
		RETURNING id::text, code_prefix, batch_name, amount::text, status, max_uses, used_count, expires_at, created_at
	`, input.CodeHash, input.CodePrefix, input.BatchName, input.Amount, input.MaxUses, input.ExpiresAt, input.CreatedBy)
	return scanRedeemCode(row)
}

func (r RedeemRepository) Disable(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE redeem_codes SET status='disabled', updated_at=now() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r RedeemRepository) Redeem(ctx context.Context, userID, codeHash, clientIP string) (RedeemCode, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return RedeemCode{}, err
	}
	defer tx.Rollback(ctx)

	var code RedeemCode
	var codeID string
	err = tx.QueryRow(ctx, `
		SELECT id::text, code_prefix, batch_name, amount::text, status, max_uses, used_count, expires_at, created_at
		FROM redeem_codes
		WHERE code_hash=$1
		FOR UPDATE
	`, codeHash).Scan(&code.ID, &code.Prefix, &code.BatchName, &code.Amount, &code.Status, &code.MaxUses, &code.UsedCount, &code.ExpiresAt, &code.CreatedAt)
	if err != nil {
		return RedeemCode{}, ErrRedeemUnavailable
	}
	codeID = code.ID
	if code.Status != "unused" || code.UsedCount >= code.MaxUses || (code.ExpiresAt != nil && code.ExpiresAt.Before(time.Now())) {
		return RedeemCode{}, ErrRedeemUnavailable
	}

	var before string
	if err := tx.QueryRow(ctx, "SELECT balance::text FROM users WHERE id=$1 FOR UPDATE", userID).Scan(&before); err != nil {
		return RedeemCode{}, err
	}
	var after string
	if err := tx.QueryRow(ctx, `
		UPDATE users
		SET balance = balance + $2::numeric, updated_at=now()
		WHERE id=$1
		RETURNING balance::text
	`, userID, code.Amount).Scan(&after); err != nil {
		return RedeemCode{}, err
	}
	var ledgerID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO balance_ledger (user_id, type, amount, balance_before, balance_after, related_type, related_id, remark)
		VALUES ($1, 'redeem', $2::numeric, $3::numeric, $4::numeric, 'redeem_code', $5, 'redeem code')
		RETURNING id::text
	`, userID, code.Amount, before, after, codeID).Scan(&ledgerID); err != nil {
		return RedeemCode{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO redeem_records (redeem_code_id, user_id, amount, ledger_id, client_ip)
		VALUES ($1,$2,$3::numeric,$4,NULLIF($5,'')::inet)
	`, codeID, userID, code.Amount, ledgerID, clientIP); err != nil {
		return RedeemCode{}, err
	}
	newStatus := "used"
	if code.MaxUses > 1 && code.UsedCount+1 < code.MaxUses {
		newStatus = "unused"
	}
	if _, err := tx.Exec(ctx, "UPDATE redeem_codes SET used_count=used_count+1, status=$2, updated_at=now() WHERE id=$1", codeID, newStatus); err != nil {
		return RedeemCode{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RedeemCode{}, err
	}
	code.UsedCount++
	code.Status = newStatus
	return code, nil
}

type redeemScanner interface {
	Scan(dest ...any) error
}

func scanRedeemCode(row redeemScanner) (RedeemCode, error) {
	var item RedeemCode
	err := row.Scan(&item.ID, &item.Prefix, &item.BatchName, &item.Amount, &item.Status, &item.MaxUses, &item.UsedCount, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}

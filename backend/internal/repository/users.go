package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	Balance       string     `json:"balance"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

type CreateUserParams struct {
	Username      string
	Email         string
	PasswordHash  string
	Role          string
	Status        string
	EmailVerified bool
}

type UpdateUserParams struct {
	ID       string
	Email    *string
	Status   *string
	Username *string
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return UserRepository{db: db}
}

func (r UserRepository) FindByUsernameOrEmail(ctx context.Context, login string) (User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE username=$1 OR email=$1
	`, login)
	return scanUser(row)
}

func (r UserRepository) FindByID(ctx context.Context, id string) (User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id=$1
	`, id)
	return scanUser(row)
}

func (r UserRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r UserRepository) ListPaged(ctx context.Context, limit, offset int) ([]User, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, rows.Err()
}

func (r UserRepository) Create(ctx context.Context, params CreateUserParams) (User, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, status, email_verified)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6)
		RETURNING id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
	`, params.Username, params.Email, params.PasswordHash, params.Role, params.Status, params.EmailVerified)
	return scanUser(row)
}

func (r UserRepository) Update(ctx context.Context, params UpdateUserParams) (User, error) {
	current, err := r.FindByID(ctx, params.ID)
	if err != nil {
		return User{}, err
	}
	username := current.Username
	email := current.Email
	status := current.Status
	if params.Username != nil {
		username = *params.Username
	}
	if params.Email != nil {
		email = *params.Email
	}
	if params.Status != nil {
		status = *params.Status
	}
	row := r.db.QueryRow(ctx, `
		UPDATE users
		SET username=$2, email=NULLIF($3, ''), status=$4, updated_at=now()
		WHERE id=$1
		RETURNING id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
	`, params.ID, username, email, status)
	return scanUser(row)
}

func (r UserRepository) SetPassword(ctx context.Context, id, passwordHash string) error {
	tag, err := r.db.Exec(ctx, "UPDATE users SET password_hash=$2, updated_at=now() WHERE id=$1", id, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r UserRepository) TouchLastLogin(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET last_login_at=now() WHERE id=$1", id)
	return err
}

func (r UserRepository) AdjustBalance(ctx context.Context, userID, operatorID, amount, ledgerType, remark string) (User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	var before string
	if err := tx.QueryRow(ctx, "SELECT balance::text FROM users WHERE id=$1 FOR UPDATE", userID).Scan(&before); err != nil {
		return User{}, err
	}

	row := tx.QueryRow(ctx, `
		UPDATE users
		SET balance = balance + $2::numeric, updated_at=now()
		WHERE id=$1 AND balance + $2::numeric >= 0
		RETURNING id::text, username, COALESCE(email, ''), password_hash, role, status, balance::text, email_verified, created_at, updated_at, last_login_at
	`, userID, amount)
	after, err := scanUser(row)
	if err != nil {
		return User{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO balance_ledger (
			user_id, type, amount, balance_before, balance_after, related_type, related_id, operator_id, remark
		)
		VALUES ($1, $2, $3::numeric, $4::numeric, $5::numeric, 'admin_adjustment', $1, NULLIF($6, '')::uuid, $7)
	`, userID, ledgerType, amount, before, after.Balance, operatorID, remark); err != nil {
		return User{}, err
	}

	return after, tx.Commit(ctx)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.Balance,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, err
	}
	return user, err
}

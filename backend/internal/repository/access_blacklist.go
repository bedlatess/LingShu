package repository

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccessBlacklistRepository struct {
	db *pgxpool.Pool
}

type AccessBlacklistEntry struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	Value      string     `json:"value"`
	Scope      string     `json:"scope"`
	Reason     string     `json:"reason"`
	Source     string     `json:"source"`
	Active     bool       `json:"active"`
	CreatedBy  string     `json:"created_by,omitempty"`
	ReleasedBy string     `json:"released_by,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	ReleasedAt *time.Time `json:"released_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type AccessBlacklistFilter struct {
	Kind   string
	Scope  string
	Active string
	Query  string
}

type CreateAccessBlacklistInput struct {
	Kind      string
	Value     string
	Scope     string
	Reason    string
	Source    string
	CreatedBy string
	ExpiresAt *time.Time
}

func NewAccessBlacklistRepository(db *pgxpool.Pool) AccessBlacklistRepository {
	return AccessBlacklistRepository{db: db}
}

func (r AccessBlacklistRepository) ListPaged(ctx context.Context, filter AccessBlacklistFilter, limit, offset int) ([]AccessBlacklistEntry, int, error) {
	filter.Kind = strings.TrimSpace(filter.Kind)
	filter.Scope = strings.TrimSpace(filter.Scope)
	filter.Active = strings.TrimSpace(filter.Active)
	filter.Query = strings.TrimSpace(filter.Query)

	var total int
	if err := r.db.QueryRow(ctx, `
		SELECT count(*)::int
		FROM access_blacklist
		WHERE ($1 = '' OR kind = $1)
		  AND ($2 = '' OR scope = $2)
		  AND ($3 = '' OR active = ($3 = 'true'))
		  AND ($4 = '' OR value ILIKE '%' || $4 || '%' OR reason ILIKE '%' || $4 || '%')
	`, filter.Kind, filter.Scope, filter.Active, filter.Query).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, kind, value, scope, reason, source, active,
		       COALESCE(created_by::text, ''), COALESCE(released_by::text, ''),
		       expires_at, released_at, created_at, updated_at
		FROM access_blacklist
		WHERE ($1 = '' OR kind = $1)
		  AND ($2 = '' OR scope = $2)
		  AND ($3 = '' OR active = ($3 = 'true'))
		  AND ($4 = '' OR value ILIKE '%' || $4 || '%' OR reason ILIKE '%' || $4 || '%')
		ORDER BY active DESC, created_at DESC
		LIMIT $5 OFFSET $6
	`, filter.Kind, filter.Scope, filter.Active, filter.Query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []AccessBlacklistEntry{}
	for rows.Next() {
		item, err := scanAccessBlacklistEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r AccessBlacklistRepository) Active(ctx context.Context) ([]AccessBlacklistEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, kind, value, scope, reason, source, active,
		       COALESCE(created_by::text, ''), COALESCE(released_by::text, ''),
		       expires_at, released_at, created_at, updated_at
		FROM access_blacklist
		WHERE active = true
		  AND (expires_at IS NULL OR expires_at > now())
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []AccessBlacklistEntry{}
	for rows.Next() {
		item, err := scanAccessBlacklistEntry(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r AccessBlacklistRepository) Create(ctx context.Context, input CreateAccessBlacklistInput) (AccessBlacklistEntry, error) {
	input.Kind = strings.TrimSpace(input.Kind)
	input.Value = strings.TrimSpace(input.Value)
	input.Scope = strings.TrimSpace(input.Scope)
	input.Reason = strings.TrimSpace(input.Reason)
	input.Source = strings.TrimSpace(input.Source)
	if input.Scope == "" {
		input.Scope = "all"
	}
	if input.Source == "" {
		input.Source = "manual"
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO access_blacklist (kind, value, scope, reason, source, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::uuid, $7)
		ON CONFLICT (kind, value, scope) WHERE active = true
		DO UPDATE SET reason=EXCLUDED.reason, source=EXCLUDED.source, expires_at=EXCLUDED.expires_at, updated_at=now()
		RETURNING id::text, kind, value, scope, reason, source, active,
		          COALESCE(created_by::text, ''), COALESCE(released_by::text, ''),
		          expires_at, released_at, created_at, updated_at
	`, input.Kind, input.Value, input.Scope, input.Reason, input.Source, input.CreatedBy, input.ExpiresAt)
	return scanAccessBlacklistEntry(row)
}

func (r AccessBlacklistRepository) Release(ctx context.Context, id, actorID string) (AccessBlacklistEntry, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE access_blacklist
		SET active=false, released_by=NULLIF($2, '')::uuid, released_at=now(), updated_at=now()
		WHERE id=$1 AND active=true
		RETURNING id::text, kind, value, scope, reason, source, active,
		          COALESCE(created_by::text, ''), COALESCE(released_by::text, ''),
		          expires_at, released_at, created_at, updated_at
	`, id, actorID)
	item, err := scanAccessBlacklistEntry(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccessBlacklistEntry{}, errors.New("blacklist entry not found or already released")
	}
	return item, err
}

func (r AccessBlacklistRepository) Matches(ctx context.Context, scope, ip, deviceID string) (AccessBlacklistEntry, bool, error) {
	items, err := r.Active(ctx)
	if err != nil {
		return AccessBlacklistEntry{}, false, err
	}
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	for _, item := range items {
		if item.Scope != "all" && item.Scope != scope {
			continue
		}
		switch item.Kind {
		case "device":
			if deviceID != "" && strings.EqualFold(item.Value, deviceID) {
				return item, true, nil
			}
		case "ip":
			if parsedIP != nil && parsedIP.Equal(net.ParseIP(item.Value)) {
				return item, true, nil
			}
		case "cidr":
			_, network, err := net.ParseCIDR(item.Value)
			if err == nil && parsedIP != nil && network.Contains(parsedIP) {
				return item, true, nil
			}
		}
	}
	return AccessBlacklistEntry{}, false, nil
}

type accessBlacklistScanner interface {
	Scan(dest ...any) error
}

func scanAccessBlacklistEntry(row accessBlacklistScanner) (AccessBlacklistEntry, error) {
	var item AccessBlacklistEntry
	err := row.Scan(
		&item.ID,
		&item.Kind,
		&item.Value,
		&item.Scope,
		&item.Reason,
		&item.Source,
		&item.Active,
		&item.CreatedBy,
		&item.ReleasedBy,
		&item.ExpiresAt,
		&item.ReleasedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Announcement struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    string     `json:"status"`
	Priority  int        `json:"priority"`
	Pinned    bool       `json:"pinned"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AnnouncementInput struct {
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    string     `json:"status"`
	Priority  int        `json:"priority"`
	Pinned    bool       `json:"pinned"`
	PublishAt *time.Time `json:"publish_at"`
	ExpireAt  *time.Time `json:"expire_at"`
}

type AnnouncementRepository struct {
	db *pgxpool.Pool
}

func NewAnnouncementRepository(db *pgxpool.Pool) AnnouncementRepository {
	return AnnouncementRepository{db: db}
}

func (r AnnouncementRepository) ListAdmin(ctx context.Context) ([]Announcement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, title, content, status, priority, pinned, publish_at, expire_at, created_at, updated_at
		FROM announcements
		ORDER BY pinned DESC, priority DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r AnnouncementRepository) ListOnline(ctx context.Context) ([]Announcement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, title, content, status, priority, pinned, publish_at, expire_at, created_at, updated_at
		FROM announcements
		WHERE status='online'
		  AND (publish_at IS NULL OR publish_at <= now())
		  AND (expire_at IS NULL OR expire_at > now())
		ORDER BY pinned DESC, priority DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r AnnouncementRepository) Create(ctx context.Context, input AnnouncementInput, actorID string) (Announcement, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO announcements (title, content, status, priority, pinned, publish_at, expire_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,'')::uuid)
		RETURNING id::text, title, content, status, priority, pinned, publish_at, expire_at, created_at, updated_at
	`, input.Title, input.Content, input.Status, input.Priority, input.Pinned, input.PublishAt, input.ExpireAt, actorID)
	return scanAnnouncement(row)
}

func (r AnnouncementRepository) Update(ctx context.Context, id string, input AnnouncementInput) (Announcement, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE announcements
		SET title=$2, content=$3, status=$4, priority=$5, pinned=$6, publish_at=$7, expire_at=$8, updated_at=now()
		WHERE id=$1
		RETURNING id::text, title, content, status, priority, pinned, publish_at, expire_at, created_at, updated_at
	`, id, input.Title, input.Content, input.Status, input.Priority, input.Pinned, input.PublishAt, input.ExpireAt)
	return scanAnnouncement(row)
}

func (r AnnouncementRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM announcements WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type announcementScanner interface {
	Scan(dest ...any) error
}

func scanAnnouncement(row announcementScanner) (Announcement, error) {
	var item Announcement
	err := row.Scan(&item.ID, &item.Title, &item.Content, &item.Status, &item.Priority, &item.Pinned, &item.PublishAt, &item.ExpireAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func scanAnnouncements(rows pgx.Rows) ([]Announcement, error) {
	items := []Announcement{}
	for rows.Next() {
		item, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

package service

import (
	"context"
	"errors"

	"lingshu/backend/internal/repository"
)

type AnnouncementService struct {
	announcements repository.AnnouncementRepository
	audits        repository.AuditRepository
}

func NewAnnouncementService(announcements repository.AnnouncementRepository, audits repository.AuditRepository) AnnouncementService {
	return AnnouncementService{announcements: announcements, audits: audits}
}

func (s AnnouncementService) ListAdmin(ctx context.Context) ([]repository.Announcement, error) {
	return s.announcements.ListAdmin(ctx)
}

func (s AnnouncementService) ListAdminPaged(ctx context.Context, page, limit int) ([]repository.Announcement, int, error) {
	return s.announcements.ListAdminPaged(ctx, limit, (page-1)*limit)
}

func (s AnnouncementService) ListOnline(ctx context.Context) ([]repository.Announcement, error) {
	return s.announcements.ListOnline(ctx)
}

func (s AnnouncementService) Create(ctx context.Context, actorID string, input repository.AnnouncementInput, ip, userAgent string) (repository.Announcement, error) {
	if input.Title == "" || input.Content == "" {
		return repository.Announcement{}, errors.New("title and content are required")
	}
	if input.Status == "" {
		input.Status = "offline"
	}
	item, err := s.announcements.Create(ctx, input, actorID)
	if err != nil {
		return repository.Announcement{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.announcement.create", TargetType: "announcement", TargetID: item.ID, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func (s AnnouncementService) Update(ctx context.Context, actorID, id string, input repository.AnnouncementInput, ip, userAgent string) (repository.Announcement, error) {
	if input.Title == "" || input.Content == "" {
		return repository.Announcement{}, errors.New("title and content are required")
	}
	item, err := s.announcements.Update(ctx, id, input)
	if err != nil {
		return repository.Announcement{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.announcement.update", TargetType: "announcement", TargetID: id, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func (s AnnouncementService) Delete(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.announcements.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.announcement.delete", TargetType: "announcement", TargetID: id, IP: ip, UserAgent: userAgent})
	return nil
}

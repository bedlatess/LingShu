package service

import (
	"context"
	"errors"
	"strings"

	"lingshu/backend/internal/repository"
)

type SettingsService struct {
	settings repository.SettingsRepository
	audits   repository.AuditRepository
}

func NewSettingsService(settings repository.SettingsRepository, audits repository.AuditRepository) SettingsService {
	return SettingsService{settings: settings, audits: audits}
}

func (s SettingsService) List(ctx context.Context) ([]repository.Setting, error) {
	return s.settings.List(ctx)
}

func (s SettingsService) ListPaged(ctx context.Context, page, limit int) ([]repository.Setting, int, error) {
	return s.settings.ListPaged(ctx, limit, (page-1)*limit)
}

func (s SettingsService) AuditLogsPaged(ctx context.Context, filter repository.AuditLogFilter, page, limit int) ([]repository.AuditLog, int, error) {
	return s.audits.ListPagedFiltered(ctx, filter, limit, (page-1)*limit)
}

func (s SettingsService) Patch(ctx context.Context, actorID string, updates []repository.SettingUpdate, ip, userAgent string) ([]repository.Setting, error) {
	if len(updates) == 0 {
		return nil, errors.New("updates are required")
	}
	seen := map[string]struct{}{}
	for i := range updates {
		updates[i].Key = strings.TrimSpace(updates[i].Key)
		if updates[i].Key == "" {
			return nil, errors.New("setting key is required")
		}
		if _, ok := seen[updates[i].Key]; ok {
			return nil, errors.New("duplicate setting key")
		}
		seen[updates[i].Key] = struct{}{}
	}
	before, _ := s.settings.List(ctx)
	existing := map[string]struct{}{}
	for _, item := range before {
		existing[item.Key] = struct{}{}
	}
	for _, update := range updates {
		if _, ok := existing[update.Key]; !ok {
			return nil, errors.New("unknown setting key: " + update.Key)
		}
	}
	after, err := s.settings.Patch(ctx, actorID, updates)
	if err != nil {
		return nil, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.settings.patch",
		TargetType: "system_settings",
		Before:     before,
		After:      updates,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return after, nil
}

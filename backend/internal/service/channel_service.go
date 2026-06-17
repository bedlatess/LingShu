package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	secret "lingshu/backend/internal/pkg/crypto"
	"lingshu/backend/internal/repository"
)

type ChannelService struct {
	channels repository.ChannelRepository
	audits   repository.AuditRepository
}

func NewChannelService(channels repository.ChannelRepository, audits repository.AuditRepository) ChannelService {
	return ChannelService{channels: channels, audits: audits}
}

func (s ChannelService) List(ctx context.Context) ([]repository.Channel, error) {
	return s.channels.List(ctx)
}

func (s ChannelService) Create(ctx context.Context, actorID string, input repository.ChannelInput, ip, userAgent string) (repository.Channel, error) {
	input = normalizeChannel(input)
	if err := validateChannel(input, true); err != nil {
		return repository.Channel{}, err
	}
	item, err := s.channels.Create(ctx, input, secret.Protect(input.APIKey))
	if err != nil {
		return repository.Channel{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.create", TargetType: "channel", TargetID: item.ID, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func (s ChannelService) Update(ctx context.Context, actorID, id string, input repository.ChannelInput, ip, userAgent string) (repository.Channel, error) {
	input = normalizeChannel(input)
	if err := validateChannel(input, false); err != nil {
		return repository.Channel{}, err
	}
	protected := ""
	if input.APIKey != "" {
		protected = secret.Protect(input.APIKey)
	}
	item, err := s.channels.Update(ctx, id, input, protected)
	if err != nil {
		return repository.Channel{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.update", TargetType: "channel", TargetID: id, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func (s ChannelService) Disable(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.channels.Disable(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.disable", TargetType: "channel", TargetID: id, IP: ip, UserAgent: userAgent})
	return nil
}

func (s ChannelService) Test(ctx context.Context, id, baseURL string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	url := strings.TrimRight(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = s.channels.MarkTest(ctx, id, false, err.Error())
		return map[string]any{"ok": false, "message": err.Error()}, nil
	}
	defer resp.Body.Close()
	ok := resp.StatusCode >= 200 && resp.StatusCode < 500
	message := resp.Status
	_ = s.channels.MarkTest(ctx, id, ok, message)
	return map[string]any{"ok": ok, "status": resp.StatusCode, "message": message}, nil
}

func (s ChannelService) BindModel(ctx context.Context, actorID string, input repository.BindChannelModelInput, ip, userAgent string) (repository.ChannelModelBinding, error) {
	if input.ChannelID == "" || input.ModelID == "" || input.UpstreamModelName == "" {
		return repository.ChannelModelBinding{}, errors.New("channel_id, model_id and upstream_model_name are required")
	}
	item, err := s.channels.BindModel(ctx, input)
	if err != nil {
		return repository.ChannelModelBinding{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.bind_model", TargetType: "channel_model", TargetID: item.ID, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func validateChannel(input repository.ChannelInput, requireKey bool) error {
	if input.Name == "" || input.BaseURL == "" {
		return errors.New("name and base_url are required")
	}
	if input.ProviderType != "openai" && input.ProviderType != "claude" && input.ProviderType != "gemini" && input.ProviderType != "custom" {
		return errors.New("invalid provider_type")
	}
	if requireKey && input.APIKey == "" {
		return errors.New("api_key is required")
	}
	return nil
}

func normalizeChannel(input repository.ChannelInput) repository.ChannelInput {
	if input.ProviderType == "" {
		input.ProviderType = "openai"
	}
	if input.Status == "" {
		input.Status = "enabled"
	}
	if input.Weight <= 0 {
		input.Weight = 1
	}
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 120
	}
	if input.RPMLimit <= 0 {
		input.RPMLimit = 60
	}
	if input.ConcurrencyLimit <= 0 {
		input.ConcurrencyLimit = 5
	}
	if input.FailThreshold <= 0 {
		input.FailThreshold = 5
	}
	return input
}

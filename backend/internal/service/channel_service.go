package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	secret "lingshu/backend/internal/pkg/crypto"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/upstream"
)

type ChannelService struct {
	channels repository.ChannelRepository
	audits   repository.AuditRepository
}

type SyncChannelModelsResult struct {
	UpstreamModels   []upstream.ProviderModel          `json:"upstream_models"`
	ExistingBindings []repository.ChannelDetailBinding `json:"existing_bindings"`
}

type ImportChannelModelsInput struct {
	Strategy string                               `json:"strategy"`
	Models   []repository.ImportChannelModelInput `json:"models"`
}

func NewChannelService(channels repository.ChannelRepository, audits repository.AuditRepository) ChannelService {
	return ChannelService{channels: channels, audits: audits}
}

func (s ChannelService) List(ctx context.Context) ([]repository.Channel, error) {
	return s.channels.List(ctx)
}

func (s ChannelService) ListPaged(ctx context.Context, page, limit int) ([]repository.Channel, int, error) {
	return s.channels.ListPaged(ctx, limit, (page-1)*limit)
}

func (s ChannelService) Detail(ctx context.Context, id string) (repository.ChannelDetail, error) {
	return s.channels.Detail(ctx, id)
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

func (s ChannelService) Delete(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.channels.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.delete", TargetType: "channel", TargetID: id, IP: ip, UserAgent: userAgent})
	return nil
}

func (s ChannelService) Test(ctx context.Context, id, baseURL string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	channel, err := s.channels.FindSecretByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = channel.BaseURL
	}
	key := secret.Unprotect(channel.APIKeyEncrypted)
	start := time.Now()
	var statusCode int
	var models []string
	if strings.EqualFold(channel.ProviderType, "anthropic") || strings.EqualFold(channel.ProviderType, "claude") {
		models, statusCode, err = upstream.ProbeAnthropic(ctx, baseURL, key)
	} else {
		models, statusCode, err = upstream.ProbeOpenAI(ctx, baseURL, key)
		if err != nil && !strings.Contains(strings.ToLower(baseURL), "/v") {
			models, statusCode, err = upstream.ProbeOpenAI(ctx, strings.TrimRight(baseURL, "/")+"/v1", key)
		}
	}
	latency := time.Since(start).Milliseconds()
	if err != nil {
		category := "network_error"
		if errors.Is(err, context.DeadlineExceeded) {
			category = "timeout"
		} else if statusCode > 0 {
			category = categorizeChannelTest(statusCode)
		}
		_ = s.channels.MarkTest(ctx, id, false, err.Error(), latency)
		result := map[string]any{"ok": false, "category": category, "message": err.Error(), "latency_ms": latency}
		if statusCode > 0 {
			result["status"] = statusCode
		}
		return result, nil
	}

	message := fmt.Sprintf("探测成功，样例模型 %d 个", len(models))
	category := categorizeChannelTest(statusCode)
	_ = s.channels.MarkTest(ctx, id, true, message, latency)
	return map[string]any{"ok": true, "status": statusCode, "category": category, "message": message, "latency_ms": latency}, nil
}

func categorizeChannelTest(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "ok"
	case status == http.StatusBadRequest:
		return "bad_request"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "auth"
	case status == http.StatusNotFound:
		return "not_found"
	case status == http.StatusTooManyRequests:
		return "rate_limit"
	case status == http.StatusInternalServerError || status == http.StatusBadGateway || status == http.StatusServiceUnavailable:
		return "server_error"
	case status == 522 || status == 524:
		return "upstream_blocked"
	default:
		return "unknown"
	}
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

func (s ChannelService) SyncModels(ctx context.Context, channelID string) (SyncChannelModelsResult, error) {
	channel, err := s.channels.FindSecretByID(ctx, channelID)
	if err != nil {
		return SyncChannelModelsResult{}, err
	}
	models, err := upstream.ProviderForType(channel.ProviderType).ListModels(ctx, channel.BaseURL, secret.Unprotect(channel.APIKeyEncrypted))
	if err != nil {
		return SyncChannelModelsResult{}, err
	}
	detail, err := s.channels.Detail(ctx, channelID)
	if err != nil {
		return SyncChannelModelsResult{}, err
	}
	return SyncChannelModelsResult{UpstreamModels: models, ExistingBindings: detail.Models}, nil
}

func (s ChannelService) ImportModels(ctx context.Context, actorID, channelID string, input ImportChannelModelsInput, ip, userAgent string) ([]repository.ImportChannelModelResult, error) {
	if input.Strategy == "" {
		input.Strategy = "create_or_bind"
	}
	if input.Strategy != "create_or_bind" && input.Strategy != "bind_existing" {
		return nil, errors.New("invalid import strategy")
	}
	normalized := make([]repository.ImportChannelModelInput, 0, len(input.Models))
	for _, item := range input.Models {
		item.UpstreamName = strings.TrimSpace(item.UpstreamName)
		item.PublicName = strings.TrimSpace(item.PublicName)
		if item.PublicName == "" {
			item.PublicName = item.UpstreamName
		}
		if item.Type == "" {
			item.Type = "chat"
		}
		if item.BillingMode == "" {
			item.BillingMode = "token"
		}
		if item.InputPricePer1K == "" {
			item.InputPricePer1K = "0"
		}
		if item.OutputPricePer1K == "" {
			item.OutputPricePer1K = "0"
		}
		if item.PricePerCall == "" {
			item.PricePerCall = "0"
		}
		if item.RateMultiplier == "" {
			item.RateMultiplier = "1.200"
		}
		if item.Status == "" {
			item.Status = "enabled"
		}
		if item.UpstreamName == "" || item.PublicName == "" {
			return nil, errors.New("upstream_name and public_name are required")
		}
		if item.Type != "chat" && item.Type != "embedding" && item.Type != "image" && item.Type != "video" {
			return nil, errors.New("invalid model type")
		}
		if item.BillingMode != "token" && item.BillingMode != "per_call" {
			return nil, errors.New("invalid billing mode")
		}
		if input.Strategy == "bind_existing" {
			item.BindExistingOnly = true
			item.InputPricePer1K = "0"
			item.OutputPricePer1K = "0"
			item.PricePerCall = "0"
		}
		normalized = append(normalized, item)
	}
	if len(normalized) == 0 {
		return nil, errors.New("models are required")
	}
	results, err := s.channels.ImportModels(ctx, channelID, normalized)
	if err != nil {
		return nil, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.import_models", TargetType: "channel", TargetID: channelID, After: results, IP: ip, UserAgent: userAgent})
	return results, nil
}

func (s ChannelService) UnbindModel(ctx context.Context, actorID, channelID, modelID, ip, userAgent string) error {
	if channelID == "" || modelID == "" {
		return errors.New("channel_id and model_id are required")
	}
	if err := s.channels.UnbindModel(ctx, channelID, modelID); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.channel.unbind_model", TargetType: "channel_model", TargetID: modelID, After: map[string]string{"channel_id": channelID, "model_id": modelID}, IP: ip, UserAgent: userAgent})
	return nil
}

func validateChannel(input repository.ChannelInput, requireKey bool) error {
	if input.Name == "" || input.BaseURL == "" {
		return errors.New("name and base_url are required")
	}
	if input.ProviderType != "openai" && input.ProviderType != "anthropic" {
		return errors.New("invalid provider_type")
	}
	if requireKey && input.APIKey == "" {
		return errors.New("api_key is required")
	}
	return nil
}

func normalizeChannel(input repository.ChannelInput) repository.ChannelInput {
	input.ProviderType = strings.ToLower(strings.TrimSpace(input.ProviderType))
	if input.ProviderType == "" {
		input.ProviderType = "openai"
	}
	if input.ProviderType == "claude" {
		input.ProviderType = "anthropic"
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

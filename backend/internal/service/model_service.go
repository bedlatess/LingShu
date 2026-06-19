package service

import (
	"context"
	"errors"

	"lingshu/backend/internal/repository"
)

type ModelService struct {
	models repository.ModelRepository
	audits repository.AuditRepository
}

func NewModelService(models repository.ModelRepository, audits repository.AuditRepository) ModelService {
	return ModelService{models: models, audits: audits}
}

func (s ModelService) List(ctx context.Context) ([]repository.Model, error) {
	return s.models.List(ctx)
}

func (s ModelService) ListPaged(ctx context.Context, page, limit int) ([]repository.Model, int, error) {
	return s.models.ListPaged(ctx, limit, (page-1)*limit)
}

func (s ModelService) Detail(ctx context.Context, id string) (repository.ModelDetail, error) {
	return s.models.Detail(ctx, id)
}

func (s ModelService) Create(ctx context.Context, actorID string, input repository.ModelInput, ip, userAgent string) (repository.Model, error) {
	input = normalizeModel(input)
	if err := validateModel(input); err != nil {
		return repository.Model{}, err
	}
	item, err := s.models.Create(ctx, input)
	if err != nil {
		return repository.Model{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.model.create", TargetType: "model", TargetID: item.ID, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func (s ModelService) Update(ctx context.Context, actorID, id string, input repository.ModelInput, ip, userAgent string) (repository.Model, error) {
	input = normalizeModel(input)
	if err := validateModel(input); err != nil {
		return repository.Model{}, err
	}
	item, err := s.models.Update(ctx, id, input)
	if err != nil {
		return repository.Model{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.model.update", TargetType: "model", TargetID: id, After: item, IP: ip, UserAgent: userAgent})
	return item, nil
}

func (s ModelService) Disable(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.models.Disable(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.model.disable", TargetType: "model", TargetID: id, IP: ip, UserAgent: userAgent})
	return nil
}

func (s ModelService) Delete(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.models.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.model.delete", TargetType: "model", TargetID: id, IP: ip, UserAgent: userAgent})
	return nil
}

func validateModel(input repository.ModelInput) error {
	if input.PublicName == "" {
		return errors.New("public_name is required")
	}
	if input.Type != "chat" && input.Type != "embedding" && input.Type != "image" && input.Type != "video" {
		return errors.New("invalid model type")
	}
	if input.BillingMode != "token" && input.BillingMode != "per_call" {
		return errors.New("invalid billing mode")
	}
	if input.RateMultiplier == "" {
		return errors.New("rate_multiplier is required")
	}
	return nil
}

func normalizeModel(input repository.ModelInput) repository.ModelInput {
	if input.Type == "" {
		input.Type = "chat"
	}
	if input.BillingMode == "" {
		input.BillingMode = "token"
	}
	if input.InputPricePer1K == "" {
		input.InputPricePer1K = "0"
	}
	if input.OutputPricePer1K == "" {
		input.OutputPricePer1K = "0"
	}
	if input.CacheCreationPricePer1K == "" {
		input.CacheCreationPricePer1K = "0"
	}
	if input.CacheReadPricePer1K == "" {
		input.CacheReadPricePer1K = "0"
	}
	if input.PricePerCall == "" {
		input.PricePerCall = "0"
	}
	if input.RateMultiplier == "" {
		input.RateMultiplier = "1.200"
	}
	if input.Status == "" {
		input.Status = "enabled"
	}
	return input
}

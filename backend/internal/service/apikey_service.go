package service

import (
	"context"
	"errors"
	"strings"

	"lingshu/backend/internal/config"
	"lingshu/backend/internal/pkg/apikey"
	"lingshu/backend/internal/repository"
)

type APIKeyService struct {
	cfg    config.Config
	keys   repository.APIKeyRepository
	audits repository.AuditRepository
}

type CreateAPIKeyInput struct {
	UserID           string   `json:"user_id"`
	Name             string   `json:"name"`
	AllowedEndpoints []string `json:"allowed_endpoints"`
}

type CreatedAPIKey struct {
	repository.APIKey
	Plaintext string `json:"plaintext"`
}

func NewAPIKeyService(cfg config.Config, keys repository.APIKeyRepository, audits repository.AuditRepository) APIKeyService {
	return APIKeyService{cfg: cfg, keys: keys, audits: audits}
}

func (s APIKeyService) ListForUser(ctx context.Context, userID string) ([]repository.APIKey, error) {
	return s.keys.ListByUser(ctx, userID)
}

func (s APIKeyService) ListForUserPaged(ctx context.Context, userID string, page, limit int) ([]repository.APIKey, int, error) {
	return s.keys.ListByUserPaged(ctx, userID, limit, (page-1)*limit)
}

func (s APIKeyService) ListAll(ctx context.Context) ([]repository.APIKey, error) {
	return s.keys.ListAll(ctx)
}

func (s APIKeyService) ListAllPaged(ctx context.Context, page, limit int) ([]repository.APIKey, int, error) {
	return s.keys.ListAllPaged(ctx, limit, (page-1)*limit)
}

func (s APIKeyService) Create(ctx context.Context, actorID string, input CreateAPIKeyInput, ip, userAgent string) (CreatedAPIKey, error) {
	if input.UserID == "" || input.Name == "" {
		return CreatedAPIKey{}, errors.New("user_id and name are required")
	}
	endpoints, err := NormalizeAllowedEndpoints(input.AllowedEndpoints)
	if err != nil {
		return CreatedAPIKey{}, err
	}
	plain, err := apikey.Generate(s.cfg.APIKeyPrefix)
	if err != nil {
		return CreatedAPIKey{}, err
	}
	created, err := s.keys.Create(ctx, repository.CreateAPIKeyParams{
		UserID:           input.UserID,
		KeyPrefix:        apikey.Prefix(plain),
		KeyHash:          apikey.Hash(plain),
		Mask:             apikey.Mask(plain),
		Name:             input.Name,
		AllowedEndpoints: endpoints,
	})
	if err != nil {
		return CreatedAPIKey{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.apikey.create",
		TargetType: "api_key",
		TargetID:   created.ID,
		After:      created,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return CreatedAPIKey{APIKey: created, Plaintext: plain}, nil
}

func (s APIKeyService) CreateForUser(ctx context.Context, userID, name string, allowedEndpoints []string) (CreatedAPIKey, error) {
	return s.createOwned(ctx, userID, name, allowedEndpoints)
}

func (s APIKeyService) createOwned(ctx context.Context, userID, name string, allowedEndpoints []string) (CreatedAPIKey, error) {
	if userID == "" || name == "" {
		return CreatedAPIKey{}, errors.New("name is required")
	}
	endpoints, err := NormalizeAllowedEndpoints(allowedEndpoints)
	if err != nil {
		return CreatedAPIKey{}, err
	}
	plain, err := apikey.Generate(s.cfg.APIKeyPrefix)
	if err != nil {
		return CreatedAPIKey{}, err
	}
	created, err := s.keys.Create(ctx, repository.CreateAPIKeyParams{
		UserID:           userID,
		KeyPrefix:        apikey.Prefix(plain),
		KeyHash:          apikey.Hash(plain),
		Mask:             apikey.Mask(plain),
		Name:             name,
		AllowedEndpoints: endpoints,
	})
	if err != nil {
		return CreatedAPIKey{}, err
	}
	return CreatedAPIKey{APIKey: created, Plaintext: plain}, nil
}

func (s APIKeyService) UpdateForUser(ctx context.Context, userID, id, name, status string, allowedEndpoints []string, updateAllowedEndpoints bool) (repository.APIKey, error) {
	if status != "" && status != "active" && status != "disabled" {
		return repository.APIKey{}, errors.New("invalid status")
	}
	endpoints := []string{}
	if updateAllowedEndpoints {
		var err error
		endpoints, err = NormalizeAllowedEndpoints(allowedEndpoints)
		if err != nil {
			return repository.APIKey{}, err
		}
	}
	return s.keys.UpdateForUser(ctx, repository.UpdateAPIKeyParams{ID: id, UserID: userID, Name: name, Status: status, AllowedEndpoints: endpoints, UpdateAllowedEndpoints: updateAllowedEndpoints})
}

func (s APIKeyService) DeleteForUser(ctx context.Context, userID, id string) error {
	return s.keys.DeleteForUser(ctx, id, userID)
}

func (s APIKeyService) Disable(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.keys.UpdateStatus(ctx, id, "disabled"); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.apikey.disable",
		TargetType: "api_key",
		TargetID:   id,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return nil
}

func (s APIKeyService) Delete(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.keys.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.apikey.delete",
		TargetType: "api_key",
		TargetID:   id,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return nil
}

var allowedGatewayEndpoints = map[string]struct{}{
	"/v1/models":           {},
	"/v1/chat/completions": {},
	"/v1/messages":         {},
	"/v1/embeddings":       {},
}

func NormalizeAllowedEndpoints(values []string) ([]string, error) {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		endpoint := strings.TrimSpace(value)
		if endpoint == "" {
			continue
		}
		if !strings.HasPrefix(endpoint, "/") {
			endpoint = "/" + endpoint
		}
		endpoint = strings.TrimRight(endpoint, "/")
		if endpoint == "" {
			continue
		}
		if _, ok := allowedGatewayEndpoints[endpoint]; !ok {
			return nil, errors.New("invalid allowed endpoint")
		}
		if _, ok := seen[endpoint]; ok {
			continue
		}
		seen[endpoint] = struct{}{}
		normalized = append(normalized, endpoint)
	}
	return normalized, nil
}

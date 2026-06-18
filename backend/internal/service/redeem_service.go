package service

import (
	"context"
	"errors"
	"time"

	redeemutil "lingshu/backend/internal/pkg/redeem"
	"lingshu/backend/internal/repository"
)

type RedeemService struct {
	redeems repository.RedeemRepository
	audits  repository.AuditRepository
}

type CreateRedeemInput struct {
	Amount    string     `json:"amount"`
	Count     int        `json:"count"`
	BatchName string     `json:"batch_name"`
	MaxUses   int        `json:"max_uses"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func NewRedeemService(redeems repository.RedeemRepository, audits repository.AuditRepository) RedeemService {
	return RedeemService{redeems: redeems, audits: audits}
}

func (s RedeemService) List(ctx context.Context) ([]repository.RedeemCode, error) {
	return s.redeems.List(ctx)
}

func (s RedeemService) ListPaged(ctx context.Context, page, limit int) ([]repository.RedeemCode, int, error) {
	return s.redeems.ListPaged(ctx, limit, (page-1)*limit)
}

func (s RedeemService) Create(ctx context.Context, actorID string, input CreateRedeemInput, ip, userAgent string) ([]repository.RedeemCode, error) {
	if input.Amount == "" {
		return nil, errors.New("amount is required")
	}
	if input.Count <= 0 {
		input.Count = 1
	}
	if input.MaxUses <= 0 {
		input.MaxUses = 1
	}
	items := make([]repository.RedeemCode, 0, input.Count)
	for i := 0; i < input.Count; i++ {
		code, err := redeemutil.GenerateCode()
		if err != nil {
			return nil, err
		}
		item, err := s.redeems.Create(ctx, repository.CreateRedeemCodeInput{
			CodeHash:   redeemutil.Hash(code),
			CodePlain:  code,
			CodePrefix: redeemutil.Prefix(code),
			BatchName:  input.BatchName,
			Amount:     input.Amount,
			MaxUses:    input.MaxUses,
			ExpiresAt:  input.ExpiresAt,
			CreatedBy:  actorID,
		})
		if err != nil {
			return nil, err
		}
		item.Code = code
		items = append(items, item)
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.redeem.create", TargetType: "redeem_code", After: map[string]any{"count": len(items), "amount": input.Amount}, IP: ip, UserAgent: userAgent})
	return items, nil
}

func (s RedeemService) Disable(ctx context.Context, actorID, id, ip, userAgent string) error {
	if err := s.redeems.Disable(ctx, id); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{ActorID: actorID, Action: "admin.redeem.disable", TargetType: "redeem_code", TargetID: id, IP: ip, UserAgent: userAgent})
	return nil
}

func (s RedeemService) Redeem(ctx context.Context, userID, code, clientIP string) (repository.RedeemCode, error) {
	return s.redeems.Redeem(ctx, userID, redeemutil.Hash(code), clientIP)
}

func (s RedeemService) Records(ctx context.Context, codeID string) ([]repository.RedeemRecord, error) {
	return s.redeems.Records(ctx, codeID)
}

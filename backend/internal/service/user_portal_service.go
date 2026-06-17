package service

import (
	"context"

	redisstore "lingshu/backend/internal/redis"
	"lingshu/backend/internal/repository"
)

type UserPortalService struct {
	users   repository.UserRepository
	models  repository.ModelRepository
	reports repository.ReportRepository
	frozen  redisstore.FrozenStore
}

type UserModelPrice struct {
	ID               string `json:"id"`
	PublicName       string `json:"public_name"`
	Type             string `json:"type"`
	Group            string `json:"group"`
	BillingMode      string `json:"billing_mode"`
	InputPricePer1K  string `json:"input_price_per_1k"`
	OutputPricePer1K string `json:"output_price_per_1k"`
	PricePerCall     string `json:"price_per_call"`
	RateMultiplier   string `json:"rate_multiplier"`
	InputUnitPrice   string `json:"input_unit_price"`
	OutputUnitPrice  string `json:"output_unit_price"`
	CallUnitPrice    string `json:"call_unit_price"`
	Status           string `json:"status"`
}

type UserDashboard struct {
	Balance         string `json:"balance"`
	TodayCharge     string `json:"today_charge"`
	MonthCharge     string `json:"month_charge"`
	Frozen          string `json:"frozen"`
	AvailableModels int    `json:"available_models"`
	TodayRequests   int    `json:"today_requests"`
}

func NewUserPortalService(users repository.UserRepository, models repository.ModelRepository, reports repository.ReportRepository, frozen redisstore.FrozenStore) UserPortalService {
	return UserPortalService{users: users, models: models, reports: reports, frozen: frozen}
}

func (s UserPortalService) Models(ctx context.Context) ([]UserModelPrice, error) {
	items, err := s.models.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserModelPrice, 0, len(items))
	for _, item := range items {
		if item.Status != "enabled" {
			continue
		}
		out = append(out, UserModelPrice{
			ID:               item.ID,
			PublicName:       item.PublicName,
			Type:             item.Type,
			Group:            item.Group,
			BillingMode:      item.BillingMode,
			InputPricePer1K:  item.InputPricePer1K,
			OutputPricePer1K: item.OutputPricePer1K,
			PricePerCall:     item.PricePerCall,
			RateMultiplier:   item.RateMultiplier,
			InputUnitPrice:   decimalMul(item.InputPricePer1K, item.RateMultiplier),
			OutputUnitPrice:  decimalMul(item.OutputPricePer1K, item.RateMultiplier),
			CallUnitPrice:    decimalMul(item.PricePerCall, item.RateMultiplier),
			Status:           item.Status,
		})
	}
	return out, nil
}

func (s UserPortalService) Dashboard(ctx context.Context, userID string) (UserDashboard, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return UserDashboard{}, err
	}
	models, err := s.Models(ctx)
	if err != nil {
		return UserDashboard{}, err
	}
	stats, err := s.reports.UserDashboard(ctx, userID)
	if err != nil {
		return UserDashboard{}, err
	}
	frozen, _ := s.frozen.Get(ctx, userID)
	return UserDashboard{
		Balance:         user.Balance,
		TodayCharge:     stats.TodayCharge,
		MonthCharge:     stats.MonthCharge,
		Frozen:          unitsToDecimalString(frozen),
		AvailableModels: len(models),
		TodayRequests:   stats.TodayRequests,
	}, nil
}

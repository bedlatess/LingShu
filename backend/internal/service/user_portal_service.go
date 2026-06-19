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
	InputPricePer1K  string `json:"-"`
	OutputPricePer1K string `json:"-"`
	PricePerCall     string `json:"-"`
	RateMultiplier   string `json:"-"`
	SupportsStream   bool   `json:"supports_stream"`
	SupportsTools    bool   `json:"supports_tools"`
	SupportsVision   bool   `json:"supports_vision"`
	InputUnitPrice   string `json:"input_unit_price"`
	OutputUnitPrice  string `json:"output_unit_price"`
	CallUnitPrice    string `json:"call_unit_price"`
	Status           string `json:"status"`
	SortOrder        int    `json:"sort_order"`
}

type UserDashboard struct {
	Balance         string `json:"balance"`
	TodayCharge     string `json:"today_charge"`
	MonthCharge     string `json:"month_charge"`
	TotalCharge     string `json:"total_charge"`
	TotalRecharge   string `json:"total_recharge"`
	Frozen          string `json:"frozen"`
	AvailableModels int    `json:"available_models"`
	TodayRequests   int    `json:"today_requests"`
}

func NewUserPortalService(users repository.UserRepository, models repository.ModelRepository, reports repository.ReportRepository, frozen redisstore.FrozenStore) UserPortalService {
	return UserPortalService{users: users, models: models, reports: reports, frozen: frozen}
}

func (s UserPortalService) Models(ctx context.Context) ([]UserModelPrice, error) {
	items, err := s.models.ListVisible(ctx)
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
			SupportsStream:   item.SupportsStream,
			SupportsTools:    item.SupportsTools,
			SupportsVision:   item.SupportsVision,
			InputUnitPrice:   decimalMul(item.InputPricePer1K, item.RateMultiplier),
			OutputUnitPrice:  decimalMul(item.OutputPricePer1K, item.RateMultiplier),
			CallUnitPrice:    decimalMul(item.PricePerCall, item.RateMultiplier),
			Status:           item.Status,
			SortOrder:        item.SortOrder,
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
		TotalCharge:     stats.TotalCharge,
		TotalRecharge:   stats.TotalRecharge,
		Frozen:          unitsToDecimalString(frozen),
		AvailableModels: len(models),
		TodayRequests:   stats.TodayRequests,
	}, nil
}

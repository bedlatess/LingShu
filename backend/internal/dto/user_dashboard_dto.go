package dto

import "lingshu/backend/internal/service"

type UserDashboardDTO struct {
	Balance         string `json:"balance"`
	TodayCharge     string `json:"today_charge"`
	MonthCharge     string `json:"month_charge"`
	TotalCharge     string `json:"total_charge"`
	TotalRecharge   string `json:"total_recharge"`
	Frozen          string `json:"frozen"`
	AvailableModels int    `json:"available_models"`
	TodayRequests   int    `json:"today_requests"`
}

func NewUserDashboardDTO(item service.UserDashboard) UserDashboardDTO {
	return UserDashboardDTO{
		Balance:         item.Balance,
		TodayCharge:     item.TodayCharge,
		MonthCharge:     item.MonthCharge,
		TotalCharge:     item.TotalCharge,
		TotalRecharge:   item.TotalRecharge,
		Frozen:          item.Frozen,
		AvailableModels: item.AvailableModels,
		TodayRequests:   item.TodayRequests,
	}
}

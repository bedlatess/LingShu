package dto

import "lingshu/backend/internal/repository"

type UserDailyStatDTO struct {
	Day         string `json:"day"`
	Requests    int    `json:"requests"`
	Successes   int    `json:"successes"`
	Failures    int    `json:"failures"`
	TotalTokens int    `json:"total_tokens"`
	Charge      string `json:"charge"`
}

type UserModelStatDTO struct {
	ModelID     string `json:"model_id"`
	Requests    int    `json:"requests"`
	TotalTokens int    `json:"total_tokens"`
	Charge      string `json:"charge"`
}

func NewUserDailyStatDTO(item repository.DailyStat) UserDailyStatDTO {
	return UserDailyStatDTO{
		Day:         item.Day,
		Requests:    item.Requests,
		Successes:   item.Successes,
		Failures:    item.Failures,
		TotalTokens: item.TotalTokens,
		Charge:      item.Charge,
	}
}

func NewUserDailyStatDTOs(items []repository.DailyStat) []UserDailyStatDTO {
	out := make([]UserDailyStatDTO, 0, len(items))
	for _, item := range items {
		out = append(out, NewUserDailyStatDTO(item))
	}
	return out
}

func NewUserModelStatDTO(item repository.ModelStat) UserModelStatDTO {
	return UserModelStatDTO{
		ModelID:     item.ModelID,
		Requests:    item.Requests,
		TotalTokens: item.TotalTokens,
		Charge:      item.Charge,
	}
}

func NewUserModelStatDTOs(items []repository.ModelStat) []UserModelStatDTO {
	out := make([]UserModelStatDTO, 0, len(items))
	for _, item := range items {
		out = append(out, NewUserModelStatDTO(item))
	}
	return out
}

package dto

import "lingshu/backend/internal/service"

const timeFormat = "2006-01-02T15:04:05Z07:00"

type UserModelConfigDTO struct {
	ID             string `json:"id"`
	PublicName     string `json:"public_name"`
	Type           string `json:"type"`
	Group          string `json:"group"`
	BillingMode    string `json:"billing_mode"`
	SupportsStream bool   `json:"supports_stream"`
	SupportsTools  bool   `json:"supports_tools"`
	SupportsVision bool   `json:"supports_vision"`
	Status         string `json:"status"`
	SortOrder      int    `json:"sort_order"`
}

func NewUserModelConfigDTO(item service.UserModelPrice) UserModelConfigDTO {
	return UserModelConfigDTO{
		ID:             item.ID,
		PublicName:     item.PublicName,
		Type:           item.Type,
		Group:          item.Group,
		BillingMode:    item.BillingMode,
		SupportsStream: item.SupportsStream,
		SupportsTools:  item.SupportsTools,
		SupportsVision: item.SupportsVision,
		Status:         item.Status,
		SortOrder:      item.SortOrder,
	}
}

func NewUserModelConfigDTOs(items []service.UserModelPrice) []UserModelConfigDTO {
	out := make([]UserModelConfigDTO, 0, len(items))
	for _, item := range items {
		out = append(out, NewUserModelConfigDTO(item))
	}
	return out
}

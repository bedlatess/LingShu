package dto

import "lingshu/backend/internal/repository"

type UserGatewayLogDTO struct {
	RequestID   string `json:"request_id"`
	ModelID     string `json:"model_id"`
	Status      string `json:"status"`
	HTTPStatus  int    `json:"http_status"`
	TotalTokens int    `json:"total_tokens"`
	Charge      string `json:"charge"`
	CreatedAt   string `json:"created_at"`
}

func NewUserGatewayLogDTO(item repository.GatewayLog) UserGatewayLogDTO {
	return UserGatewayLogDTO{
		RequestID:   item.RequestID,
		ModelID:     item.ModelID,
		Status:      item.Status,
		HTTPStatus:  item.HTTPStatus,
		TotalTokens: item.TotalTokens,
		Charge:      item.Charge,
		CreatedAt:   item.CreatedAt.Format(timeFormat),
	}
}

func NewUserGatewayLogDTOs(items []repository.GatewayLog) []UserGatewayLogDTO {
	out := make([]UserGatewayLogDTO, 0, len(items))
	for _, item := range items {
		out = append(out, NewUserGatewayLogDTO(item))
	}
	return out
}

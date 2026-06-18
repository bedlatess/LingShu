package dto

import "lingshu/backend/internal/repository"

type UserLedgerRecordDTO struct {
	Type          string `json:"type"`
	Amount        string `json:"amount"`
	BalanceBefore string `json:"balance_before"`
	BalanceAfter  string `json:"balance_after"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"created_at"`
}

func NewUserLedgerRecordDTO(item repository.LedgerRecord) UserLedgerRecordDTO {
	return UserLedgerRecordDTO{
		Type:          item.Type,
		Amount:        item.Amount,
		BalanceBefore: item.BalanceBefore,
		BalanceAfter:  item.BalanceAfter,
		Remark:        item.Remark,
		CreatedAt:     item.CreatedAt.Format(timeFormat),
	}
}

func NewUserLedgerRecordDTOs(items []repository.LedgerRecord) []UserLedgerRecordDTO {
	out := make([]UserLedgerRecordDTO, 0, len(items))
	for _, item := range items {
		out = append(out, NewUserLedgerRecordDTO(item))
	}
	return out
}

package service

import (
	"context"
	"errors"
	"strings"

	"lingshu/backend/internal/pkg/password"
	"lingshu/backend/internal/repository"
)

type AdminUserService struct {
	users  repository.UserRepository
	audits repository.AuditRepository
	keys   repository.APIKeyRepository
}

type CreateUserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserInput struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Status   *string `json:"status"`
}

type AdjustBalanceInput struct {
	Amount string `json:"amount"`
	Remark string `json:"remark"`
}

func NewAdminUserService(users repository.UserRepository, audits repository.AuditRepository, keys ...repository.APIKeyRepository) AdminUserService {
	service := AdminUserService{users: users, audits: audits}
	if len(keys) > 0 {
		service.keys = keys[0]
	}
	return service
}

func (s AdminUserService) List(ctx context.Context) ([]repository.User, error) {
	return s.users.List(ctx)
}

func (s AdminUserService) ListPaged(ctx context.Context, page, limit int) ([]repository.User, int, error) {
	return s.users.ListPaged(ctx, limit, (page-1)*limit)
}

func (s AdminUserService) Get(ctx context.Context, id string) (repository.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s AdminUserService) Create(ctx context.Context, actorID string, input CreateUserInput, ip, userAgent string) (repository.User, error) {
	if strings.TrimSpace(input.Username) == "" || len(input.Password) < 8 {
		return repository.User{}, errors.New("username and password with at least 8 characters are required")
	}
	role := input.Role
	if role == "" {
		role = "user"
	}
	if role != "admin" && role != "user" {
		return repository.User{}, errors.New("invalid role")
	}
	hash, err := password.Hash(input.Password)
	if err != nil {
		return repository.User{}, err
	}
	user, err := s.users.Create(ctx, repository.CreateUserParams{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hash,
		Role:         role,
		Status:       "active",
	})
	if err != nil {
		return repository.User{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.user.create",
		TargetType: "user",
		TargetID:   user.ID,
		After:      user,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return user, nil
}

func (s AdminUserService) Update(ctx context.Context, actorID, id string, input UpdateUserInput, ip, userAgent string) (repository.User, error) {
	before, err := s.users.FindByID(ctx, id)
	if err != nil {
		return repository.User{}, err
	}
	if input.Status != nil && *input.Status != "active" && *input.Status != "banned" {
		return repository.User{}, errors.New("invalid status")
	}
	after, err := s.users.Update(ctx, repository.UpdateUserParams{
		ID:       id,
		Username: input.Username,
		Email:    input.Email,
		Status:   input.Status,
	})
	if err != nil {
		return repository.User{}, err
	}
	if input.Status != nil && *input.Status == "banned" && s.keys.HasStore() {
		if err := s.keys.DisableByUser(ctx, id); err != nil {
			return repository.User{}, err
		}
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.user.update",
		TargetType: "user",
		TargetID:   id,
		Before:     before,
		After:      after,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return after, nil
}

func (s AdminUserService) ResetPassword(ctx context.Context, actorID, id, newPassword, ip, userAgent string) error {
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	hash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.SetPassword(ctx, id, hash); err != nil {
		return err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.user.reset_password",
		TargetType: "user",
		TargetID:   id,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return nil
}

func (s AdminUserService) Ban(ctx context.Context, actorID, id, ip, userAgent string) (repository.User, error) {
	status := "banned"
	return s.Update(ctx, actorID, id, UpdateUserInput{Status: &status}, ip, userAgent)
}

func (s AdminUserService) AdjustBalance(ctx context.Context, actorID, id string, input AdjustBalanceInput, ip, userAgent string) (repository.User, error) {
	amount := strings.TrimSpace(input.Amount)
	if amount == "" || strings.TrimSpace(input.Remark) == "" {
		return repository.User{}, errors.New("amount and remark are required")
	}
	ledgerType := "admin_grant"
	if strings.HasPrefix(amount, "-") {
		ledgerType = "admin_deduct"
	}
	before, _ := s.users.FindByID(ctx, id)
	after, err := s.users.AdjustBalance(ctx, id, actorID, amount, ledgerType, input.Remark)
	if err != nil {
		return repository.User{}, err
	}
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.user.adjust_balance",
		TargetType: "user",
		TargetID:   id,
		Before:     before,
		After: map[string]any{
			"user":        after,
			"amount":      amount,
			"ledger_type": ledgerType,
			"remark":      input.Remark,
		},
		IP:        ip,
		UserAgent: userAgent,
	})
	return after, nil
}

func (s AdminUserService) AuditCount(ctx context.Context) (int64, error) {
	return s.audits.Count(ctx)
}

package service

import (
	"context"
	"errors"
	"time"

	"lingshu/backend/internal/config"
	"lingshu/backend/internal/pkg/password"
	"lingshu/backend/internal/pkg/token"
	"lingshu/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrForbidden          = errors.New("forbidden")
)

type AuthService struct {
	cfg   config.Config
	users repository.UserRepository
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token string          `json:"token"`
	User  repository.User `json:"user"`
}

func NewAuthService(cfg config.Config, users repository.UserRepository) AuthService {
	return AuthService{cfg: cfg, users: users}
}

func (s AuthService) Login(ctx context.Context, login, plainPassword string) (LoginResult, error) {
	user, err := s.users.FindByUsernameOrEmail(ctx, login)
	if err != nil || !password.Verify(user.PasswordHash, plainPassword) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if user.Status != "active" {
		return LoginResult{}, ErrUserDisabled
	}
	signed, err := token.Sign(s.cfg.JWTSecret, user.ID, user.Role, 24*time.Hour)
	if err != nil {
		return LoginResult{}, err
	}
	_ = s.users.TouchLastLogin(ctx, user.ID)
	return LoginResult{Token: signed, User: user}, nil
}

func (s AuthService) Register(ctx context.Context, input RegisterInput) (repository.User, error) {
	if !s.cfg.RegistrationEnabled {
		return repository.User{}, ErrForbidden
	}
	if input.Username == "" || len(input.Password) < 8 {
		return repository.User{}, ErrInvalidCredentials
	}
	hash, err := password.Hash(input.Password)
	if err != nil {
		return repository.User{}, err
	}
	return s.users.Create(ctx, repository.CreateUserParams{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hash,
		Role:         "user",
		Status:       "active",
	})
}

func (s AuthService) Me(ctx context.Context, userID string) (repository.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !password.Verify(user.PasswordHash, oldPassword) {
		return ErrInvalidCredentials
	}
	hash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.users.SetPassword(ctx, userID, hash)
}

package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/config"
	"lingshu/backend/internal/pkg/password"
	"lingshu/backend/internal/pkg/token"
	"lingshu/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrForbidden          = errors.New("forbidden")
	ErrCaptchaRequired    = errors.New("captcha token required")
)

type AuthService struct {
	cfg      config.Config
	users    repository.UserRepository
	settings repository.SettingsRepository
	email    EmailService
	redis    *redis.Client
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
	Captcha  string `json:"captcha_token"`
}

type LoginResult struct {
	Token string          `json:"token"`
	User  repository.User `json:"user"`
}

func NewAuthService(cfg config.Config, users repository.UserRepository, settings repository.SettingsRepository, email EmailService, redisClient *redis.Client) AuthService {
	return AuthService{cfg: cfg, users: users, settings: settings, email: email, redis: redisClient}
}

func (s AuthService) Login(ctx context.Context, login, plainPassword, ip, captcha string) (LoginResult, error) {
	if err := s.checkLoginLock(ctx, login, ip); err != nil {
		return LoginResult{}, err
	}
	if err := s.requireCaptchaIfEnabled(ctx, captcha); err != nil {
		return LoginResult{}, err
	}
	user, err := s.users.FindByUsernameOrEmail(ctx, login)
	if err != nil || !password.Verify(user.PasswordHash, plainPassword) {
		_ = s.recordLoginFailure(ctx, login, ip)
		return LoginResult{}, ErrInvalidCredentials
	}
	if user.Status != "active" {
		_ = s.recordLoginFailure(ctx, login, ip)
		return LoginResult{}, ErrUserDisabled
	}
	signed, err := token.Sign(s.cfg.JWTSecret, user.ID, user.Role, 24*time.Hour)
	if err != nil {
		return LoginResult{}, err
	}
	_ = s.users.TouchLastLogin(ctx, user.ID)
	_ = s.clearLoginFailures(ctx, login, ip)
	return LoginResult{Token: signed, User: user}, nil
}

func (s AuthService) Register(ctx context.Context, input RegisterInput) (repository.User, error) {
	mode, err := s.registrationMode(ctx)
	if err != nil {
		return repository.User{}, err
	}
	if mode == "closed" || mode == "invite" {
		return repository.User{}, ErrForbidden
	}
	if input.Username == "" || input.Email == "" || len(input.Password) < 8 {
		return repository.User{}, ErrInvalidCredentials
	}
	if err := s.requireCaptchaIfEnabled(ctx, input.Captcha); err != nil {
		return repository.User{}, err
	}
	if err := s.email.VerifyCode(ctx, "register", input.Email, input.Code); err != nil {
		return repository.User{}, err
	}
	hash, err := password.Hash(input.Password)
	if err != nil {
		return repository.User{}, err
	}
	return s.users.Create(ctx, repository.CreateUserParams{
		Username:      input.Username,
		Email:         input.Email,
		PasswordHash:  hash,
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
	})
}

var ErrLoginLocked = errors.New("login temporarily locked")

const (
	loginFailureLimit = 5
	loginLockWindow   = 15 * time.Minute
	loginFailTTL      = 15 * time.Minute
)

func (s AuthService) checkLoginLock(ctx context.Context, login, ip string) error {
	if s.redis == nil {
		return nil
	}
	for _, key := range loginLockKeys(login, ip) {
		count, err := s.redis.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrLoginLocked
		}
	}
	return nil
}

func (s AuthService) recordLoginFailure(ctx context.Context, login, ip string) error {
	if s.redis == nil {
		return nil
	}
	for _, subject := range loginFailSubjects(login, ip) {
		key := "login_fail:" + subject
		count, err := s.redis.Incr(ctx, key).Result()
		if err != nil {
			return err
		}
		if count == 1 {
			_ = s.redis.Expire(ctx, key, loginFailTTL).Err()
		}
		if count >= loginFailureLimit {
			_ = s.redis.Set(ctx, "login_lock:"+subject, "1", loginLockWindow).Err()
		}
	}
	return nil
}

func (s AuthService) clearLoginFailures(ctx context.Context, login, ip string) error {
	if s.redis == nil {
		return nil
	}
	keys := []string{}
	for _, subject := range loginFailSubjects(login, ip) {
		keys = append(keys, "login_fail:"+subject)
	}
	if len(keys) == 0 {
		return nil
	}
	return s.redis.Del(ctx, keys...).Err()
}

func loginLockKeys(login, ip string) []string {
	subjects := loginFailSubjects(login, ip)
	keys := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		keys = append(keys, "login_lock:"+subject)
	}
	return keys
}

func loginFailSubjects(login, ip string) []string {
	subjects := []string{}
	if trimmed := strings.TrimSpace(strings.ToLower(login)); trimmed != "" {
		subjects = append(subjects, "account:"+trimmed)
	}
	if trimmed := strings.TrimSpace(ip); trimmed != "" {
		subjects = append(subjects, "ip:"+trimmed)
	}
	return subjects
}

func (s AuthService) SendEmailCode(ctx context.Context, purpose, email, captcha string) error {
	if purpose == "register" {
		mode, err := s.registrationMode(ctx)
		if err != nil {
			return err
		}
		if mode == "closed" || mode == "invite" {
			return ErrForbidden
		}
		if err := s.requireCaptchaIfEnabled(ctx, captcha); err != nil {
			return err
		}
	}
	return s.email.SendCode(ctx, purpose, email)
}

func (s AuthService) ForgotPassword(ctx context.Context, email, captcha string) error {
	if err := s.requireCaptchaIfEnabled(ctx, captcha); err != nil {
		return err
	}
	return s.email.SendCode(ctx, "reset", email)
}

func (s AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidCredentials
	}
	if err := s.email.VerifyCode(ctx, "reset", email, code); err != nil {
		return err
	}
	user, err := s.users.FindByUsernameOrEmail(ctx, email)
	if err != nil {
		return ErrInvalidCredentials
	}
	hash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.users.SetPassword(ctx, user.ID, hash)
}

func (s AuthService) requireCaptchaIfEnabled(ctx context.Context, captcha string) error {
	settings, err := s.settings.GetMap(ctx, "captcha_enabled")
	if err != nil {
		return err
	}
	if settings["captcha_enabled"] == "true" && strings.TrimSpace(captcha) == "" {
		return ErrCaptchaRequired
	}
	return nil
}

func (s AuthService) registrationMode(ctx context.Context) (string, error) {
	settings, err := s.settings.GetMap(ctx, "registration_mode", "registration_enabled")
	if err != nil {
		return "", err
	}
	mode := settings["registration_mode"]
	if mode == "" {
		if s.cfg.RegistrationEnabled || settings["registration_enabled"] == "true" {
			return "open", nil
		}
		return "closed", nil
	}
	if mode != "open" && mode != "invite" && mode != "closed" {
		return "closed", nil
	}
	return mode, nil
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

package service

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/repository"
)

var (
	ErrEmailNotConfigured = errors.New("smtp is not configured")
	ErrEmailCodeCooldown  = errors.New("email code was sent recently")
	ErrInvalidEmailCode   = errors.New("invalid email code")
)

type EmailService struct {
	settings repository.SettingsRepository
	redis    *redis.Client
}

func NewEmailService(settings repository.SettingsRepository, redisClient *redis.Client) EmailService {
	return EmailService{settings: settings, redis: redisClient}
}

func (s EmailService) SendCode(ctx context.Context, purpose, email string) error {
	purpose = normalizeEmailPurpose(purpose)
	email = strings.TrimSpace(strings.ToLower(email))
	if purpose == "" || email == "" {
		return errors.New("purpose and email are required")
	}
	cooldownKey := "email_code_cooldown:" + purpose + ":" + email
	if s.redis != nil {
		ok, err := s.redis.SetNX(ctx, cooldownKey, "1", time.Minute).Result()
		if err != nil {
			return err
		}
		if !ok {
			return ErrEmailCodeCooldown
		}
	}
	code := randomEmailCode()
	if err := s.send(ctx, email, code, purpose); err != nil {
		if s.redis != nil {
			_ = s.redis.Del(ctx, cooldownKey).Err()
		}
		return err
	}
	if s.redis != nil {
		return s.redis.Set(ctx, emailCodeKey(purpose, email), code, 10*time.Minute).Err()
	}
	return nil
}

func (s EmailService) VerifyCode(ctx context.Context, purpose, email, code string) error {
	purpose = normalizeEmailPurpose(purpose)
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)
	if purpose == "" || email == "" || code == "" {
		return ErrInvalidEmailCode
	}
	if s.redis == nil {
		return ErrInvalidEmailCode
	}
	stored, err := s.redis.Get(ctx, emailCodeKey(purpose, email)).Result()
	if err != nil {
		return ErrInvalidEmailCode
	}
	if stored != code {
		return ErrInvalidEmailCode
	}
	_ = s.redis.Del(ctx, emailCodeKey(purpose, email)).Err()
	return nil
}

func (s EmailService) send(ctx context.Context, to, code, purpose string) error {
	cfg, err := s.smtpConfig(ctx)
	if err != nil {
		return err
	}
	subject := "LingShu verification code"
	if purpose == "reset" {
		subject = "LingShu password reset code"
	}
	body := fmt.Sprintf("Your LingShu verification code is %s. It expires in 10 minutes.\n\nLingShu 验证码：%s，10 分钟内有效。", code, code)
	message := "From: " + cfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	if cfg.TLS {
		return sendMailTLS(addr, cfg.Host, auth, cfg.From, []string{to}, []byte(message))
	}
	return smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(message))
}

type smtpConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
	TLS  bool
}

func (s EmailService) smtpConfig(ctx context.Context) (smtpConfig, error) {
	settings, err := s.settings.GetMap(ctx, "smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from", "smtp_tls")
	if err != nil {
		return smtpConfig{}, err
	}
	port, _ := strconv.Atoi(firstNonEmpty(settings["smtp_port"], "587"))
	cfg := smtpConfig{
		Host: strings.TrimSpace(settings["smtp_host"]),
		Port: port,
		User: strings.TrimSpace(settings["smtp_user"]),
		Pass: settings["smtp_pass"],
		From: strings.TrimSpace(settings["smtp_from"]),
		TLS:  firstNonEmpty(settings["smtp_tls"], "true") == "true",
	}
	if cfg.From == "" {
		cfg.From = cfg.User
	}
	if cfg.Host == "" || cfg.Port <= 0 || cfg.User == "" || cfg.Pass == "" || cfg.From == "" {
		return smtpConfig{}, ErrEmailNotConfigured
	}
	return cfg, nil
}

func sendMailTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	return writer.Close()
}

func normalizeEmailPurpose(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "register", "reset":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return ""
	}
}

func emailCodeKey(purpose, email string) string {
	return "email_code:" + purpose + ":" + email
}

func randomEmailCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

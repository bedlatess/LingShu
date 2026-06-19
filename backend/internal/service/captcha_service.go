package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrCaptchaNotConfigured = errors.New("captcha is not configured")
	ErrInvalidCaptcha       = errors.New("invalid captcha token")
)

type CaptchaVerifier interface {
	Verify(ctx context.Context, token, remoteIP string) error
}

type SettingsGetter interface {
	GetMap(ctx context.Context, keys ...string) (map[string]string, error)
}

type CaptchaService struct {
	settings SettingsGetter
	client   *http.Client
}

func NewCaptchaService(settings SettingsGetter) CaptchaService {
	return CaptchaService{
		settings: settings,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (s CaptchaService) Verify(ctx context.Context, tokenValue, remoteIP string) error {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return ErrCaptchaRequired
	}
	settings, err := s.settings.GetMap(ctx, "captcha_provider", "captcha_secret_key", "captcha_verify_url")
	if err != nil {
		return err
	}
	provider := strings.TrimSpace(strings.ToLower(settings["captcha_provider"]))
	secret := strings.TrimSpace(settings["captcha_secret_key"])
	verifyURL := strings.TrimSpace(settings["captcha_verify_url"])
	if provider == "" || secret == "" {
		return ErrCaptchaNotConfigured
	}
	if verifyURL == "" {
		switch provider {
		case "turnstile", "cloudflare_turnstile":
			verifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
		case "hcaptcha":
			verifyURL = "https://hcaptcha.com/siteverify"
		default:
			return ErrCaptchaNotConfigured
		}
	}
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("response", tokenValue)
	if trimmed := strings.TrimSpace(remoteIP); trimmed != "" {
		values.Set("remoteip", trimmed)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ErrInvalidCaptcha
	}
	if resp.StatusCode >= 400 || !result.Success {
		return ErrInvalidCaptcha
	}
	return nil
}

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestNormalizeAllowedEndpoints(t *testing.T) {
	got, err := NormalizeAllowedEndpoints([]string{"v1/chat/completions", "/v1/messages", "/v1/messages", ""})
	if err != nil {
		t.Fatalf("normalize endpoints: %v", err)
	}
	want := []string{"/v1/chat/completions", "/v1/messages"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestCaptchaVerifierRequiresRemoteSuccess(t *testing.T) {
	ctx := context.Background()
	seenResponse := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		seenResponse = r.Form.Get("response")
		_ = json.NewEncoder(w).Encode(map[string]any{"success": seenResponse == "valid-token"})
	}))
	t.Cleanup(server.Close)

	settings := fakeSettings(map[string]string{
		"captcha_enabled":    "true",
		"captcha_provider":   "turnstile",
		"captcha_secret_key": "secret",
		"captcha_verify_url": server.URL,
	})
	svc := AuthService{settings: settings, captcha: NewCaptchaService(settings)}

	if err := svc.requireCaptchaIfEnabled(ctx, "", "127.0.0.1"); err != ErrCaptchaRequired {
		t.Fatalf("empty captcha err = %v, want %v", err, ErrCaptchaRequired)
	}
	if err := svc.requireCaptchaIfEnabled(ctx, "anything", "127.0.0.1"); err != ErrInvalidCaptcha {
		t.Fatalf("invalid captcha err = %v, want %v", err, ErrInvalidCaptcha)
	}
	if err := svc.requireCaptchaIfEnabled(ctx, "valid-token", "127.0.0.1"); err != nil {
		t.Fatalf("valid captcha err = %v", err)
	}
}

type fakeSettings map[string]string

func (f fakeSettings) GetMap(_ context.Context, keys ...string) (map[string]string, error) {
	out := map[string]string{}
	for _, key := range keys {
		out[key] = f[key]
	}
	return out, nil
}

func TestNormalizeAllowedEndpointsRejectsUnknown(t *testing.T) {
	if _, err := NormalizeAllowedEndpoints([]string{"/v1/images/generations"}); err == nil {
		t.Fatalf("expected unknown endpoint to be rejected")
	}
}

func TestLoginFailureLock(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	svc := AuthService{redis: client}

	for i := 0; i < loginFailureLimit; i++ {
		if err := svc.recordLoginFailure(ctx, "user@example.com", "127.0.0.1"); err != nil {
			t.Fatalf("record failure: %v", err)
		}
	}
	if err := svc.checkLoginLock(ctx, "user@example.com", "127.0.0.1"); err != ErrLoginLocked {
		t.Fatalf("lock err = %v, want %v", err, ErrLoginLocked)
	}
	if ttl := mr.TTL("login_lock:account:user@example.com"); ttl != loginLockWindow {
		t.Fatalf("account lock ttl = %s, want %s", ttl, loginLockWindow)
	}
}

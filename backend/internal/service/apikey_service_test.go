package service

import (
	"context"
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

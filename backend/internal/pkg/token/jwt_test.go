package token

import (
	"testing"
	"time"
)

func TestSignAndParse(t *testing.T) {
	raw, err := Sign("secret", "user-id", "admin", time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	claims, err := Parse("secret", raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-id" || claims.Role != "admin" {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	raw, err := Sign("secret", "user-id", "admin", time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := Parse("wrong", raw); err == nil {
		t.Fatal("expected wrong secret to fail")
	}
}

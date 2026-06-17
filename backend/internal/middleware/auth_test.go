package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lingshu/backend/internal/pkg/token"
)

func TestAdminOnlyAllowsAdmin(t *testing.T) {
	raw, err := token.Sign("secret", "1", "admin", time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	handler := JWTAuth("secret")(AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestAdminOnlyRejectsUser(t *testing.T) {
	raw, err := token.Sign("secret", "1", "user", time.Minute)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	handler := JWTAuth("secret")(AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

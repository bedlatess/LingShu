package middleware

import (
	"context"
	"net/http"
	"strings"

	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/pkg/token"
)

type contextKey string

const authContextKey contextKey = "auth"

type AuthUser struct {
	ID   string
	Role string
}

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpx.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			claims, err := token.Parse(secret, strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), authContextKey, AuthUser{ID: claims.UserID, Role: claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUser(r.Context())
		if !ok || user.Role != "admin" {
			httpx.Error(w, http.StatusForbidden, "admin role required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CurrentUser(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(authContextKey).(AuthUser)
	return user, ok
}

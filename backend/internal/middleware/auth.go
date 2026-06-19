package middleware

import (
	"context"
	"net/http"
	"strings"

	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/pkg/token"
	"lingshu/backend/internal/repository"
)

type contextKey string

const authContextKey contextKey = "auth"

type AuthUser struct {
	ID   string
	Role string
}

type userLookup interface {
	FindByID(ctx context.Context, id string) (repository.User, error)
}

func JWTAuth(secret string, lookups ...userLookup) func(http.Handler) http.Handler {
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
			if len(lookups) > 0 && lookups[0] != nil {
				user, err := lookups[0].FindByID(r.Context(), claims.UserID)
				if err != nil || user.Status != "active" {
					httpx.Error(w, http.StatusUnauthorized, "user disabled")
					return
				}
				if user.TokenRevokedAt != nil && claims.IssuedAt != nil && !claims.IssuedAt.Time.After(*user.TokenRevokedAt) {
					httpx.Error(w, http.StatusUnauthorized, "token revoked")
					return
				}
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

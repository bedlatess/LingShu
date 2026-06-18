package middleware

import (
	"context"
	"net/http"
	"strings"

	"lingshu/backend/internal/pkg/apikey"
	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/repository"
)

const gatewayContextKey contextKey = "gateway_principal"

func APIKeyAuth(keys repository.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimSpace(r.Header.Get("x-api-key"))
			if raw == "" {
				header := r.Header.Get("Authorization")
				if !strings.HasPrefix(header, "Bearer ") {
					httpx.ErrorJSON(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key", "invalid_api_key")
					return
				}
				raw = strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			}
			principal, err := keys.FindPrincipalByHash(r.Context(), apikey.Hash(raw))
			if err != nil || principal.UserStatus != "active" || principal.KeyStatus != "active" {
				httpx.ErrorJSON(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key", "invalid_api_key")
				return
			}
			ctx := context.WithValue(r.Context(), gatewayContextKey, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CurrentGatewayPrincipal(ctx context.Context) (repository.APIKeyPrincipal, bool) {
	principal, ok := ctx.Value(gatewayContextKey).(repository.APIKeyPrincipal)
	return principal, ok
}

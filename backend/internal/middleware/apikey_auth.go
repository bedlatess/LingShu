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
			if !endpointAllowed(r.URL.Path, principal.AllowedEndpoints) {
				httpx.ErrorJSON(w, http.StatusForbidden, "endpoint_not_allowed", "api key is not allowed to access this endpoint", "endpoint_not_allowed")
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

func endpointAllowed(path string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}
	for _, endpoint := range allowed {
		candidate := strings.TrimRight(endpoint, "/")
		if candidate == "" {
			candidate = "/"
		}
		if candidate == path {
			return true
		}
		if equivalentGatewayEndpoint(candidate, path) {
			return true
		}
	}
	return false
}

func equivalentGatewayEndpoint(candidate, path string) bool {
	return (candidate == "/messages" && path == "/v1/messages") || (candidate == "/v1/messages" && path == "/messages")
}

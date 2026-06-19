package middleware

import (
	"context"
	"net/http"

	"lingshu/backend/internal/pkg/httpx"
	"lingshu/backend/internal/service"
)

type accessGuard interface {
	Check(ctx context.Context, subject service.AccessSubject) (service.AccessBlacklistMatch, error)
	DeviceSecret(ctx context.Context) string
}

func AccessBlacklist(guard accessGuard, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if guard == nil {
				next.ServeHTTP(w, r)
				return
			}
			match, err := guard.Check(r.Context(), service.AccessSubject{
				Scope:    scope,
				IP:       httpx.ClientIP(r, httpx.SettingsFromContext(r.Context())),
				DeviceID: httpx.VerifiedDeviceID(r, guard.DeviceSecret(r.Context())),
			})
			if err != nil {
				httpx.ErrorJSON(w, http.StatusServiceUnavailable, "security_check_failed", "security check failed", "security_check_failed")
				return
			}
			if match.Blocked {
				httpx.ErrorJSON(w, http.StatusForbidden, "access_denied", "操作异常，请联系客服", "access_denied")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

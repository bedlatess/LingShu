package httpx

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type trustedProxySettings interface {
	GetMap(ctx context.Context, keys ...string) (map[string]string, error)
}

type settingsContextKey struct{}

func WithSettings(settings trustedProxySettings) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), settingsContextKey{}, settings)))
		})
	}
}

func SettingsFromContext(ctx context.Context) trustedProxySettings {
	settings, _ := ctx.Value(settingsContextKey{}).(trustedProxySettings)
	return settings
}

func ClientIP(r *http.Request, settings trustedProxySettings) string {
	if trustProxy(r.Context(), settings) {
		if ip := proxyHeaderIP(r, trustedProxyHop(r.Context(), settings)); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func trustProxy(ctx context.Context, settings trustedProxySettings) bool {
	if settings == nil {
		return false
	}
	values, err := settings.GetMap(ctx, "trusted_proxy_enabled")
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(values["trusted_proxy_enabled"]), "true")
}

func proxyHeaderIP(r *http.Request, hop int) string {
	if value := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); value != "" {
		if net.ParseIP(value) != nil {
			return value
		}
	}
	if value := strings.TrimSpace(r.Header.Get("X-Real-IP")); value != "" {
		if net.ParseIP(value) != nil {
			return value
		}
	}
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")
	index := len(parts) - hop
	if index < 0 {
		index = 0
	}
	candidate := strings.TrimSpace(parts[index])
	if net.ParseIP(candidate) == nil {
		return ""
	}
	return candidate
}

func trustedProxyHop(ctx context.Context, settings trustedProxySettings) int {
	if settings == nil {
		return 1
	}
	values, err := settings.GetMap(ctx, "trusted_proxy_hops")
	if err != nil {
		return 1
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(values["trusted_proxy_hops"]))
	if err != nil || parsed <= 0 {
		return 1
	}
	return parsed
}

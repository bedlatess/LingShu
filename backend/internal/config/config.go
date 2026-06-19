package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv                        string
	AppPort                       string
	AppPublicURL                  string
	DatabaseURL                   string
	RedisURL                      string
	JWTSecret                     string
	KeyEncryptionSecret           string
	DefaultRateMultiplier         string
	DefaultMaxTokens              int
	APIKeyPrefix                  string
	RegistrationEnabled           bool
	DefaultUserRPMLimit           int
	DefaultUserConcurrencyLimit   int
	DefaultGatewayTimeoutSeconds  int
	AllowedOrigins                []string
	GatewayMaxBodyBytes           int64
	CleanupEnabled                bool
	CleanupLogRetentionDays       int
	CleanupAuditRetentionDays     int
	CleanupAnnouncementGraceDays  int
	CleanupRedeemGraceDays        int
	ChannelHealerEnabled          bool
	ChannelHealerIntervalSeconds  int
	ChannelHealerSuccessThreshold int
	AdminUser                     string
	AdminPass                     string
}

func Load() Config {
	return Config{
		AppEnv:                        env("APP_ENV", "development"),
		AppPort:                       env("APP_PORT", "8080"),
		AppPublicURL:                  env("APP_PUBLIC_URL", "http://localhost:8080"),
		DatabaseURL:                   env("DATABASE_URL", "postgres://lingshu:lingshu@localhost:5432/lingshu?sslmode=disable"),
		RedisURL:                      env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:                     env("JWT_SECRET", "change-me"),
		KeyEncryptionSecret:           env("KEY_ENCRYPTION_SECRET", "change-me-32-bytes"),
		DefaultRateMultiplier:         env("DEFAULT_RATE_MULTIPLIER", "1.2"),
		DefaultMaxTokens:              envInt("DEFAULT_MAX_TOKENS", 4096),
		APIKeyPrefix:                  env("API_KEY_PREFIX", "lsk_live_"),
		RegistrationEnabled:           envBool("REGISTRATION_ENABLED", false),
		DefaultUserRPMLimit:           envInt("DEFAULT_USER_RPM_LIMIT", 60),
		DefaultUserConcurrencyLimit:   envInt("DEFAULT_USER_CONCURRENCY_LIMIT", 5),
		DefaultGatewayTimeoutSeconds:  envInt("DEFAULT_GATEWAY_TIMEOUT_SECONDS", 120),
		AllowedOrigins:                envStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		GatewayMaxBodyBytes:           envInt64("GATEWAY_MAX_BODY_BYTES", 2*1024*1024),
		CleanupEnabled:                envBool("CLEANUP_ENABLED", false),
		CleanupLogRetentionDays:       envInt("CLEANUP_LOG_RETENTION_DAYS", 30),
		CleanupAuditRetentionDays:     envInt("CLEANUP_AUDIT_RETENTION_DAYS", 30),
		CleanupAnnouncementGraceDays:  envInt("CLEANUP_ANNOUNCEMENT_GRACE_DAYS", 30),
		CleanupRedeemGraceDays:        envInt("CLEANUP_REDEEM_GRACE_DAYS", 90),
		ChannelHealerEnabled:          envBool("CHANNEL_HEALER_ENABLED", true),
		ChannelHealerIntervalSeconds:  envInt("CHANNEL_HEALER_INTERVAL_SECONDS", 300),
		ChannelHealerSuccessThreshold: envInt("CHANNEL_HEALER_SUCCESS_THRESHOLD", 3),
		AdminUser:                     env("ADMIN_USER", "admin"),
		AdminPass:                     env("ADMIN_PASS", "change-me"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envStringSlice(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

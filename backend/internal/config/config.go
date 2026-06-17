package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv                       string
	AppPort                      string
	AppPublicURL                 string
	DatabaseURL                  string
	RedisURL                     string
	JWTSecret                    string
	KeyEncryptionSecret          string
	DefaultRateMultiplier        string
	APIKeyPrefix                 string
	RegistrationEnabled          bool
	DefaultUserRPMLimit          int
	DefaultUserConcurrencyLimit  int
	DefaultGatewayTimeoutSeconds int
	AdminUser                    string
	AdminPass                    string
}

func Load() Config {
	return Config{
		AppEnv:                       env("APP_ENV", "development"),
		AppPort:                      env("APP_PORT", "8080"),
		AppPublicURL:                 env("APP_PUBLIC_URL", "http://localhost:8080"),
		DatabaseURL:                  env("DATABASE_URL", "postgres://lingshu:lingshu@localhost:5432/lingshu?sslmode=disable"),
		RedisURL:                     env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:                    env("JWT_SECRET", "change-me"),
		KeyEncryptionSecret:          env("KEY_ENCRYPTION_SECRET", "change-me-32-bytes"),
		DefaultRateMultiplier:        env("DEFAULT_RATE_MULTIPLIER", "1.2"),
		APIKeyPrefix:                 env("API_KEY_PREFIX", "lsk_live_"),
		RegistrationEnabled:          envBool("REGISTRATION_ENABLED", false),
		DefaultUserRPMLimit:          envInt("DEFAULT_USER_RPM_LIMIT", 60),
		DefaultUserConcurrencyLimit:  envInt("DEFAULT_USER_CONCURRENCY_LIMIT", 5),
		DefaultGatewayTimeoutSeconds: envInt("DEFAULT_GATEWAY_TIMEOUT_SECONDS", 120),
		AdminUser:                    env("ADMIN_USER", "admin"),
		AdminPass:                    env("ADMIN_PASS", "change-me"),
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

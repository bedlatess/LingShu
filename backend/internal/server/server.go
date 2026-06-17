package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/config"
	adminhandler "lingshu/backend/internal/handler/admin"
	authhandler "lingshu/backend/internal/handler/auth"
	gatewayhandler "lingshu/backend/internal/handler/gateway"
	userhandler "lingshu/backend/internal/handler/user"
	"lingshu/backend/internal/middleware"
	redisstore "lingshu/backend/internal/redis"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type Server struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func New(cfg config.Config, db *pgxpool.Pool, redisClient *redis.Client) http.Handler {
	s := &Server{DB: db, Redis: redisClient}
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	modelRepo := repository.NewModelRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	gatewayRepo := repository.NewGatewayRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	redeemRepo := repository.NewRedeemRepository(db)
	reportRepo := repository.NewReportRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	frozenStore := redisstore.NewFrozenStore(redisClient)
	authService := service.NewAuthService(cfg, userRepo)
	adminUserService := service.NewAdminUserService(userRepo, auditRepo)
	apiKeyService := service.NewAPIKeyService(cfg, apiKeyRepo, auditRepo)
	modelService := service.NewModelService(modelRepo, auditRepo)
	channelService := service.NewChannelService(channelRepo, auditRepo)
	gatewayService := service.NewGatewayService(gatewayRepo, frozenStore)
	announcementService := service.NewAnnouncementService(announcementRepo, auditRepo)
	redeemService := service.NewRedeemService(redeemRepo, auditRepo)
	reportService := service.NewReportService(reportRepo)
	userPortalService := service.NewUserPortalService(userRepo, modelRepo, reportRepo, frozenStore)
	settingsService := service.NewSettingsService(settingsRepo, auditRepo)
	authHandler := authhandler.New(authService)
	adminUsers := adminhandler.NewUserHandler(adminUserService)
	adminKeys := adminhandler.NewAPIKeyHandler(apiKeyService)
	adminModels := adminhandler.NewModelHandler(modelService)
	adminChannels := adminhandler.NewChannelHandler(channelService)
	adminAnnouncements := adminhandler.NewAnnouncementHandler(announcementService)
	adminRedeems := adminhandler.NewRedeemHandler(redeemService)
	userHandler := userhandler.New(announcementService, redeemService, apiKeyService, userPortalService)
	adminReports := adminhandler.NewReportHandler(reportService)
	adminSettings := adminhandler.NewSettingsHandler(settingsService, auditRepo)
	userReports := userhandler.NewReportHandler(reportService)
	gatewayHandler := gatewayhandler.New(gatewayService)

	r := chi.NewRouter()
	r.Use(cors)
	r.Get("/healthz", s.healthz)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/register", authHandler.Register)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))
			r.Get("/me", authHandler.Me)
			r.Post("/change-password", authHandler.ChangePassword)
		})
	})
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret))
		r.Use(middleware.AdminOnly)
		r.Get("/audit-count", adminUsers.AuditCount)
		r.Get("/audit-logs", adminSettings.AuditLogs)
		r.Get("/dashboard", adminReports.Dashboard)
		r.Get("/settings", adminSettings.List)
		r.Patch("/settings", adminSettings.Patch)
		r.Get("/users", adminUsers.List)
		r.Post("/users", adminUsers.Create)
		r.Patch("/users/{id}", adminUsers.Update)
		r.Post("/users/{id}/reset-password", adminUsers.ResetPassword)
		r.Post("/users/{id}/ban", adminUsers.Ban)
		r.Post("/users/{id}/balance", adminUsers.AdjustBalance)
		r.Get("/api-keys", adminKeys.List)
		r.Post("/api-keys", adminKeys.Create)
		r.Patch("/api-keys/{id}", adminKeys.Disable)
		r.Get("/models", adminModels.List)
		r.Post("/models", adminModels.Create)
		r.Patch("/models/{id}", adminModels.Update)
		r.Post("/models/{id}/disable", adminModels.Disable)
		r.Get("/channels", adminChannels.List)
		r.Post("/channels", adminChannels.Create)
		r.Patch("/channels/{id}", adminChannels.Update)
		r.Post("/channels/{id}/test", adminChannels.Test)
		r.Post("/channels/{id}/disable", adminChannels.Disable)
		r.Post("/channel-models", adminChannels.BindModel)
		r.Get("/announcements", adminAnnouncements.List)
		r.Post("/announcements", adminAnnouncements.Create)
		r.Patch("/announcements/{id}", adminAnnouncements.Update)
		r.Delete("/announcements/{id}", adminAnnouncements.Delete)
		r.Get("/redeem-codes", adminRedeems.List)
		r.Post("/redeem-codes", adminRedeems.Create)
		r.Post("/redeem-codes/batch", adminRedeems.Create)
		r.Post("/redeem-codes/{id}/disable", adminRedeems.Disable)
		r.Get("/gateway-requests", adminReports.Logs)
		r.Get("/balance-ledger", adminReports.Ledger)
	})
	r.Route("/api/user", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret))
		r.Get("/dashboard", userHandler.Dashboard)
		r.Get("/models", userHandler.Models)
		r.Get("/api-keys", userHandler.APIKeys)
		r.Post("/api-keys", userHandler.CreateAPIKey)
		r.Patch("/api-keys/{id}", userHandler.UpdateAPIKey)
		r.Delete("/api-keys/{id}", userHandler.DeleteAPIKey)
		r.Get("/announcements", userHandler.Announcements)
		r.Post("/redeem", userHandler.Redeem)
		r.Get("/usage/logs", userReports.Logs)
		r.Get("/usage/ledger", userReports.Ledger)
		r.Get("/usage/stats/daily", userReports.Daily)
		r.Get("/usage/stats/models", userReports.Models)
	})
	r.Route("/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(apiKeyRepo))
		r.Get("/models", gatewayHandler.Models)
		r.Post("/chat/completions", gatewayHandler.ChatCompletions)
	})
	return r
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := s.DB.Ping(ctx); err != nil {
		dbStatus = "error"
	}

	redisStatus := "ok"
	if err := s.Redis.Ping(ctx).Err(); err != nil {
		redisStatus = "error"
	}

	status := http.StatusOK
	if dbStatus != "ok" || redisStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":   http.StatusText(status),
		"database": dbStatus,
		"redis":    redisStatus,
	})
}

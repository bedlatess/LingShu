package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/config"
	adminhandler "lingshu/backend/internal/handler/admin"
	authhandler "lingshu/backend/internal/handler/auth"
	gatewayhandler "lingshu/backend/internal/handler/gateway"
	publichandler "lingshu/backend/internal/handler/public"
	userhandler "lingshu/backend/internal/handler/user"
	"lingshu/backend/internal/job"
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
	apiKeyRepo := repository.NewAPIKeyRepositoryWithCache(db, redisClient)
	modelRepo := repository.NewModelRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	gatewayRepo := repository.NewGatewayRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	redeemRepo := repository.NewRedeemRepository(db)
	reportRepo := repository.NewReportRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	frozenStore := redisstore.NewFrozenStore(redisClient)
	emailService := service.NewEmailService(settingsRepo, redisClient)
	authService := service.NewAuthService(cfg, userRepo, settingsRepo, emailService, redisClient)
	adminUserService := service.NewAdminUserService(userRepo, auditRepo, apiKeyRepo)
	apiKeyService := service.NewAPIKeyService(cfg, apiKeyRepo, auditRepo)
	modelService := service.NewModelService(modelRepo, auditRepo)
	channelService := service.NewChannelService(channelRepo, auditRepo)
	gatewayService := service.NewGatewayService(gatewayRepo, frozenStore, cfg.DefaultMaxTokens)
	announcementService := service.NewAnnouncementService(announcementRepo, auditRepo)
	redeemService := service.NewRedeemService(redeemRepo, auditRepo)
	reportService := service.NewReportService(reportRepo)
	opsService := service.NewOpsService(db)
	userPortalService := service.NewUserPortalService(userRepo, modelRepo, reportRepo, frozenStore)
	settingsService := service.NewSettingsService(settingsRepo, auditRepo)
	cleaner := job.NewCleaner(db, redisClient, job.CleanerConfig{
		LogRetentionDays:      cfg.CleanupLogRetentionDays,
		AuditRetentionDays:    cfg.CleanupAuditRetentionDays,
		AnnouncementGraceDays: cfg.CleanupAnnouncementGraceDays,
		RedeemGraceDays:       cfg.CleanupRedeemGraceDays,
	})
	authHandler := authhandler.New(authService)
	adminUsers := adminhandler.NewUserHandler(adminUserService, apiKeyService, reportService)
	adminKeys := adminhandler.NewAPIKeyHandler(apiKeyService)
	adminModels := adminhandler.NewModelHandler(modelService)
	adminChannels := adminhandler.NewChannelHandler(channelService)
	adminAnnouncements := adminhandler.NewAnnouncementHandler(announcementService)
	adminRedeems := adminhandler.NewRedeemHandler(redeemService)
	userHandler := userhandler.New(announcementService, redeemService, apiKeyService, userPortalService)
	adminReports := adminhandler.NewReportHandler(reportService)
	adminOps := adminhandler.NewOpsHandler(opsService)
	adminSettings := adminhandler.NewSettingsHandler(settingsService, auditRepo)
	adminCleanup := adminhandler.NewCleanupHandler(cleaner)
	userReports := userhandler.NewReportHandler(reportService)
	gatewayHandler := gatewayhandler.New(gatewayService)
	publicHandler := publichandler.New(db)
	if cfg.CleanupEnabled {
		job.NewScheduler(cleaner).Start(context.Background())
	}
	if cfg.ChannelHealerEnabled {
		job.NewChannelHealer(channelRepo, channelService, redisClient, cfg.ChannelHealerIntervalSeconds, cfg.ChannelHealerSuccessThreshold).Start(context.Background())
	}

	r := chi.NewRouter()
	r.Use(corsWith(cfg.AllowedOrigins))
	r.Get("/healthz", s.healthz)
	r.Get("/api/public/models", publicHandler.ListModels)
	r.Get("/api/public/site-info", publicHandler.SiteInfo)
	r.Get("/api/public/legal/{slug}", publicHandler.Legal)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/email/send-code", authHandler.SendEmailCode)
		r.Post("/register", authHandler.Register)
		r.Post("/forgot", authHandler.Forgot)
		r.Post("/reset", authHandler.Reset)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret, userRepo))
			r.Get("/me", authHandler.Me)
			r.Post("/change-password", authHandler.ChangePassword)
		})
	})
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret, userRepo))
		r.Use(middleware.AdminOnly)
		r.Get("/audit-count", adminUsers.AuditCount)
		r.Get("/audit-logs", adminSettings.AuditLogs)
		r.Post("/audit-logs/cleanup", adminSettings.CleanupAuditLogs)
		r.Get("/dashboard", adminReports.Dashboard)
		r.Get("/ops", adminOps.Dashboard)
		r.Get("/reports/daily", adminReports.Daily)
		r.Get("/reports/by-user", adminReports.ByUser)
		r.Get("/reports/by-model", adminReports.ByModel)
		r.Get("/reports/by-channel", adminReports.ByChannel)
		r.Get("/settings", adminSettings.List)
		r.Patch("/settings", adminSettings.Patch)
		r.Post("/cleanup/run", adminCleanup.Run)
		r.Get("/cleanup/history", adminCleanup.History)
		r.Get("/users", adminUsers.List)
		r.Post("/users", adminUsers.Create)
		r.Get("/users/{id}", adminUsers.Get)
		r.Patch("/users/{id}", adminUsers.Update)
		r.Post("/users/{id}/reset-password", adminUsers.ResetPassword)
		r.Post("/users/{id}/ban", adminUsers.Ban)
		r.Post("/users/{id}/balance", adminUsers.AdjustBalance)
		r.Get("/users/{id}/logs", adminUsers.UserLogs)
		r.Get("/users/{id}/ledger", adminUsers.UserLedger)
		r.Get("/users/{id}/api-keys", adminUsers.UserAPIKeys)
		r.Get("/users/{id}/summary", adminUsers.UserSummary)
		r.Get("/api-keys", adminKeys.List)
		r.Post("/api-keys", adminKeys.Create)
		r.Patch("/api-keys/{id}", adminKeys.Disable)
		r.Delete("/api-keys/{id}", adminKeys.Delete)
		r.Get("/models", adminModels.List)
		r.Post("/models", adminModels.Create)
		r.Get("/models/{id}", adminModels.Detail)
		r.Patch("/models/{id}", adminModels.Update)
		r.Post("/models/{id}/disable", adminModels.Disable)
		r.Delete("/models/{id}", adminModels.Delete)
		r.Get("/channel-presets", adminChannels.Presets)
		r.Post("/channels/detect", adminChannels.Detect)
		r.Get("/channels", adminChannels.List)
		r.Post("/channels", adminChannels.Create)
		r.Get("/channels/{id}", adminChannels.Detail)
		r.Patch("/channels/{id}", adminChannels.Update)
		r.Post("/channels/{id}/test", adminChannels.Test)
		r.Post("/channels/{id}/sync-models", adminChannels.SyncModels)
		r.Post("/channels/{id}/import-models", adminChannels.ImportModels)
		r.Post("/channels/{id}/disable", adminChannels.Disable)
		r.Delete("/channels/{id}", adminChannels.Delete)
		r.Post("/channel-models", adminChannels.BindModel)
		r.Delete("/channels/{channelID}/models/{modelID}", adminChannels.UnbindModel)
		r.Get("/announcements", adminAnnouncements.List)
		r.Post("/announcements", adminAnnouncements.Create)
		r.Patch("/announcements/{id}", adminAnnouncements.Update)
		r.Delete("/announcements/{id}", adminAnnouncements.Delete)
		r.Get("/redeem-codes", adminRedeems.List)
		r.Post("/redeem-codes", adminRedeems.Create)
		r.Post("/redeem-codes/batch", adminRedeems.Create)
		r.Post("/redeem-codes/{id}/disable", adminRedeems.Disable)
		r.Get("/redeem-codes/{id}/records", adminRedeems.Records)
		r.Get("/gateway-requests", adminReports.Logs)
		r.Get("/balance-ledger", adminReports.Ledger)
	})
	r.Route("/api/user", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret, userRepo))
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
		r.Use(middleware.MaxBody(gatewayMaxBodyBytes(cfg.GatewayMaxBodyBytes)))
		r.Use(middleware.APIKeyAuth(apiKeyRepo))
		r.Get("/models", gatewayHandler.Models)
		r.Post("/chat/completions", gatewayHandler.ChatCompletions)
		r.Post("/messages", gatewayHandler.Messages)
		r.Post("/embeddings", gatewayHandler.Embeddings)
	})
	return r
}

func gatewayMaxBodyBytes(value int64) int64 {
	if value <= 0 {
		return 2 * 1024 * 1024
	}
	return value
}

func corsWith(allowed []string) func(http.Handler) http.Handler {
	allowAll := len(allowed) == 1 && allowed[0] == "*"
	set := make(map[string]bool, len(allowed))
	for _, origin := range allowed {
		set[strings.TrimSpace(origin)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowAll || set[origin]) {
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

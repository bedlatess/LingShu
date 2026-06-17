package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"lingshu/backend/internal/bootstrap"
	"lingshu/backend/internal/config"
	"lingshu/backend/internal/server"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	if err := bootstrap.Migrate(ctx, db, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := bootstrap.SeedAdmin(ctx, db, cfg); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	handler := server.New(cfg, db, redisClient)
	httpServer := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("lingshu backend listening on :%s", cfg.AppPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

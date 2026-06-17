package main

import (
	"context"
	"log"

	"lingshu/backend/internal/bootstrap"
	"lingshu/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
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
	log.Println("migrations applied")
}

package main

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/snet-commerce/merchant/internal/infrastructure/db/postgres"

	"github.com/snet-commerce/merchant/internal/config"
	"github.com/snet-commerce/merchant/internal/infrastructure/logger"
)

func main() {
	logger, err := logger.Development()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	cfg, err := config.Build()
	if err != nil {
		logger.Fatalf("failed to initialize config - %s", err)
	}

	db, err := postgres.Connect(
		cfg.Postgres.PostgresURL,
		postgres.Config{
			MaxOpenConns:    cfg.Postgres.MaxOpenConns,
			MaxIdleConns:    cfg.Postgres.MaxIdleConns,
			ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
			ConnMaxIdleTime: cfg.Postgres.ConnMaxIdleTime,
		},
	)
	if err != nil {
		logger.Fatalf("failed to establish connection to database - %s", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("failed to close database connection - %s", err)
		}
	}()
}

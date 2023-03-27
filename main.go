package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/snet-commerce/merchant/internal/config"
	"github.com/snet-commerce/merchant/internal/handler"
	"github.com/snet-commerce/merchant/internal/infrastructure/db/postgres"
	"github.com/snet-commerce/merchant/internal/infrastructure/logger"
	pb "github.com/snet-commerce/merchant/proto"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logger.ForEnv(cfg.Environment)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	client, err := postgres.Connect(
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
		if err := client.Close(); err != nil {
			logger.Errorf("failed to close database connection - %s", err)
		}
	}()

	merchantHandler := handler.NewMerchantHandler(client, logger)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServerPort))
	if err != nil {
		logrus.Fatal(err)
	}

	srv := grpc.NewServer()
	pb.RegisterMerchantServiceServer(srv, merchantHandler)

	shutdownCh := make(chan os.Signal, 1)
	errorCh := make(chan error, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	go func() {
		logger.Infof("server is listening at port %d", cfg.ServerPort)
		if err := srv.Serve(lis); err != nil {
			logger.Errorf("gRPC server error occurred")
			errorCh <- err
		}
	}()

	select {
	case <-shutdownCh:
		logger.Info("shutdown signal has been sent, stopping the server...")
		srv.Stop()
	case err = <-errorCh:
		logger.Errorf("shutting down the server because of error - %s", err)
	}
}

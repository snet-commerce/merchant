package main

import (
	"fmt"
	"log"
	"net"
	"os"

	grpcpb "buf.build/gen/go/snet-commerce/merchant/grpc/go/merchant/v1/merchantv1grpc"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/snet-commerce/gorch"
	"google.golang.org/grpc"

	"github.com/snet-commerce/merchant/internal/config"
	"github.com/snet-commerce/merchant/internal/handler"
	"github.com/snet-commerce/merchant/internal/infrastructure/db/postgres"
	"github.com/snet-commerce/merchant/internal/infrastructure/logger"
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
		logger.Fatalf("failed to establish connection to database - %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServerPort))
	if err != nil {
		logger.Fatalf("failed to start listener: %v", err)
	}

	merchantHandler := handler.NewMerchantHandler(client.Merchant, logger)

	srv := grpc.NewServer()
	grpcpb.RegisterMerchantServiceServer(srv, merchantHandler)

	orch := gorch.New(gorch.WithStopSignals(os.Interrupt))
	orch.After(func() error {
		return lis.Close()
	}).After(func() error {
		return client.Close()
	})

	err = orch.StartAsync(func() error {
		logger.Infof("starting server on port %d...", cfg.ServerPort)
		if err := srv.Serve(lis); err != nil {
			logger.Errorf("failed to start the server: %s", err)
			return err
		}
		return nil
	})
	if err != nil {
		logger.Fatalf("error occurred on server stratup: %v", err)
	}

	select {
	case err := <-orch.ErrorChannel():
		logger.Errorf("shutting down the server because of unexpected error - %s", err)
	case <-orch.StopChannel():
		logger.Info("shutdown signal has been sent, stopping the server...")
	}
	srv.Stop()
	orch.Wait()
}

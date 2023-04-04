package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	grpcpb "buf.build/gen/go/snet-commerce/merchant/grpc/go/merchant/v1/merchantv1grpc"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/snet-commerce/gorch"
	"github.com/snet-commerce/merchant/internal/infrastructure/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/snet-commerce/merchant/internal/config"
	"github.com/snet-commerce/merchant/internal/handler"
	"github.com/snet-commerce/merchant/internal/infrastructure/db/postgres"
	"github.com/snet-commerce/merchant/internal/infrastructure/logger"
)

const telemetryShutdownTimeout = 3 * time.Second

func main() {
	// configuration
	cfg, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}

	// logger
	logger, err := logger.ForEnv(cfg.Environment)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	// ent client to postgres
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

	// init tracer
	tracer, err := telemetry.ZipkinTracer(
		cfg.Telemetry.ZipkinURL,
		telemetry.WithTracerServiceName(cfg.ServiceName),
		telemetry.WithTracerRatio(cfg.Telemetry.Ratio),
		telemetry.WithTracerLogger(zap.NewStdLog(logger.Desugar())),
	)
	if err != nil {
		logger.Fatalf("failed to setup zipkin tracer - %v", err)
	}
	// set tracer to open telemetry
	otel.SetTracerProvider(tracer)

	// start listener on port from config
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServerPort))
	if err != nil {
		logger.Fatalf("failed to start listener: %v", err)
	}

	// handlers
	merchantHandler := handler.NewMerchantHandler(client.Merchant, logger)

	// init server
	srv := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	grpcpb.RegisterMerchantServiceServer(srv, merchantHandler)

	// start application orchestrator
	orch := gorch.New(gorch.WithStopSignals(os.Interrupt))
	orch.After(func() error {
		logger.Info("shutting down the server...")
		srv.Stop()
		return nil
	}).After(func() error {
		return lis.Close()
	}).After(func() error {
		return client.Close()
	}).After(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), telemetryShutdownTimeout)
		defer cancel()
		return tracer.Shutdown(ctx)
	})

	grpcSrvStarter := func() error {
		logger.Infof("starting server on port %d...", cfg.ServerPort)
		if err := srv.Serve(lis); err != nil {
			logger.Errorf("server error occurred: %v", err)
			return err
		}
		return nil
	}

	for err := range orch.Serve(grpcSrvStarter) {
		logger.Errorf("stopping the server because of unexpected error - %v", err)
		orch.Stop()
	}
}

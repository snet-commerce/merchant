package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	grpcpb "buf.build/gen/go/snet-commerce/merchant/grpc/go/merchant/v1/merchantv1grpc"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	stdLogger := zap.NewStdLog(logger.Desugar())

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
		telemetry.WithTracerLogger(stdLogger),
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

	// init gRPC server
	apiSrv := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	grpcpb.RegisterMerchantServiceServer(apiSrv, merchantHandler)

	// init metrics server
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:     fmt.Sprintf(":%d", cfg.Telemetry.MetricsPort),
		ErrorLog: stdLogger,
		Handler:  handler,
	}

	// start application orchestrator
	orch := gorch.New(gorch.WithStopSignals(os.Interrupt))
	orch.After(func() error {
		logger.Info("shutting down gRPC server...")
		apiSrv.Stop()
		return nil
	}).After(func() error {
		logger.Info("shutting down metrics server...")
		return metricsSrv.Close()
	}).After(func() error {
		return lis.Close()
	}).After(func() error {
		return client.Close()
	}).After(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), telemetryShutdownTimeout)
		defer cancel()
		return tracer.Shutdown(ctx)
	})

	grpcSrvRunner := func() error {
		logger.Infof("starting gRPC server on port %d...", cfg.ServerPort)
		if err := apiSrv.Serve(lis); err != nil {
			logger.Errorf("gRPC server error occurred: %v", err)
			return err
		}
		return nil
	}

	metricsSrvRunner := func() error {
		logger.Infof("starting metrics server on port %d...", cfg.Telemetry.MetricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil {
			logger.Errorf("metrics server error occurred: %v", err)
			return err
		}
		return nil
	}

	for err := range orch.Serve(grpcSrvRunner, metricsSrvRunner) {
		logger.Errorf("stopping the server because of unexpected error - %v", err)
		orch.Stop()
	}
}

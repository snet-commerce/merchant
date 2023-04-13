package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"time"

	grpcpb "buf.build/gen/go/snet-commerce/merchant/grpc/go/merchant/v1/merchantv1grpc"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/snet-commerce/gorch"
	"github.com/snet-commerce/merchant/internal/infrastructure/logger"
	"github.com/snet-commerce/merchant/internal/infrastructure/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"github.com/snet-commerce/merchant/internal/config"
	"github.com/snet-commerce/merchant/internal/handler"
	"github.com/snet-commerce/merchant/internal/infrastructure/db/postgres"
)

const (
	envDebug      = "debug"
	envProduction = "production"
)

const tracerShutdownTimeout = 3 * time.Second

func main() {
	// configuration
	cfg, err := config.Build()
	if err != nil {
		stdlog.Fatal(err)
	}

	// logger
	zapLogger, err := zapLogger(cfg.Environment)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer zapLogger.Sync()

	logger := zapLogger.Sugar()

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
	trc, err := tracer(cfg, logger)
	if err != nil {
		logger.Fatalf("failed to setup tracer - %v", err)
	}
	// set tracer to open telemetry
	otel.SetTracerProvider(trc)

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

	// start application orchestrator
	orch := gorch.New(gorch.WithStopSignals(os.Interrupt))
	orch.After(func() error {
		logger.Info("shutting down gRPC server...")
		apiSrv.Stop()
		return nil
	}).After(func() error {
		return lis.Close()
	}).After(func() error {
		return client.Close()
	}).After(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), tracerShutdownTimeout)
		defer cancel()
		return trc.Shutdown(ctx)
	})

	grpcSrvRunner := func() error {
		logger.Infof("starting gRPC server on port %d...", cfg.ServerPort)
		if err := apiSrv.Serve(lis); err != nil {
			logger.Errorf("gRPC server error occurred: %v", err)
			return err
		}
		return nil
	}

	for err := range orch.Serve(grpcSrvRunner) {
		logger.Errorf("stopping the server because of unexpected error - %v", err)
		orch.Stop()
	}
}

func tracer(cfg *config.Config, logger *zap.SugaredLogger) (*sdktrace.TracerProvider, error) {
	if cfg.Environment == envDebug {
		return telemetry.StdoutTracer()
	}
	return telemetry.ZipkinTracer(
		cfg.Telemetry.ZipkinURL,
		telemetry.WithTracerServiceName(cfg.ServiceName),
		telemetry.WithTracerRatio(cfg.Telemetry.Ratio),
		telemetry.WithTracerLogger(zap.NewStdLog(logger.Desugar())),
	)
}

func zapLogger(env string) (*zap.Logger, error) {
	srvField := zap.Fields(zap.Field{
		Key:    "service",
		Type:   zapcore.StringType,
		String: "merchant service",
	})

	if env == envProduction {
		return logger.Production(srvField)
	}

	return logger.Development(srvField)
}

package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	production = "prod"
	develop    = "dev"
)

func ForEnv(env string) (logger *zap.SugaredLogger, err error) {
	switch env {
	case production:
		logger, err = Production()
	case develop:
		logger, err = Development()
	default:
		err = fmt.Errorf("%s environment is not defined", env)
	}
	return logger, err
}

func Production() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true
	cfg.InitialFields = map[string]any{
		"service": "merchants-service",
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

func Development() (*zap.SugaredLogger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]any{
		"service": "merchant service",
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

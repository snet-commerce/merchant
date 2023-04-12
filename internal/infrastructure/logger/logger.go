package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Development() (*zap.Logger, error) {
	return config().Build()
}

func Production() (*zap.Logger, error) {
	cfg := config()
	cfg.DisableStacktrace = true
	return cfg.Build()
}

func config() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]any{
		"service": "merchant service",
	}
	return cfg
}

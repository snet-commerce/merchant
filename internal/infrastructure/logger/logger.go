package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Development(opts ...zap.Option) (*zap.Logger, error) {
	return config().Build(opts...)
}

func Production(opts ...zap.Option) (*zap.Logger, error) {
	cfg := config()
	cfg.DisableStacktrace = true
	return cfg.Build(opts...)
}

func config() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

package config

import (
	"time"

	"github.com/caarlos0/env/v7"
)

type Telemetry struct {
	ZipkinURL   string  `env:"ZIPKIN_TRACER_URL"`
	Ratio       float64 `env:"ZIPKIN_TRACER_RATIO" envDefault:"1"`
	MetricsPort int     `env:"METRICS_PORT" envDefault:"2112"`
}

type PostgresConfig struct {
	PostgresURL     string        `env:"POSTGRES_URL,notEmpty"`
	MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"2"`
	ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME"`
	ConnMaxIdleTime time.Duration `env:"POSTGRES_CONN_IDLE_TIME"`
}

type Config struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"merchants-service"`
	ServerPort  int    `env:"SERVER_PORT" envDefault:"8080"`
	Environment string `env:"ENV" envDefault:"dev"`
	Postgres    PostgresConfig
	Telemetry   Telemetry
}

func Build() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

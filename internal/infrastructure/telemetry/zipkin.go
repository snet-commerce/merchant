package telemetry

import (
	"log"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

type tracerConfigFunc func(cfg *tracerConfig)

type tracerConfig struct {
	ratio   *float64
	logger  *log.Logger
	service string
}

func WithTracerRatio(ratio float64) tracerConfigFunc {
	return func(cfg *tracerConfig) {
		cfg.ratio = &ratio
	}
}

func WithTracerLogger(logger *log.Logger) tracerConfigFunc {
	return func(cfg *tracerConfig) {
		cfg.logger = logger
	}
}

func WithTracerServiceName(s string) tracerConfigFunc {
	return func(cfg *tracerConfig) {
		cfg.service = s
	}
}

func ZipkinTracer(url string, opts ...tracerConfigFunc) (*sdktrace.TracerProvider, error) {
	var cfg tracerConfig
	for _, optFn := range opts {
		if optFn != nil {
			optFn(&cfg)
		}
	}

	zipkinOpts := make([]zipkin.Option, 0)
	if cfg.logger != nil {
		zipkinOpts = []zipkin.Option{zipkin.WithLogger(cfg.logger)}
	}

	exporter, err := zipkin.New(url, zipkinOpts...)
	if err != nil {
		return nil, err
	}

	rscAttrs := []attribute.KeyValue{attribute.String("exporter", "zipkin")}
	if cfg.service != "" {
		rscAttrs = append(rscAttrs, semconv.ServiceName(cfg.service))
	}

	tracerOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter), // TODO: think of batch config
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, rscAttrs...)),
	}

	if cfg.ratio != nil {
		tracerOpts = append(tracerOpts, sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(*cfg.ratio))))
	}

	prv := sdktrace.NewTracerProvider(tracerOpts...)

	return prv, nil
}

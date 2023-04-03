package telemetry

//import (
//	"go.opentelemetry.io/otel/attribute"
//	"go.opentelemetry.io/otel/exporters/prometheus"
//	"go.opentelemetry.io/otel/sdk/metric"
//	"go.opentelemetry.io/otel/sdk/resource"
//	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
//)
//
//type meterConfigFunc func(cfg *meterConfig)
//
//type meterConfig struct {
//	service string
//}
//
//func WithMeterServiceName(s string) meterConfigFunc {
//	return func(cfg *meterConfig) {
//		cfg.service = s
//	}
//}
//
//func PrometheusMeter(opts ...meterConfigFunc) (*metric.MeterProvider, error) {
//	var cfg meterConfig
//	for _, optFn := range opts {
//		if optFn != nil {
//			optFn(&cfg)
//		}
//	}
//
//	exporter, err := prometheus.New()
//	if err != nil {
//		return nil, err
//	}
//
//	rscAttrs := []attribute.KeyValue{attribute.String("exporter", "prometheus")}
//	if cfg.service != "" {
//		rscAttrs = append(rscAttrs, semconv.ServiceName(cfg.service))
//	}
//
//	prv := metric.NewMeterProvider(
//		metric.WithReader(exporter),
//		metric.WithResource(resource.NewWithAttributes(semconv.SchemaURL, rscAttrs...)),
//	)
//
//	return prv, nil
//}

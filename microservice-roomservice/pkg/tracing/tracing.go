package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"microservice-roomservice/pkg/logger"
)

// TracerConfig holds configuration for OpenTelemetry tracer
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	Enabled        bool
}

// InitTracer initializes OpenTelemetry tracer with OTLP exporter
func InitTracer(cfg TracerConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		logger.Info().Msg("OpenTelemetry tracing is disabled")
		return func(ctx context.Context) error { return nil }, nil
	}

	logger.Info().Msgf("Initializing OpenTelemetry tracer: service=%s, endpoint=%s", cfg.ServiceName, cfg.OTLPEndpoint)

	// Create resource with service information (simplified for container environments)
	res := resource.NewWithAttributes(
		"",
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		semconv.DeploymentEnvironmentKey.String(cfg.Environment),
	)

	// Create OTLP HTTP exporter (works better in same-node scenarios)
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use insecure for internal cluster communication
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to W3C Trace Context and Baggage
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	logger.Info().Msg("OpenTelemetry tracer initialized successfully")

	// Return shutdown function
	return func(ctx context.Context) error {
		logger.Info().Msg("Shutting down OpenTelemetry tracer...")
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
		logger.Info().Msg("OpenTelemetry tracer shutdown complete")
		return nil
	}, nil
}

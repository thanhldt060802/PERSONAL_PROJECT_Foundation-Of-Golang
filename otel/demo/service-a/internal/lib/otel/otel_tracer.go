package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig configures the distributed tracing component.
type TracerConfig struct {
	ServiceName    string            // Name of the service
	ServiceVersion string            // Version of the service
	EndPoint       string            // OTLP endpoint for exporting tracing data
	Insecure       bool              // Allow HTTP schema, instead of HTTPS
	HttpHeader     map[string]string // Additional HTTP headers
}

// initTracer initializes the Trace, returns Tracer and a cleanup function.
// Spans are exported using OTLP HTTP protocol with batch processing.
func initTracer(config *TracerConfig) (trace.Tracer, func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.EndPoint),
	}
	if config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(config.HttpHeader) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(config.HttpHeader))
	}

	// Create OTLP HTTP exporter for sending traces
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		stdLog.Fatalf("[error] Failed to create exporter for Tracer: %v", err)
	}

	// Create resource with service metadata
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.ServiceVersion),
		attribute.String("host.ip", getLocalIP()),
	)

	// Create Tracer provider with batch span processor for efficient export
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)

	// Configure trace context propagation for cross-service tracing (HTTP, gRPC)
	// This enables distributed tracing across service boundaries
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Init Tracer, cleanup function for Tracer
	tracer := otel.Tracer(config.ServiceName + "/otel")
	shutdown := func(ctx context.Context) {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			stdLog.Printf("[error] Failed to shut down Tracer provider: %v", err)
		}
	}

	// Return Tracer and cleanup function for Tracer
	return tracer, shutdown
}

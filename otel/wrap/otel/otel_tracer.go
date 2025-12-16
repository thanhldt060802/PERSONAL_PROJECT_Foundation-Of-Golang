package otel

import (
	"context"
	"sync"
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

var (
	// tracer is the global tracer instance for creating spans
	tracer trace.Tracer
	// tracerOnce makes sure tracer instance only one time
	tracerOnce sync.Once
)

// initTracer initializes the OpenTelemetry tracer with OTLP exporter
// and configures trace context propagation for distributed tracing.
//
// Parameters:
//   - config: Configuration including service info and OTLP endpoint
//
// Returns:
//   - func(ctx context.Context): A cleanup function to shutdown the tracer provider

func initTracer(config *ObserverConfig) func(ctx context.Context) {
	var shutdown func(ctx context.Context)

	tracerOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create OTLP HTTP exporter for sending traces
		exporter, err := otlptracehttp.New(
			ctx,
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint(config.EndPoint),
		)
		if err != nil {
			stdLog.Fatalf("Failed to create exporter for Tracer: %v", err)
		}

		// Create resource with service metadata
		resource := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			attribute.String("service.instance.ip", getLocalIP()),
		)

		// Create tracer provider with batch span processor for efficient export
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

		tracer = otel.Tracer(config.ServiceName + "/otel")

		shutdown = func(ctx context.Context) {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				stdLog.Printf("Error occurred when shutting down Tracer provider: %v", err)
			}
		}
	})

	// Return cleanup function
	return shutdown
}

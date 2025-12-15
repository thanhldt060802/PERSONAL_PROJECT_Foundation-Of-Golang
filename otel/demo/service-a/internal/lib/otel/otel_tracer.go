package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
)

// INIT TRACER

func initTracer(config *ObserverConfig) func(ctx context.Context) {
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(config.EndPoint),
	)
	if err != nil {
		stdLog.Fatalf("Failed to create exporter for Tracer: %v", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)

	// Set policy for cross-service (HTTP, gRPC)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	tracer = otel.Tracer(config.ServiceName + "/observer")

	return func(ctx context.Context) {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			stdLog.Printf("Error occurred when shutting down Tracer provider: %v", err)
		}
	}
}

package otel

import (
	"context"
	"fmt"
	stdLog "log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

func initTracer(config *ObserverEndPointConfig) func() {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(fmt.Sprintf("%v:%v", config.Host, config.Port)),
	)
	if err != nil {
		stdLog.Fatalf("Failed to create exporter for Tracer: %v", err.Error())
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	) // Important for cross-service and pub/sub system

	return func() {
		tracerProvider.Shutdown(ctx)
	}
}

package opentelemetry

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

var ShutdownTracer func() = func() {}

type TracerEndPointConfig struct {
	ServiceName string
	Host        string
	Port        int
}

func NewTracer(config TracerEndPointConfig) func() {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(fmt.Sprintf("%v:%v", config.Host, config.Port)),
	)
	if err != nil {
		log.Fatalf("Create exporter failed: %v", err.Error())
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

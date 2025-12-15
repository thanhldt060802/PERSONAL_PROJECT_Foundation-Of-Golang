package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// DEFINE HYBRID SPAN

func NewHybridSpan(ctx context.Context, operation string) (context.Context, *HybridSpan) {
	ctxSpan, span := tracer.Start(ctx, operation, trace.WithTimestamp(time.Now()))

	hybridSpan := HybridSpan{
		coreSpan:       span,
		ctx:            ctxSpan,
		spanAttributes: make(map[string]any),
	}
	return ctxSpan, &hybridSpan
}

type HybridSpan struct {
	coreSpan trace.Span

	ctx context.Context
	err error

	spanAttributes map[string]any
}

// DEFINE STANDARD FEATURE FOR SPAN

func (span *HybridSpan) Done() {
	attrs := mapToAttribute(span.spanAttributes)
	span.coreSpan.SetAttributes(attrs...)

	if span.err != nil {
		span.coreSpan.RecordError(span.err)
		span.coreSpan.SetStatus(codes.Error, span.err.Error())
		span.coreSpan.End(trace.WithStackTrace(true))
	} else {
		span.coreSpan.SetStatus(codes.Ok, "success")
		span.coreSpan.End()
	}
}

func (span *HybridSpan) Context() context.Context {
	return span.ctx
}

func (span *HybridSpan) SetError(err error) {
	span.err = err
}

// DEFINE ADDITIONAL FEATURE FOR SPAN

func (span *HybridSpan) SetAttribute(key string, value any) {
	span.spanAttributes[key] = value
}

func (span *HybridSpan) AddEvent(eventName string, eventAttributes map[string]any) {
	attrs := mapToAttribute(eventAttributes)
	span.coreSpan.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// CROSS PUB/SUB SYSTEM FEATURE DEFINITION FOR SPAN

type TraceCarrier propagation.MapCarrier

func (span *HybridSpan) ExportTraceCarrier() TraceCarrier {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(span.ctx, carrier)

	return TraceCarrier(carrier)
}

func (traceCarrier TraceCarrier) ExtractContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(traceCarrier))
}

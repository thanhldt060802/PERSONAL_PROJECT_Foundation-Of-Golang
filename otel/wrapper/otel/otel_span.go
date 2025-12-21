package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func NewSpan(ctx context.Context, operation string) (context.Context, *Span) {
	spanCtx, coreSpan := tracer.Start(ctx, operation, trace.WithTimestamp(time.Now()))

	span := Span{
		coreSpan:       coreSpan,
		parentCtx:      ctx,
		spanCtx:        spanCtx,
		spanAttributes: make(map[string]any),
	}
	return spanCtx, &span
}

type Span struct {
	coreSpan trace.Span // The underlying OpenTelemetry span

	parentCtx context.Context // Parent context of this span
	spanCtx   context.Context // Context containing this span
	err       error           // Error to be recorded when span ends

	spanAttributes map[string]any // Attributes to be added to the span
}

func (span *Span) Done() {
	// Convert and set all accumulated attributes
	attrs := mapToAttribute(span.spanAttributes)
	span.coreSpan.SetAttributes(attrs...)

	if span.err != nil {
		// Record error and set error status
		span.coreSpan.RecordError(span.err)
		span.coreSpan.SetStatus(codes.Error, span.err.Error())
		span.coreSpan.End(trace.WithStackTrace(true))
	} else {
		// Set success status
		span.coreSpan.SetStatus(codes.Ok, "success")
		span.coreSpan.End()
	}
}

func (span *Span) ParentContext() context.Context {
	return span.parentCtx
}

func (span *Span) Context() context.Context {
	return span.spanCtx
}

func (span *Span) SetError(err error) {
	span.err = err
}

func (span *Span) SetAttribute(key string, value any) {
	span.spanAttributes[key] = value
}

func (span *Span) AddEvent(eventName string, eventAttributes map[string]any) {
	attrs := mapToAttribute(eventAttributes)
	span.coreSpan.AddEvent(eventName, trace.WithAttributes(attrs...))
}

type TraceCarrier propagation.MapCarrier

func ExportTraceCarrier(ctx context.Context) TraceCarrier {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return TraceCarrier(carrier)
}

func (traceCarrier TraceCarrier) ExtractContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(traceCarrier))
}

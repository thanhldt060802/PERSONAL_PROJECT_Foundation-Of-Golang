package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// CUSTOM SPAN

type CustomSpan struct {
	trace.Span
	Err error
}

func (span *CustomSpan) End() {
	if span.Err != nil {
		span.RecordError(span.Err)
		span.SetStatus(codes.Error, span.Err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	span.Span.End()
}

// INTERNAL USAGE

func StartSpanInternal(ctx context.Context) (context.Context, *CustomSpan) {
	modulePath, actionName := callbackInfo()
	ctx, span := otel.Tracer(modulePath).Start(ctx, actionName)

	customSpan := CustomSpan{
		Span: span,
	}
	return ctx, &customSpan
}

// PUB/SUB SYSTEM USAGE

type MessageTracing struct {
	TraceContext propagation.MapCarrier
	Payload      any
}

func (msgTrace *MessageTracing) ExtractTraceContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(msgTrace.TraceContext))
}

package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

/*
`CustomSpan` is based on `trace.Span`.
*/
type CustomSpan struct {
	trace.Span
	Err error
}

/*
This function is used to close span on each context.
*/
func (span *CustomSpan) End() {
	if span.Err != nil {
		span.RecordError(span.Err)
		span.SetStatus(codes.Error, span.Err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	span.Span.End()
}

/*
This function is used to create span for internal tracing.
*/
func StartSpanInternal(ctx context.Context) (context.Context, *CustomSpan) {
	modulePath, actionName := callbackInfo()
	ctx, span := otel.Tracer(modulePath).Start(ctx, actionName)

	customSpan := CustomSpan{
		Span: span,
	}
	return ctx, &customSpan
}

/*
This function is used to create span for cross-service tracing, ctx will be auto injected into request heeader,
by the way in another service we can recreate a span with the ctx information from the extracted header.
*/
func StartSpanCrossService(ctx context.Context, method string, url string) (context.Context, *CustomSpan, *http.Request, error) {
	modulePath, actionName := callbackInfo()
	ctx, span := otel.Tracer(modulePath).Start(ctx, actionName)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, nil, nil, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	customSpan := CustomSpan{
		Span: span,
	}
	return ctx, &customSpan, req, nil
}

/*
`MessageTracing` is used for pub/sub system tracing.
*/
type MessageTracing struct {
	SpanContext propagation.MapCarrier
	Payload     any
}

/*
This function is used to inject ctx into `SpanContext` of `MessageTracing`, then we can continue send `MessageTracing` via pub/sub system.
*/
func (msgTrace *MessageTracing) Inject(ctx context.Context) {
	otel.GetTextMapPropagator().Inject(ctx, msgTrace.SpanContext)
}

/*
This function is used to extract context from `SpanContext` of `MessageTracing` when consumed by consumer on pub/sub system,
from which we can use StartSpanInternal to create span for internal tracing.
*/
func (msgTrace *MessageTracing) ExtractSpanContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(msgTrace.SpanContext))
}

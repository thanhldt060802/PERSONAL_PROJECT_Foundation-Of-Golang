package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type IHybridSpan interface {
	// End span to end collection flow
	End()

	// Inject to header of request to receiver service will detect chain context
	InjectToRequestHeader(rHeader http.Header)
	// Export trace carrier to add into payload which will be sent via Pub/Sub system
	ExportTraceCarrier() TraceCarrier

	// Continue declare features for Logger
	// ...

	// Continue declare features for Meter
	// ...
}

type HybridSpan struct {
	Ctx   context.Context
	Error error

	trace.Span
	// Logger
	// Metric
}

func NewHybridSpan(ctx context.Context) (context.Context, *HybridSpan) {
	modulePath, actionName := callbackInfo()
	ctxSpan, span := otel.Tracer(modulePath).Start(ctx, actionName)

	hybridSpan := HybridSpan{
		Ctx:  ctxSpan,
		Span: span,
	}
	return ctxSpan, &hybridSpan
}

func (span *HybridSpan) End() {
	if span.Error != nil {
		span.RecordError(span.Error)
		span.SetStatus(codes.Error, span.Error.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	span.Span.End()
}

func (span *HybridSpan) InjectToRequestHeader(rHeader http.Header) {
	otel.GetTextMapPropagator().Inject(span.Ctx, propagation.HeaderCarrier(rHeader))
}

func (span *HybridSpan) ExportTraceCarrier() TraceCarrier {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(span.Ctx, carrier)

	return TraceCarrier(carrier)
}

type TraceCarrier propagation.MapCarrier

func (traceCarrier TraceCarrier) ExtractContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(traceCarrier))
}

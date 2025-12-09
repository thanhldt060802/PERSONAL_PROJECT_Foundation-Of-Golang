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

	Info(format string, args ...any)
	Warn(format string, args ...any)
	Debug(format string, args ...any)
	Error(format string, args ...any)

	// Meter feature
	// ...
}

type HybridSpan struct {
	Ctx context.Context
	Err error

	trace.Span
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
	if span.Err != nil {
		span.RecordError(span.Err)
		span.SetStatus(codes.Error, span.Err.Error())
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

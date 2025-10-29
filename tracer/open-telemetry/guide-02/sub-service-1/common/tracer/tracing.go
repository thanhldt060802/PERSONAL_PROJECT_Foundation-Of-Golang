package tracer

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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

// CROSS-SERVICE USAGE

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

	customSpan := CustomSpan{
		Span: span,
	}
	return ctx, &customSpan, req, nil
}

package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// NewSpan creates a new tracing Span for the given operation.
// Returns the Span context and a Span wrapper that must be closed with Done().
//
// Example:
//
//	ctx, span := observer.NewSpan(ctx, "database.query")
//	defer span.Done()
//	span.SetAttribute("query", "SELECT * FROM users")
func (o *Observer) NewSpan(ctx context.Context, operation string) (context.Context, *Span) {
	spanCtx, coreSpan := o.tracer.Start(ctx, operation, trace.WithTimestamp(time.Now()))

	span := Span{
		coreSpan:       coreSpan,
		parentCtx:      ctx,
		spanCtx:        spanCtx,
		spanAttributes: make(map[string]any),
	}
	return spanCtx, &span
}

// Span wraps an OpenTelemetry Span with additional functionality.
// Attributes and errors are accumulated and applied when Done() is called.
type Span struct {
	coreSpan trace.Span // The underlying OpenTelemetry Span

	parentCtx context.Context // Parent context of this Span
	spanCtx   context.Context // Context containing this Span
	err       error           // Error to be recorded when Span ends

	spanAttributes map[string]any // Attributes to be added to the Span
}

// Done finalizes the Span by:
//   - Applying all accumulated attributes
//   - Recording any error and setting error status
//   - Ending the Span with timestamp
//
// Must be called to ensure Span is exported.
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

// ParentContext returns the context before this Span was created.
// Useful for creating sibling spans instead of child spans.
func (span *Span) ParentContext() context.Context {
	return span.parentCtx
}

// Context returns the context containing this Span.
// Use this context to create child spans or propagate trace context.
func (span *Span) Context() context.Context {
	return span.spanCtx
}

// SetError marks the Span as failed.
// The error will be recorded when Done() is called.
func (span *Span) SetError(err error) {
	span.err = err
}

// SetAttribute adds a key-value attribute to the Span.
// Attributes provide additional context about the operation.
// Common attributes: user_id, request_id, http.status_code, db.statement
func (span *Span) SetAttribute(key string, value any) {
	span.spanAttributes[key] = value
}

// AddEvent records a point-in-time event within the Span.
// Useful for marking important moments like cache hits or retry attempts.
//
// Example:
//
//	span.AddEvent("cache.hit", map[string]any{"key": "user:123"})
func (span *Span) AddEvent(eventName string, eventAttributes map[string]any) {
	attrs := mapToAttribute(eventAttributes)
	span.coreSpan.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// TraceCarrier wraps trace context for propagation across process boundaries.
// Used with message queues, job systems, or any async communication.
type TraceCarrier propagation.MapCarrier

// ExportTraceCarrier extracts trace context from the given context.
// The returned TraceCarrier can be serialized and sent to another service.
//
// Example:
//
//	carrier := otel.ExportTraceCarrier(ctx)
//	// Send carrier via message queue, store in Redis, etc.
func ExportTraceCarrier(ctx context.Context) TraceCarrier {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return TraceCarrier(carrier)
}

// ExtractContext recreates a context from the trace carrier.
// Use this to continue the trace in another service or async job.
//
// Example:
//
//	ctx := carrier.ExtractContext()
//	ctx, span := otel.NewSpan(ctx, "AsyncJob")
func (traceCarrier TraceCarrier) ExtractContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(traceCarrier))
}

// IsZero reports whether the TraceCarrier contains no propagation data.
//
// It returns true when the carrier is either nil or empty (len == 0).
// In both cases, the carrier is considered to have no trace context.
func (traceCarrier TraceCarrier) IsZero() bool {
	return len(traceCarrier) == 0
}

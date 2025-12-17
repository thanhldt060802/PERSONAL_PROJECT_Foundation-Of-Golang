package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// NewSpan creates a new tracing span with enhanced functionality.
// It wraps the standard OpenTelemetry span with additional features for easier use.
//
// Parameters:
//   - ctx: Parent context
//   - operation: Name of the operation being traced
//
// Returns:
//   - context.Context: New context containing the span
//   - *Span: The created span wrapper
//
// Example:
//
//	ctx, span := NewSpan(ctx, "database.query")
//	defer span.Done()
//
//	span.SetAttribute("query", "SELECT * FROM users")
//	// ... perform operation ...
//	if err != nil {
//	    span.SetError(err)
//	}
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

// Span is a wrapper around OpenTelemetry's trace.Span that provides
// additional functionality and a simpler API for common tracing operations.
type Span struct {
	coreSpan trace.Span // The underlying OpenTelemetry span

	parentCtx context.Context // Parent context of this span
	spanCtx   context.Context // Context containing this span
	err       error           // Error to be recorded when span ends

	spanAttributes map[string]any // Attributes to be added to the span
}

// Done finalizes the span by setting all attributes, recording any errors,
// and properly ending the span with appropriate status.
// This method should always be called when the span is complete, typically using defer.
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

// ParentContext returns the parent context of this span.
// Use this context for same level operations to maintain trace hierarchy.
func (span *Span) ParentContext() context.Context {
	return span.spanCtx
}

// Context returns the context containing this span.
// Use this context for child level operations to maintain trace hierarchy.
func (span *Span) Context() context.Context {
	return span.spanCtx
}

// SetError sets an error to be recorded when the span ends.
// The error will cause the span to be marked with error status.
//
// Parameters:
//   - err: The error that occurred during the operation
func (span *Span) SetError(err error) {
	span.err = err
}

// SetAttribute adds an attribute to the span.
// Attributes are metadata that provide additional context about the operation.
//
// Parameters:
//   - key: Attribute name
//   - value: Attribute value (supports various types)
func (span *Span) SetAttribute(key string, value any) {
	span.spanAttributes[key] = value
}

// AddEvent adds a timed event to the span.
// Events represent significant moments during the span's lifetime.
//
// Parameters:
//   - eventName: Name of the event
//   - eventAttributes: Additional attributes for the event
func (span *Span) AddEvent(eventName string, eventAttributes map[string]any) {
	attrs := mapToAttribute(eventAttributes)
	span.coreSpan.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// TraceCarrier is a type alias for propagation.MapCarrier used to
// serialize and transport trace context across service boundaries.
type TraceCarrier propagation.MapCarrier

// ExportTraceCarrier exports the trace context into a carrier format
// that can be transmitted to other services by Pub/Sub environment.
//
// Returns:
//   - TraceCarrier: Serialized trace context
//
// Example:
//
//	carrier := ExportTraceCarrier(ctx)
//	// Assign to payload and publish via Pub/Sub environment
func ExportTraceCarrier(ctx context.Context) TraceCarrier {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return TraceCarrier(carrier)
}

// ExtractContext extracts a context from the trace carrier.
// Use this on the receiving service to continue the trace.
//
// Returns:
//   - context.Context: Context containing the extracted trace information
//
// Example:
//
//	ctx := carrier.ExtractContext()
//	// Use ctx for operations in the receiving service
func (traceCarrier TraceCarrier) ExtractContext() context.Context {
	return otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(traceCarrier))
}

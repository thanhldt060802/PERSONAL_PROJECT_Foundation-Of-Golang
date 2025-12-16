package otel

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	// logger is the global structured logger instance
	logger *slog.Logger
	// loggerOnce makes sure logger instance only one time
	loggerOnce sync.Once
)

// LogLevel represents the severity level of log messages
type LogLevel string

const (
	LOG_LEVEL_INFO  LogLevel = "info"  // Informational messages
	LOG_LEVEL_WARN  LogLevel = "warn"  // Warning messages
	LOG_LEVEL_DEBUG LogLevel = "debug" // Debug messages for development
	LOG_LEVEL_ERROR LogLevel = "error" // Error messages
)

// initLogger initializes the OpenTelemetry logger with both remote exporter and optional local file logging.
// It creates a multi-handler logger that can write to both OpenTelemetry collector and local files.
//
// Parameters:
//   - config: Configuration for the logger including service info and log file settings
//
// Returns:
//   - func(ctx context.Context): A cleanup function to shutdown the logger provider
func initLogger(config *ObserverConfig) func(ctx context.Context) {
	var shutdown func(ctx context.Context)

	loggerOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create OTLP HTTP exporter for sending logs to OpenTelemetry collector
		exporter, err := otlploghttp.New(
			ctx,
			otlploghttp.WithInsecure(),
			otlploghttp.WithEndpoint(config.EndPoint),
		)
		if err != nil {
			stdLog.Fatalf("Failed to create exporter for Logger: %v", err.Error())
		}

		// Create resource with service metadata
		resource := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
		)

		// Create logger provider with batch processor for efficient log export
		loggerProvider := log.NewLoggerProvider(
			log.WithProcessor(log.NewBatchProcessor(exporter)),
			log.WithResource(resource),
		)

		// Create OpenTelemetry slog handler
		otelHandler := otelslog.NewHandler(
			config.ServiceName,
			otelslog.WithLoggerProvider(loggerProvider),
		)

		multiHandler := []slog.Handler{
			otelHandler,
		}

		writers := []io.Writer{os.Stdout}

		// Configure log level for local handler
		localHandlerOption := slog.HandlerOptions{}
		switch config.LocalLogLevel {
		case LOG_LEVEL_INFO:
			{
				localHandlerOption.Level = slog.LevelInfo
			}
		case LOG_LEVEL_WARN:
			{
				localHandlerOption.Level = slog.LevelWarn
			}
		case LOG_LEVEL_DEBUG:
			{
				localHandlerOption.Level = slog.LevelDebug
			}
		case LOG_LEVEL_ERROR:
			{
				localHandlerOption.Level = slog.LevelError
			}
		default:
			{
				localHandlerOption.Level = slog.LevelInfo
			}
		}

		var logFile *os.File
		// Setup local file logging
		if config.LocalLogFile != "" {
			// Create log directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(config.LocalLogFile), 0755); err != nil {
				stdLog.Fatalf("Failed to create local log file dir for Logger: %v", err.Error())
			}

			// Open log file for writing
			file, err := os.OpenFile(config.LocalLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				stdLog.Fatalf("Failed to open local log file for Logger: %v", err.Error())
			}
			logFile = file
			writers = append(writers, logFile)
		}

		// Write to both stdout and file
		multiWriter := io.MultiWriter(writers...)

		// Create JSON handler for local logging
		fileHandler := slog.NewJSONHandler(multiWriter, &localHandlerOption)
		multiHandler = append(multiHandler, fileHandler)

		// Init logger with multi handler
		logger = slog.New(newMultiHandler(multiHandler...))

		shutdown = func(ctx context.Context) {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				stdLog.Printf("Error occurred when shutting down Logger provider: %v", err)
			}
			if logFile != nil {
				logFile.Close()
			}
		}
	})

	// Return cleanup function
	return shutdown
}

// multiHandler is a custom slog.Handler that dispatches log records to multiple handlers.
// It automatically enriches log records with trace information and client IP.
type multiHandler struct {
	handlers []slog.Handler
}

// newMultiHandler creates a new multiHandler with the given handlers
func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

// Enabled reports whether any of the handlers will handle the given level
func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle enriches the log record with tracing and client IP information,
// then dispatches it to all registered handlers
func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	traceID, spanID := getTraceInfo(ctx)
	clientIP := getClientIpFromCtx(ctx)

	// Clone and enrich the record with additional attributes
	r := record.Clone()
	r.AddAttrs(
		slog.String("trace_id", traceID),
		slog.String("span_id", spanID),
		slog.String("client_ip", clientIP),
	)

	// Dispatch to all handlers
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the provided attributes
func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

// WithGroup returns a new Handler with the given group name
func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// InfoLog logs an informational message with automatic source file metadata.
// The message is formatted using fmt.Sprintf with the provided format and arguments.
//
// Parameters:
//   - ctx: Context for trace correlation
//   - format: Format string for the log message
//   - args: Arguments to format into the message
func InfoLog(ctx context.Context, format string, args ...any) {
	logWithMeta(ctx, slog.LevelInfo, format, args...)
}

// WarnLog logs a warning message with automatic source file metadata.
// The message is formatted using fmt.Sprintf with the provided format and arguments.
//
// Parameters:
//   - ctx: Context for trace correlation
//   - format: Format string for the log message
//   - args: Arguments to format into the message
func WarnLog(ctx context.Context, format string, args ...any) {
	logWithMeta(ctx, slog.LevelWarn, format, args...)
}

// DebugLog logs a debug message with automatic source file metadata.
// The message is formatted using fmt.Sprintf with the provided format and arguments.
//
// Parameters:
//   - ctx: Context for trace correlation
//   - format: Format string for the log message
//   - args: Arguments to format into the message
func DebugLog(ctx context.Context, format string, args ...any) {
	logWithMeta(ctx, slog.LevelDebug, format, args...)
}

// ErrorLog logs an error message with automatic source file metadata.
// The message is formatted using fmt.Sprintf with the provided format and arguments.
//
// Parameters:
//   - ctx: Context for trace correlation
//   - format: Format string for the log message
//   - args: Arguments to format into the message
func ErrorLog(ctx context.Context, format string, args ...any) {
	logWithMeta(ctx, slog.LevelError, format, args...)
}

// logWithMeta logs an message with level and automatic source file metadata.
// The message is formatted using fmt.Sprintf with the provided format and arguments.
//
// Parameters:
//   - ctx: Context for trace correlation
//   - level: Level for the log message
//   - format: Format string for the log message
//   - args: Arguments to format into the message
func logWithMeta(ctx context.Context, level slog.Level, format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(2)
	srcFile := filepath.Base(path)
	meta := fmt.Sprintf("%s:%d", srcFile, numLine)
	msg := fmt.Sprintf(format, args...)
	logger.LogAttrs(
		ctx,
		level,
		msg,
		slog.String("meta", meta),
	)
}

// getClientIpFromCtx retrieves the client IP address from the context.
// Returns an empty string if the context is nil or the IP is not found.
func getClientIpFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	ip, _ := ctx.Value(ClientIP).(string)
	return ip
}

// getTraceInfo extracts the trace ID and span ID from the context.
// Returns empty strings if no valid trace context is found.
func getTraceInfo(ctx context.Context) (string, string) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return "", ""
	}
	spanContext := span.SpanContext()
	return spanContext.TraceID().String(), spanContext.SpanID().String()
}

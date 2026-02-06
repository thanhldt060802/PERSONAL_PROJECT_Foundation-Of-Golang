package otel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

// Error definitions for Logger.
var (
	// ErrLoggerUnconfigured occurs when using Logger without including Logger option when initializing Otel Observer.
	ErrLoggerUnconfigured = errors.New("logger is unconfigured")
)

// LogLevel defines the severity level for logging using Logger.
type LogLevel string

// Log level definitions for Logger.
const (
	// LOG_LEVEL_INFO is used for informational messages.
	LOG_LEVEL_INFO LogLevel = "info"
	// LOG_LEVEL_WARN is used for warning messages.
	LOG_LEVEL_WARN LogLevel = "warn"
	// LOG_LEVEL_DEBUG is used for debug messages for development.
	LOG_LEVEL_DEBUG LogLevel = "debug"
	// LOG_LEVEL_ERROR is used for error messages.
	LOG_LEVEL_ERROR LogLevel = "error"
)

// LoggerConfig configures structured logging with OpenTelemetry integration.
type LoggerConfig struct {
	ServiceName    string            // Name of the service
	ServiceVersion string            // Version of the service
	EndPoint       string            // OTLP endpoint for exporting log data
	Insecure       bool              // Allow HTTP schema, instead of HTTPS
	HttpHeader     map[string]string // Additional HTTP headers

	LocalLogFile  string   // Path to local log file
	LocalLogLevel LogLevel // Log level for local file logging
}

// initLogger initializes the Logger, returns Logger and a cleanup function.
// Logs are sent to both OTLP endpoint and local output (stdout + optional file).
// Each log entry includes trace_id and span_id for correlation with traces.
func initLogger(config *LoggerConfig) (*slog.Logger, func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(config.EndPoint),
	}
	if config.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if len(config.HttpHeader) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(config.HttpHeader))
	}

	// Create OTLP HTTP exporter for sending logs to OpenTelemetry collector
	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		stdLog.Fatalf("[error] Failed to create exporter for Logger: %v", err.Error())
	}

	// Create resource with service metadata
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.ServiceVersion),
		attribute.String("host.ip", getLocalIP()),
	)

	// Create Logger provider with batch processor for efficient log export
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
			stdLog.Fatalf("[error] Failed to create local log file dir for Logger: %v", err.Error())
		}

		// Open log file for writing
		file, err := os.OpenFile(config.LocalLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			stdLog.Fatalf("[error] Failed to open local log file for Logger: %v", err.Error())
		}
		logFile = file
		writers = append(writers, logFile)
	}

	// Write to both stdout and file
	multiWriter := io.MultiWriter(writers...)

	// Create JSON handler for local logging
	localHandler := slog.NewJSONHandler(multiWriter, &localHandlerOption)
	multiHandler = append(multiHandler, localHandler)

	// Init Logger with multi handler, cleanup function for Logger
	logger := slog.New(newMultiHandler(multiHandler...))
	shutdown := func(ctx context.Context) {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			stdLog.Printf("[error] Failed to shut down Logger provider: %v", err)
		}
		if logFile != nil {
			logFile.Close()
		}
	}

	// Return Logger and cleanup function for Logger
	return logger, shutdown
}

// multiHandler dispatches log records to multiple handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

// Enabled returns true if any handler is enabled for the given level.
func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle enriches the log record with trace context and dispatches to all handlers.
func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	traceID, spanID := getTraceInfo(ctx)

	// Clone and enrich the record with additional attributes
	r := record.Clone()
	r.AddAttrs(
		slog.String("trace_id", traceID),
		slog.String("span_id", spanID),
	)

	// Dispatch to all handlers
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// Context-aware logging functions.
// These functions extract trace_id and span_id from context automatically.

// InfoLogWithCtx logs an informational message with trace context.
func (o *Observer) InfoLogWithCtx(ctx context.Context, format string, args ...any) {
	o.logWithMeta(ctx, slog.LevelInfo, format, args...)
}

// WarnLogWithCtx logs a warning message with trace context.
func (o *Observer) WarnLogWithCtx(ctx context.Context, format string, args ...any) {
	o.logWithMeta(ctx, slog.LevelWarn, format, args...)
}

// DebugLogWithCtx logs a debug message with trace context.
func (o *Observer) DebugLogWithCtx(ctx context.Context, format string, args ...any) {
	o.logWithMeta(ctx, slog.LevelDebug, format, args...)
}

// ErrorLogWithCtx logs an error message with trace context.
func (o *Observer) ErrorLogWithCtx(ctx context.Context, format string, args ...any) {
	o.logWithMeta(ctx, slog.LevelError, format, args...)
}

// Context-less logging functions.
// Use these when context is not available.

// InfoLog logs an informational message without trace context.
func (o *Observer) InfoLog(format string, args ...any) {
	o.logWithMeta(context.Background(), slog.LevelInfo, format, args...)
}

// WarnLog logs a warning message without trace context.
func (o *Observer) WarnLog(format string, args ...any) {
	o.logWithMeta(context.Background(), slog.LevelWarn, format, args...)
}

// DebugLog logs a debug message without trace context.
func (o *Observer) DebugLog(format string, args ...any) {
	o.logWithMeta(context.Background(), slog.LevelDebug, format, args...)
}

// ErrorLog logs an error message without trace context.
func (o *Observer) ErrorLog(format string, args ...any) {
	o.logWithMeta(context.Background(), slog.LevelError, format, args...)
}

// logWithMeta adds source file location to log entries.
func (o *Observer) logWithMeta(ctx context.Context, level slog.Level, format string, args ...any) {
	if o.logger == nil {
		stdLog.Printf("[error] Failed to use Logger: %v", ErrLoggerUnconfigured)
		return
	}

	_, path, numLine, _ := runtime.Caller(2)
	srcFile := filepath.Base(path)
	meta := fmt.Sprintf("%s:%d", srcFile, numLine)
	msg := fmt.Sprintf(format, args...)
	o.logger.LogAttrs(
		ctx,
		level,
		msg,
		slog.String("meta", meta),
	)
}

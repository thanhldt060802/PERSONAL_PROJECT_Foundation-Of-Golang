package otel

import (
	"context"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

var logger *slog.Logger

func initLogger(config *ObserverEndPointConfig) func() {
	ctx := context.Background()

	exporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(fmt.Sprintf("%v:%v", config.Host, config.Port)),
	)
	if err != nil {
		stdLog.Fatalf("Failed to create exporter for Logger: %v", err.Error())
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(resource),
	)

	otelHandler := otelslog.NewHandler(
		config.ServiceName,
		otelslog.WithLoggerProvider(loggerProvider),
	)

	multiHandler := []slog.Handler{
		otelHandler,
	}

	var logFile *os.File
	if config.LocalLogFile != "" {
		if err := os.Mkdir(filepath.Dir(config.LocalLogFile), 0755); err != nil {
			stdLog.Fatalf("Failed to create local log file dir for Logger: %v", err.Error())
		}
		file, err := os.OpenFile(config.LocalLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			stdLog.Fatalf("Failed to open local log file for Logger: %v", err.Error())
		}
		logFile = file

		multiWriter := io.MultiWriter(os.Stdout, file)

		fileHandler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		multiHandler = append(multiHandler, fileHandler)
	}

	logger = slog.New(newMultiHandler(multiHandler...))

	return func() {
		loggerProvider.Shutdown(ctx)
		if logFile != nil {
			logFile.Close()
		}
	}
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record); err != nil {
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

func (span *HybridSpan) Info(format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	logger.With("meta", fmt.Sprintf("%s:%d", srcFile, numLine)).InfoContext(span.Ctx, format, args...)
}

func (span *HybridSpan) Warn(format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	logger.With("meta", fmt.Sprintf("%s:%d", srcFile, numLine)).WarnContext(span.Ctx, format, args...)
}

func (span *HybridSpan) Debug(format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	logger.With("meta", fmt.Sprintf("%s:%d", srcFile, numLine)).DebugContext(span.Ctx, format, args...)
}

func (span *HybridSpan) Error(format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	logger.With("meta", fmt.Sprintf("%s:%d", srcFile, numLine)).ErrorContext(span.Ctx, format, args...)
}

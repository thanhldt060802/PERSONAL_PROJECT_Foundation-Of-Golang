package otel

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

type LoggerConfig struct {
	ServiceName string
	Host        string
	Port        int
	LogFilePath string // Đường dẫn file log local
}

func NewLogger(config LoggerConfig) (*slog.Logger, func()) {
	ctx := context.Background()

	// 1. Tạo OTel log exporter (gửi lên collector)
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(fmt.Sprintf("%v:%v", config.Host, config.Port)),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create log exporter: %v", err))
	}

	// 2. Tạo resource (giống tracer)
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	// 3. Tạo LoggerProvider
	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(res),
	)

	// 4. Tạo file để ghi log local
	logFile, err := os.OpenFile(config.LogFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file: %v", err))
	}

	// 5. Tạo multi-writer (ghi cả file và stdout)
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// 6. Tạo OTel handler (gửi lên collector + có trace context)
	otelHandler := otelslog.NewHandler(config.ServiceName,
		otelslog.WithLoggerProvider(loggerProvider),
	)

	// 7. Tạo JSON handler cho file local
	fileHandler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// 8. Combine cả 2 handlers
	multiHandler := NewMultiHandler(otelHandler, fileHandler)

	logger := slog.New(multiHandler)

	// Cleanup function
	cleanup := func() {
		loggerProvider.Shutdown(ctx)
		logFile.Close()
	}

	return logger, cleanup
}

// MultiHandler để ghi log vào nhiều destinations
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

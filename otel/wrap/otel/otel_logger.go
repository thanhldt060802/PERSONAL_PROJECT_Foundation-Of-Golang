package otel

import (
	"context"
	"fmt"
	"io"
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

var (
	logger *slog.Logger
)

type LogLevel string

const (
	LOG_LEVEL_INFO  LogLevel = "info"
	LOG_LEVEL_WARN  LogLevel = "warn"
	LOG_LEVEL_DEBUG LogLevel = "debug"
	LOG_LEVEL_ERROR LogLevel = "error"
)

// INIT LOGGER

func initLogger(config *ObserverConfig) func(ctx context.Context) {
	exporter, err := otlploghttp.New(
		context.Background(),
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(config.EndPoint),
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
		if err := os.MkdirAll(filepath.Dir(config.LocalLogFile), 0755); err != nil {
			stdLog.Fatalf("Failed to create local log file dir for Logger: %v", err.Error())
		}
		file, err := os.OpenFile(config.LocalLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			stdLog.Fatalf("Failed to open local log file for Logger: %v", err.Error())
		}
		logFile = file

		multiWriter := io.MultiWriter(os.Stdout, file)

		fileHandlerOption := slog.HandlerOptions{}
		switch config.LocalLogLevel {
		case LOG_LEVEL_INFO:
			{
				fileHandlerOption.Level = slog.LevelInfo
			}
		case LOG_LEVEL_WARN:
			{
				fileHandlerOption.Level = slog.LevelWarn
			}
		case LOG_LEVEL_DEBUG:
			{
				fileHandlerOption.Level = slog.LevelDebug
			}
		case LOG_LEVEL_ERROR:
			{
				fileHandlerOption.Level = slog.LevelError
			}
		default:
			{
				fileHandlerOption.Level = slog.LevelInfo
			}
		}

		fileHandler := slog.NewJSONHandler(multiWriter, &fileHandlerOption)
		multiHandler = append(multiHandler, fileHandler)
	}

	logger = slog.New(newMultiHandler(multiHandler...))

	return func(ctx context.Context) {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			stdLog.Printf("Error occurred when shutting down Logger provider: %v", err)
		}
		if logFile != nil {
			logFile.Close()
		}
	}
}

// DEFINE MULTI HANDLER FOR LOGGER

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
		r := record.Clone()
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

// DEFINE LOGGER FEATURES

func InfoLog(ctx context.Context, format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	meta := fmt.Sprintf("%s:%d", srcFile, numLine)
	msg := fmt.Sprintf(format, args...)
	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		msg,
		slog.String("client_ip", getClientIpFromCtx(ctx)),
		slog.String("meta", meta),
	)
}

func WarnLog(ctx context.Context, format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	meta := fmt.Sprintf("%s:%d", srcFile, numLine)
	msg := fmt.Sprintf(format, args...)
	logger.LogAttrs(
		ctx,
		slog.LevelWarn,
		msg,
		slog.String("client_ip", getClientIpFromCtx(ctx)),
		slog.String("meta", meta),
	)
}

func DebugLog(ctx context.Context, format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	meta := fmt.Sprintf("%s:%d", srcFile, numLine)
	msg := fmt.Sprintf(format, args...)
	logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		msg,
		slog.String("client_ip", getClientIpFromCtx(ctx)),
		slog.String("meta", meta),
	)
}

func ErrorLog(ctx context.Context, format string, args ...any) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	meta := fmt.Sprintf("%s:%d", srcFile, numLine)
	msg := fmt.Sprintf(format, args...)
	logger.LogAttrs(
		ctx,
		slog.LevelError,
		msg,
		slog.String("client_ip", getClientIpFromCtx(ctx)),
		slog.String("meta", meta),
	)
}

func getClientIpFromCtx(ctx context.Context) string {
	ip, _ := ctx.Value(ClientIP).(string)
	return ip
}

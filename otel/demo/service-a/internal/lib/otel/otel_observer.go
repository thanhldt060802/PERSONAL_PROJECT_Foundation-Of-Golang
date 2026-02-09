package otel

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Observer manages lifecycle of all OpenTelemetry components.
type Observer struct {
	// Main feature

	tracer                 trace.Tracer            // Tracer instance for creating tracing spans
	logger                 *slog.Logger            // Logger instance for structured logging
	meter                  metric.Meter            // Meter instance for collecting metrics
	metricCollectorManager *metricCollectorManager // Metric collector manager for all registered metric

	// Other feature

	cache Cache // Cache for storing Trace Carriers (trace context)

	shutdowns []func(context.Context) // List of shutdown functions for cleanup
}

// Shutdown flushes all pending telemetry data and cleans up resources.
// It should be called before application exit.
func (o *Observer) Shutdown() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	for _, shutdown := range o.shutdowns {
		shutdown(shutdownCtx)
	}
}

// ObserverOption configures the Otel Observer during initialization.
type ObserverOption interface {
	apply(obsv *Observer)
}

// observerOptionFunc implements ObserverOption using a function.
type observerOptionFunc func(*Observer)

func (obsvOptFunc observerOptionFunc) apply(obsv *Observer) {
	obsvOptFunc(obsv)
}

// WithTracer enables distributed tracing with the given configuration.
// Returns nil if config is nil.
// This is mandatory if using Tracer; otherwise, it will no effect if Tracer is used without configuring it when initializing Otel Observer.
func WithTracer(config *TracerConfig) ObserverOption {
	return observerOptionFunc(func(o *Observer) {
		if config == nil {
			return
		}

		tracer, shutdown := initTracer(config)

		o.tracer = tracer
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

// WithLogger enables structured logging with OpenTelemetry integration.
// Logs are exported to OTLP endpoint and optionally written to local file.
// Returns nil if config is nil.
// This is mandatory if using Logger; otherwise, it will no effect if Logger is used without configuring it when initializing Otel Observer.
func WithLogger(config *LoggerConfig) ObserverOption {
	return observerOptionFunc(func(o *Observer) {
		if config == nil {
			return
		}

		logger, shutdown := initLogger(config)

		o.logger = logger
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

// WithMeter enables metrics collection and export.
// Supports Counter, UpDownCounter, Histogram, and Gauge metric types.
// Returns nil if config is nil.
// This is mandatory if using Meter; otherwise, it will no effect if Meter is used without configuring it when initializing Otel Observer.
func WithMeter(config *MeterConfig) ObserverOption {
	return observerOptionFunc(func(o *Observer) {
		if config == nil {
			return
		}

		if config.MetricCollectionInterval <= 0 {
			config.MetricCollectionInterval = defaultMeterInterval
		}

		meter, metricCollectorManager, shutdown := initMeter(config)

		o.meter = meter
		o.metricCollectorManager = metricCollectorManager
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

// WithRedisCache enables Redis-based trace context storage for async operations.
// Useful for propagating trace context across message queues or job systems.
// Returns nil if config is nil.
// This is mandatory if using Cache; otherwise, it will crash if Cache is used without configuring it when initializing Otel Observer.
func WithRedisCache(config *RedisConfig) ObserverOption {
	return observerOptionFunc(func(o *Observer) {
		if config == nil {
			return
		}

		if config.PoolSize <= 0 {
			config.PoolSize = defaultRedisPoolSize
		}
		if config.PoolTimeoutSec <= 0 {
			config.PoolTimeoutSec = defaultRedisPoolTimeoutSec
		}
		if config.IdleTimeoutSec <= 0 {
			config.IdleTimeoutSec = defaultRedisIdleTimeoutSec
		}
		if config.ReadTimeoutSec <= 0 {
			config.ReadTimeoutSec = defaultRedisReadTimeoutSec
		}
		if config.WriteTimeoutSec <= 0 {
			config.WriteTimeoutSec = defaultRedisWriteTimeoutSec
		}

		redisCache := initRedisCache(config)

		o.cache = redisCache
	})
}

// init sets some configs for OpenTelemetry.
func init() {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		stdLog.Printf("[error] Error occurred: %v", cause)
	}))
}

// NewOtelObserver initializes Otel Observer (OpenTelemetry Observer) with the given options.
// Returns a *Observer.
//
// Example:
//
//	observer := otel.NewOtelObserver(
//	    otel.WithTracer(&otel.TracerConfig{...}),
//	    otel.WithLogger(&otel.LoggerConfig{...}),
//	)
//	defer observer.shutdown()
func NewOtelObserver(opts ...ObserverOption) *Observer {
	obsv := &Observer{
		shutdowns: make([]func(context.Context), 0),
	}

	for _, opt := range opts {
		opt.apply(obsv)
	}

	if obsv.tracer == nil {
		obsv.tracer = otel.Tracer("default-tracer")
		stdLog.Printf("[warning] Tracer is unconfigured, using the default alternative Tracer")
	}

	return obsv
}

package otel

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
)

// observerOnce ensures Otel Observer is initialized only once.
var observerOnce sync.Once

// observer manages lifecycle of all OpenTelemetry components.
type observer struct {
	shutdowns []func(context.Context)
}

// ObserverOption configures the Otel Observer during initialization.
type ObserverOption interface {
	apply(obsv *observer)
}

// observerOptionFunc implements ObserverOption using a function.
type observerOptionFunc func(*observer)

func (obsvOptFunc observerOptionFunc) apply(obsv *observer) {
	obsvOptFunc(obsv)
}

// WithTracer enables distributed tracing with the given configuration.
// Returns nil if config is nil.
// This is mandatory if using Tracer; otherwise, it will no effect if Tracer is used without configuring it when initializing Otel Observer.
func WithTracer(config *TracerConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if config == nil {
			return
		}

		shutdown := initTracer(config)
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

// WithLogger enables structured logging with OpenTelemetry integration.
// Logs are exported to OTLP endpoint and optionally written to local file.
// Returns nil if config is nil.
// This is mandatory if using Logger; otherwise, it will no effect if Logger is used without configuring it when initializing Otel Observer.
func WithLogger(config *LoggerConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if config == nil {
			return
		}

		shutdown := initLogger(config)
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

// WithMeter enables metrics collection and export.
// Supports Counter, UpDownCounter, Histogram, and Gauge metric types.
// Returns nil if config is nil.
// This is mandatory if using Meter; otherwise, it will no effect if Meter is used without configuring it when initializing Otel Observer.
func WithMeter(config *MeterConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if config == nil {
			return
		}

		if config.MetricCollectionInterval <= 0 {
			config.MetricCollectionInterval = defaultMeterInterval
		}

		shutdown := initMeter(config)
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

// WithRedisCache enables Redis-based trace context storage for async operations.
// Useful for propagating trace context across message queues or job systems.
// Returns nil if config is nil.
// This is mandatory if using Cache; otherwise, it will crash if Cache is used without configuring it when initializing Otel Observer.
func WithRedisCache(config *RedisConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
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

		initRedisCache(config)
	})
}

// NewOtelObserver initializes Otel Observer (OpenTelemetry Observer) with the given options.
// It can only be called once (singleton pattern).
// Returns a shutdown function that must be called before application exit.
//
// Example:
//
//	shutdown := NewOtelObserver(
//	    WithTracer(&TracerConfig{...}),
//	    WithLogger(&LoggerConfig{...}),
//	)
//	defer shutdown()
func NewOtelObserver(opts ...ObserverOption) func() {
	var shutdown func()

	observerOnce.Do(func() {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
			stdLog.Printf("Error occurred: %v", cause)
		}))

		obsv := &observer{
			shutdowns: make([]func(context.Context), 0),
		}

		for _, opt := range opts {
			opt.apply(obsv)
		}

		shutdown = func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			for _, shutdown := range obsv.shutdowns {
				shutdown(shutdownCtx)
			}
		}
	})

	return shutdown
}

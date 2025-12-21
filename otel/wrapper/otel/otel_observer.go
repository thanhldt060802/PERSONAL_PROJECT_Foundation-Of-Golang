package otel

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
)

var (
	observerOnce sync.Once
)

type observer struct {
	shutdowns []func(context.Context)
}

type ObserverOption interface {
	apply(obsv *observer)
}

type observerOptionFunc func(*observer)

func (obsvOptFunc observerOptionFunc) apply(obsv *observer) {
	obsvOptFunc(obsv)
}

func WithTracer(cfg *TracerConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if cfg == nil {
			return
		}

		shutdown := initTracer(cfg)
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

func WithLogger(cfg *LoggerConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if cfg == nil {
			return
		}

		shutdown := initLogger(cfg)
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

func WithMeter(cfg *MeterConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if cfg == nil {
			return
		}

		if cfg.MetricCollectionInterval <= 0 {
			cfg.MetricCollectionInterval = defaultMeterInterval
		}

		shutdown := initMeter(cfg)
		o.shutdowns = append(o.shutdowns, shutdown)
	})
}

func WithRedisCache(cfg *RedisConfig) ObserverOption {
	return observerOptionFunc(func(o *observer) {
		if cfg == nil {
			return
		}

		if cfg.PoolSize <= 0 {
			cfg.PoolSize = defaultRedisPoolSize
		}
		if cfg.PoolTimeoutSec <= 0 {
			cfg.PoolTimeoutSec = defaultRedisPoolTimeoutSec
		}
		if cfg.IdleTimeoutSec <= 0 {
			cfg.IdleTimeoutSec = defaultRedisIdleTimeoutSec
		}
		if cfg.ReadTimeoutSec <= 0 {
			cfg.ReadTimeoutSec = defaultRedisReadTimeoutSec
		}
		if cfg.WriteTimeoutSec <= 0 {
			cfg.WriteTimeoutSec = defaultRedisWriteTimeoutSec
		}

		initRedisCache(cfg)
	})
}

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

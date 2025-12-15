package otel

import (
	"context"
	"time"
)

// SETUP OTEL OBSERVER

type ObserverConfig struct {
	ServiceName string
	EndPoint    string

	LocalLogFile  string
	LocalLogLevel LogLevel

	MetricCollectionInterval time.Duration
	metricDefs               []*MetricDef
}

func (config *ObserverConfig) AddMetricCollecter(metricDef *MetricDef) {
	config.metricDefs = append(config.metricDefs, metricDef)
}

func NewOtelObserver(config *ObserverConfig) func() {
	shutdownTracer := initTracer(config)
	shutdownLogger := initLogger(config)
	shutdownMeter := initMeter(config)

	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownTracer(shutdownCtx)
		shutdownLogger(shutdownCtx)
		shutdownMeter(shutdownCtx)
	}
}

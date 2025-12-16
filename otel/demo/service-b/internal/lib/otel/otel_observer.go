package otel

import (
	"context"
	"time"
)

// ObserverConfig holds the configuration for initializing OpenTelemetry observability components.
// It includes settings for tracer, logger, and meter.
type ObserverConfig struct {
	ServiceName    string // Name of the service
	ServiceVersion string // Version of the service
	EndPoint       string // OTLP endpoint for exporting telemetry data

	LocalLogFile  string   // Path to local log file (optional)
	LocalLogLevel LogLevel // Log level for local file logging

	MetricCollectionInterval time.Duration // Interval for collecting and exporting metrics
	metricDefs               []*MetricDef  // List of metric definitions to register
}

// AddMetricCollecter adds a metric definition to the configuration.
// Call this method before initializing the observer to register custom metrics.
//
// Parameters:
//   - metricDef: Definition of the metric to register
func (config *ObserverConfig) AddMetricCollecter(metricDef *MetricDef) {
	config.metricDefs = append(config.metricDefs, metricDef)
}

// NewOtelObserver initializes all OpenTelemetry components (tracer, logger, meter)
// and returns a cleanup function to gracefully shutdown all components.
//
// Parameters:
//   - config: Configuration for all observability components
//
// Returns:
//   - func(): A cleanup function that should be called when the application is shutting down
//
// Example:
//
//	shutdown := NewOtelObserver(config)
//	defer shutdown()
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

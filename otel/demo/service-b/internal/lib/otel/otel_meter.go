package otel

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

// meter is global Meter instance for metric collecting.
var meter metric.Meter

// mCollector is global metric manager for all registered metric.
var mCollector *metricCollector

// Default Meter settings.
const (
	// defaultMeterInterval is interval duration for metric collection.
	defaultMeterInterval = time.Millisecond * 10000
	// defaultGaugeMetricTTL is time to live for a gauge metric.
	defaultGaugeMetricTTL = time.Millisecond * 60000
)

// MetricName is a type-safe metric name identifier.
type MetricName string

// String returns the value string base.
func (mName MetricName) String() string {
	return string(mName)
}

// Get returns the metric name with prefix.
func (mName MetricName) Get() MetricName {
	return metricNamePrefix + mName
}

// Metric name prefix for Meter to avoid naming conflicts.
const metricNamePrefix MetricName = "custom_"

// MetricType defines the type of metric to collect.
type MetricType string

// Metric type definitions for Meter.
const (
	// METRIC_TYPE_COUNTER is used for creating a monotonically increasing counter.
	METRIC_TYPE_COUNTER MetricType = "counter"
	// METRIC_TYPE_UP_DOWN_COUNTER is used for creating a counter that can increase and decrease.
	METRIC_TYPE_UP_DOWN_COUNTER MetricType = "up-down-counter"
	// METRIC_TYPE_HISTOGRAM is used for creating a distribution of values collector.
	METRIC_TYPE_HISTOGRAM MetricType = "histogram"
	// METRIC_TYPE_GAUGE is used for creating a point-in-time value collector.
	METRIC_TYPE_GAUGE MetricType = "gauge"
)

// MeterConfig configures the metrics collection component
type MeterConfig struct {
	ServiceName    string            // Name of the service
	ServiceVersion string            // Version of the service
	EndPoint       string            // OTLP endpoint for exporting telemetry data
	Insecure       bool              // Allow HTTP schema, instead of HTTPS
	HttpHeader     map[string]string // Additional HTTP headers

	MetricCollectionInterval time.Duration // Interval for collecting and exporting metrics
	MetricDefs               []*MetricDef  // List of metric definitions to register
}

// initMeter initializes the global Meter and returns a cleanup function.
// Metrics are collected periodically and exported via OTLP HTTP.
func initMeter(config *MeterConfig) func(ctx context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(config.EndPoint),
	}
	if config.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(config.HttpHeader) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(config.HttpHeader))
	}

	// Create OTLP HTTP exporter for sending metrics
	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		stdLog.Fatalf("Failed to create exporter for Meter: %v", err)
	}

	// Create resource with service metadata
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.ServiceVersion),
		attribute.String("host.ip", getLocalIP()),
	)

	// Create Meter provider with periodic reader for automatic metric collection
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(config.MetricCollectionInterval))),
		sdkmetric.WithResource(resource),
	)

	otel.SetMeterProvider(meterProvider)

	// Init Meter
	meter = otel.Meter(config.ServiceName)
	mCollector = newMetricCollector()

	// Register all configured metrics
	for _, metricDef := range config.MetricDefs {
		switch metricDef.Type {
		case METRIC_TYPE_COUNTER:
			{
				if err := mCollector.registerCounter(metricDef); err != nil {
					stdLog.Fatalf("Failed to register Counter '%s' for Meter: %v", metricDef.Name, err)
				}
			}
		case METRIC_TYPE_UP_DOWN_COUNTER:
			{
				if err := mCollector.registerUpDownCounter(metricDef); err != nil {
					stdLog.Fatalf("Failed to register UpDownCounter '%s' for Meter: %v", metricDef.Name, err)
				}
			}
		case METRIC_TYPE_HISTOGRAM:
			{
				if err := mCollector.registerHistogram(metricDef); err != nil {
					stdLog.Fatalf("Failed to register Histogram '%s' for Meter: %v", metricDef.Name, err)
				}
			}
		case METRIC_TYPE_GAUGE:
			{
				if err := mCollector.registerGauge(metricDef); err != nil {
					stdLog.Fatalf("Failed to register Gauge '%s' for Meter: %v", metricDef.Name, err)
				}
			}
		default:
			{
				stdLog.Fatalf("Metric type '%s' is not valid", metricDef.Type)
			}
		}
	}

	// Return cleanup function
	return func(ctx context.Context) {
		if err := meterProvider.Shutdown(ctx); err != nil {
			stdLog.Printf("Error occurred when shutting down Meter provider: %v", err)
		}
	}
}

// metricCollector manages all registered metrics.
type metricCollector struct {
	counters       map[MetricName]metric.Int64Counter
	upDownCounters map[MetricName]metric.Int64UpDownCounter
	histograms     map[MetricName]metric.Float64Histogram
	gauges         map[MetricName]*observableGaugeState
}

// gaugeValue stores the current gauge value with metadata.
type gaugeValue struct {
	value     float64
	attrs     []attribute.KeyValue
	updatedAt time.Time
}

// observableGaugeState wraps an observable gauge with its current value.
type observableGaugeState struct {
	instrument  metric.Float64ObservableGauge
	currentVals map[string]*gaugeValue
	mu          sync.RWMutex
}

func newMetricCollector() *metricCollector {
	return &metricCollector{
		counters:       make(map[MetricName]metric.Int64Counter),
		upDownCounters: make(map[MetricName]metric.Int64UpDownCounter),
		histograms:     make(map[MetricName]metric.Float64Histogram),
		gauges:         make(map[MetricName]*observableGaugeState),
	}
}

// MetricDef defines a metric to be registered.
type MetricDef struct {
	Type        MetricType // Type of metric
	Name        MetricName // Name of metric
	Description string     // Description of metric
	Unit        string     // Unit of metric
}

// registerCounter creates and registers a counter metric.
func (mc *metricCollector) registerCounter(metricDef *MetricDef) error {
	if _, exists := mc.counters[metricDef.Name.Get()]; exists {
		return fmt.Errorf("counter '%s' already exists", metricDef.Name)
	}

	opts := []metric.Int64CounterOption{
		metric.WithDescription(metricDef.Description),
	}
	if metricDef.Unit != "" {
		opts = append(opts, metric.WithUnit(metricDef.Unit))
	}

	counter, err := meter.Int64Counter(metricDef.Name.Get().String(), opts...)
	if err != nil {
		return fmt.Errorf("failed to create counter '%s': %v", metricDef.Name, err)
	}

	mc.counters[metricDef.Name.Get()] = counter
	return nil
}

// registerUpDownCounter creates and registers an up-down counter metric.
func (mc *metricCollector) registerUpDownCounter(metricDef *MetricDef) error {
	if _, exists := mc.upDownCounters[metricDef.Name.Get()]; exists {
		return fmt.Errorf("updowncounter '%s' already exists", metricDef.Name)
	}

	opts := []metric.Int64UpDownCounterOption{
		metric.WithDescription(metricDef.Description),
	}
	if metricDef.Unit != "" {
		opts = append(opts, metric.WithUnit(metricDef.Unit))
	}

	updown, err := meter.Int64UpDownCounter(metricDef.Name.Get().String(), opts...)
	if err != nil {
		return fmt.Errorf("failed to create updowncounter '%s': %v", metricDef.Name, err)
	}

	mc.upDownCounters[metricDef.Name.Get()] = updown
	return nil
}

// registerHistogram creates and registers a histogram metric.
func (mc *metricCollector) registerHistogram(metricDef *MetricDef) error {
	if _, exists := mc.histograms[metricDef.Name.Get()]; exists {
		return fmt.Errorf("histogram '%s' already exists", metricDef.Name)
	}

	opts := []metric.Float64HistogramOption{
		metric.WithDescription(metricDef.Description),
	}
	if metricDef.Unit != "" {
		opts = append(opts, metric.WithUnit(metricDef.Unit))
	}

	histo, err := meter.Float64Histogram(metricDef.Name.Get().String(), opts...)
	if err != nil {
		return fmt.Errorf("failed to create histogram '%s': %v", metricDef.Name, err)
	}

	mc.histograms[metricDef.Name.Get()] = histo
	return nil
}

// registerGauge creates and registers a gauge metric with callback.
func (mc *metricCollector) registerGauge(metricDef *MetricDef) error {
	if _, exists := mc.gauges[metricDef.Name.Get()]; exists {
		return fmt.Errorf("gauge '%s' already exists", metricDef.Name)
	}

	opts := []metric.Float64ObservableGaugeOption{
		metric.WithDescription(metricDef.Description),
	}
	if metricDef.Unit != "" {
		opts = append(opts, metric.WithUnit(metricDef.Unit))
	}

	gauge, err := meter.Float64ObservableGauge(metricDef.Name.Get().String(), opts...)
	if err != nil {
		return fmt.Errorf("failed to create gauge '%s': %v", metricDef.Name, err)
	}

	gaugeState := &observableGaugeState{
		instrument:  gauge,
		currentVals: make(map[string]*gaugeValue),
	}

	// Register callback to observe gauge values during collection
	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			gaugeState.mu.RLock()
			defer gaugeState.mu.RUnlock()

			now := time.Now()

			for key, gaugeValue := range gaugeState.currentVals {
				if now.Sub(gaugeValue.updatedAt) > defaultGaugeMetricTTL {
					delete(gaugeState.currentVals, key)
				}
			}

			for _, gaugeValue := range gaugeState.currentVals {
				o.ObserveFloat64(gaugeState.instrument, gaugeValue.value,
					metric.WithAttributes(gaugeValue.attrs...),
				)
			}
			return nil
		},
		gauge,
	)
	if err != nil {
		return fmt.Errorf("failed to register gauge callback '%s': %v", metricDef.Name, err)
	}

	mc.gauges[metricDef.Name.Get()] = gaugeState
	return nil
}

// Context-aware metric recording functions.
// These functions extract trace_id and span_id from context automatically.

// RecordCounterWithCtx increments a counter by the given value.
// Counter values must be non-negative.
//
// Example:
//
//	otel.RecordCounterWithCtx(ctx, "requests", 1, map[string]any{"method": "GET"})
func RecordCounterWithCtx(ctx context.Context, name MetricName, value int64, metricAttrs map[string]any) {
	counter, ok := mCollector.counters[name.Get()]
	if !ok {
		stdLog.Printf("Counter '%s' not found", name)
		return
	}

	if value < 0 {
		stdLog.Printf("Value of Counter '%s' must be non-negative", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)
	counter.Add(ctx, value, metric.WithAttributes(attrs...))
}

// RecordUpDownCounterWithCtx adds the value to an up-down counter.
// Value can be positive (increment) or negative (decrement).
//
// Example:
//
//	otel.RecordUpDownCounterWithCtx(ctx, "connections", 1, map[string]any{"type": "websocket"})
//	otel.RecordUpDownCounterWithCtx(ctx, "connections", -1, map[string]any{"type": "websocket"})
func RecordUpDownCounterWithCtx(ctx context.Context, name MetricName, value int64, metricAttrs map[string]any) {
	upDownCounter, ok := mCollector.upDownCounters[name.Get()]
	if !ok {
		stdLog.Printf("UpDownCounter '%s' not found", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)
	upDownCounter.Add(ctx, value, metric.WithAttributes(attrs...))
}

// RecordHistogramWithCtx records a value in a histogram.
// Histograms aggregate value distributions (e.g., latency percentiles).
//
// Example:
//
//	otel.RecordHistogramWithCtx(ctx, "latency", 123.45, map[string]any{"endpoint": "/api/users"})
func RecordHistogramWithCtx(ctx context.Context, name MetricName, value float64, metricAttrs map[string]any) {
	histogram, ok := mCollector.histograms[name.Get()]
	if !ok {
		stdLog.Printf("Histogram '%s' not found", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)
	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

// Context-less metric recording functions.
// Use these when context is not available.

// RecordCounter increments a counter without trace context
func RecordCounter(name MetricName, value int64, metricAttrs map[string]any) {
	RecordCounterWithCtx(context.Background(), name, value, metricAttrs)
}

// RecordUpDownCounter updates an up-down counter without trace context
func RecordUpDownCounter(name MetricName, value int64, metricAttrs map[string]any) {
	RecordUpDownCounterWithCtx(context.Background(), name, value, metricAttrs)
}

// RecordHistogram records a histogram value without trace context
func RecordHistogram(name MetricName, value float64, metricAttrs map[string]any) {
	RecordHistogramWithCtx(context.Background(), name, value, metricAttrs)
}

// RecordGauge updates a gauge to the given value.
// Gauges represent current state (e.g., CPU usage, queue size).
//
// Example:
//
//	otel.RecordGauge("memory_usage", 75.5, map[string]any{"host": "server-1"})
func RecordGauge(name MetricName, value float64, metricAttrs map[string]any) {
	gaugeState, ok := mCollector.gauges[name.Get()]
	if !ok {
		stdLog.Printf("Gauge '%s' not found", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)
	key := hashAttrs(attrs)

	gaugeState.mu.Lock()
	defer gaugeState.mu.Unlock()

	// Update gauge value
	if _, ok := gaugeState.currentVals[key]; !ok {
		gaugeState.currentVals[key] = &gaugeValue{}
	}
	gaugeState.currentVals[key].value = value
	gaugeState.currentVals[key].attrs = attrs
	gaugeState.currentVals[key].updatedAt = time.Now()
}

func hashAttrs(attrs []attribute.KeyValue) string {
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].Key < attrs[j].Key
	})

	b := strings.Builder{}
	for _, a := range attrs {
		b.WriteString(string(a.Key))
		b.WriteString("=")
		b.WriteString(a.Value.Emit())
		b.WriteString("|")
	}
	return b.String()
}

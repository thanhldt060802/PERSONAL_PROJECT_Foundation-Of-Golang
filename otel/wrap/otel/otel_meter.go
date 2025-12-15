package otel

import (
	"context"
	"fmt"
	"math"
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

var (
	meter      metric.Meter
	mCollector *metricCollector
)

type MetricName string

func (mName MetricName) String() string {
	return string(mName)
}

func (mName MetricName) Get() MetricName {
	return METRIC_NAME_PREFIX + mName
}

const (
	METRIC_NAME_PREFIX MetricName = "custom_"
)

type MetricType string

const (
	METRIC_TYPE_COUNTER         MetricType = "counter"
	METRIC_TYPE_UP_DOWN_COUNTER MetricType = "up-down-counter"
	METRIC_TYPE_HISTOGRAM       MetricType = "histogram"
	METRIC_TYPE_GAUGE           MetricType = "gauge"
)

const (
	GaugeUpdateMinInterval    = 10 * time.Millisecond
	GaugeUpdateMinValueChange = 0.0001
)

// INIT METER

func initMeter(config *ObserverConfig) func(ctx context.Context) {
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(config.EndPoint),
	)
	if err != nil {
		stdLog.Fatalf("Failed to create exporter for Meter: %v", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
	)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(config.MetricCollectionInterval))),
		sdkmetric.WithResource(resource),
	)

	otel.SetMeterProvider(meterProvider)

	meter = otel.Meter(config.ServiceName)
	mCollector = newMetricCollector()

	for _, metricDef := range config.metricDefs {
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

	return func(ctx context.Context) {
		if err := meterProvider.Shutdown(ctx); err != nil {
			stdLog.Printf("Error occurred when shutting down Meter provider: %v", err)
		}
	}
}

// DEFINE METRIC COLLECTOR AND DEFINE METRIC DEFINITION FOR METRIC COLLECTOR

type metricCollector struct {
	counters       map[MetricName]metric.Int64Counter
	upDownCounters map[MetricName]metric.Int64UpDownCounter
	histograms     map[MetricName]metric.Float64Histogram
	gauges         map[MetricName]*observableGaugeState
}

type gaugeValue struct {
	value     float64
	attrs     []attribute.KeyValue
	updatedAt time.Time
}

type observableGaugeState struct {
	instrument metric.Float64ObservableGauge
	currentVal *gaugeValue
	mu         sync.RWMutex
}

func newMetricCollector() *metricCollector {
	return &metricCollector{
		counters:       make(map[MetricName]metric.Int64Counter),
		upDownCounters: make(map[MetricName]metric.Int64UpDownCounter),
		histograms:     make(map[MetricName]metric.Float64Histogram),
		gauges:         make(map[MetricName]*observableGaugeState),
	}
}

type (
	MetricDef struct {
		Type        MetricType
		Name        MetricName
		Description string
		Unit        string
	}
)

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
		instrument: gauge,
		currentVal: &gaugeValue{},
	}

	// Register callback for OTel reader pull
	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			gaugeState.mu.RLock()
			defer gaugeState.mu.RUnlock()

			o.ObserveFloat64(gaugeState.instrument, gaugeState.currentVal.value,
				metric.WithAttributes(gaugeState.currentVal.attrs...),
			)
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

// DEFINE METER FEATURES

func RecordCounter(ctx context.Context, name MetricName, value int64, metricAttrs map[string]any) {
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

func RecordUpDownCounter(ctx context.Context, name MetricName, value int64, metricAttrs map[string]any) {
	upDownCounter, ok := mCollector.upDownCounters[name.Get()]
	if !ok {
		stdLog.Printf("UpDownCounter '%s' not found", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)
	upDownCounter.Add(ctx, value, metric.WithAttributes(attrs...))
}

func RecordHistogram(ctx context.Context, name MetricName, value float64, metricAttrs map[string]any) {
	histogram, ok := mCollector.histograms[name.Get()]
	if !ok {
		stdLog.Printf("Histogram '%s' not found", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)
	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

func RecordGauge(name MetricName, value float64, metricAttrs map[string]any) {
	gaugeState, ok := mCollector.gauges[name.Get()]
	if !ok {
		stdLog.Printf("Gauge '%s' not found", name)
		return
	}

	attrs := mapToAttribute(metricAttrs)

	gaugeState.mu.Lock()
	defer gaugeState.mu.Unlock()

	prevVal := gaugeState.currentVal.value
	prevTime := gaugeState.currentVal.updatedAt

	if math.Abs(prevVal-value) < GaugeUpdateMinValueChange {
		return
	}

	if time.Since(prevTime) < GaugeUpdateMinInterval {
		return
	}

	gaugeState.currentVal.value = value
	gaugeState.currentVal.attrs = attrs
	gaugeState.currentVal.updatedAt = time.Now()
}

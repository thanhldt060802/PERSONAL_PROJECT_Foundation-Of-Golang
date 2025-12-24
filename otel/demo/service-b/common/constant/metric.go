package constant

import "thanhldt060802/internal/lib/otel"

var (
	// Counter
	HTTP_REQUESTS otel.MetricName = "http_requests"

	// UpDownCounter
	ACTIVE_JOBS otel.MetricName = "active_jobs"

	// Histogram
	JOB_PROCESS_DATA_SIZE otel.MetricName = "job_process_data_size"

	// Gauge
	CPU_USAGE otel.MetricName = "cpu_usage"
)

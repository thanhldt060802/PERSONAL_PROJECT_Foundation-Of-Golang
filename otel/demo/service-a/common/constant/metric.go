package constant

import "thanhldt060802/internal/lib/otel"

var (
	// Counter
	HTTP_REQUESTS_TOTAL otel.MetricName = "http_requests_total"

	// UpDownCounter
	ACTIVE_JOBS otel.MetricName = "active_jobs"

	// Histogram
	JOB_PROCESS_LATENCY_SEC otel.MetricName = "job_process_latency_sec"

	// Gauge
	CPU_USAGE_PERCENT otel.MetricName = "service_a_cpu_usage_percent"
)

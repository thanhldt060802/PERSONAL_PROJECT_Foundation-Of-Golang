package otel

import (
	"context"
	"log"
	"math"
	"net"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// stdLog is used for internal logging
var stdLog = log.New(os.Stdout, "[otel] ", log.LstdFlags)

// mapToAttribute converts a map to OpenTelemetry attributes.
// Supports common Go types: string, bool, int, int64, uint, uint64, float32, float64
// and their slice variants. Unsupported types are logged and skipped.
func mapToAttribute(attrMap map[string]any) []attribute.KeyValue {
	if len(attrMap) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(attrMap))

	for k, v := range attrMap {
		if v == nil {
			continue
		}

		switch val := v.(type) {

		// Scalar string type
		case string:
			{
				attrs = append(attrs, attribute.String(k, val))
			}

		// Boolean type
		case bool:
			{
				attrs = append(attrs, attribute.Bool(k, val))
			}

		// Integer types
		case int:
			{
				attrs = append(attrs, attribute.Int64(k, int64(val)))
			}
		case int64:
			{
				attrs = append(attrs, attribute.Int64(k, val))
			}
		case uint:
			{
				attrs = append(attrs, attribute.Int64(k, int64(val)))
			}
		case uint64:
			{
				// Only convert if within int64 range
				if val <= math.MaxInt64 {
					attrs = append(attrs, attribute.Int64(k, int64(val)))
				}
			}

		// Floating-point types
		case float32:
			{
				attrs = append(attrs, attribute.Float64(k, float64(val)))
			}
		case float64:
			{
				attrs = append(attrs, attribute.Float64(k, val))
			}

		// Slice types
		case []string:
			{
				attrs = append(attrs, attribute.StringSlice(k, val))
			}
		case []bool:
			{
				attrs = append(attrs, attribute.BoolSlice(k, val))
			}
		case []int:
			{
				// Convert []int to []int64
				convVal := make([]int64, len(val))
				for i := range val {
					convVal[i] = int64(val[i])
				}
				attrs = append(attrs, attribute.Int64Slice(k, convVal))
			}
		case []int64:
			{
				attrs = append(attrs, attribute.Int64Slice(k, val))
			}
		case []float64:
			{
				attrs = append(attrs, attribute.Float64Slice(k, val))
			}

		// Unsupported type
		default:
			stdLog.Printf("Pair[key:value] with value type is not allowed, key '%s' will be dropped", k)
		}
	}

	return attrs
}

// getTraceInfo extracts trace_id and span_id from context.
// Returns empty strings if context has no active span.
func getTraceInfo(ctx context.Context) (string, string) {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return "", ""
	}
	spanContext := span.SpanContext()
	return spanContext.TraceID().String(), spanContext.SpanID().String()
}

// getLocalIP returns the first non-loopback IPv4 address of the machine.
// Used to identify the host in telemetry data.
// Returns empty string if no suitable address is found.
func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				return ip.String()
			}
		}
	}

	return ""
}

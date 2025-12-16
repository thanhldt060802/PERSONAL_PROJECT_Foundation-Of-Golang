package otel

import (
	"log"
	"math"
	"os"

	"go.opentelemetry.io/otel/attribute"
)

// stdLog is the standard logger for internal logging within the otel package
var stdLog = log.New(os.Stdout, "[otel] ", log.LstdFlags)

// mapToAttribute converts a map of arbitrary values to OpenTelemetry attributes.
// It handles type conversion and validation for supported OpenTelemetry attribute types.
// Unsupported types are logged and dropped.
//
// Parameters:
//   - attrMap: Map of attribute keys and values
//
// Returns:
//   - []attribute.KeyValue: Slice of OpenTelemetry attributes
//
// Supported types:
//   - string, bool
//   - int, int64, uint, uint64
//   - float32, float64
//   - []string, []bool, []int, []int64, []float64
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

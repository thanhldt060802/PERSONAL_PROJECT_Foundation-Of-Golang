package otel

import (
	"log"
	"math"
	"os"

	"go.opentelemetry.io/otel/attribute"
)

var stdLog = log.New(os.Stdout, "[otel] ", log.LstdFlags)

// Accept for String, StringSlice, Bool, BoolSlice, Int, IntSlice, Int64, Int64Slice, Float64, Float64Slice.
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

		case string:
			{
				attrs = append(attrs, attribute.String(k, val))
			}
		case bool:
			{
				attrs = append(attrs, attribute.Bool(k, val))
			}
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
				if val <= math.MaxInt64 {
					attrs = append(attrs, attribute.Int64(k, int64(val)))
				}
			}

		case float32:
			{
				attrs = append(attrs, attribute.Float64(k, float64(val)))
			}
		case float64:
			{
				attrs = append(attrs, attribute.Float64(k, val))
			}

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

		default:
			stdLog.Printf("Pair[key:value] with value type is not allowed, key '%s' will be dropped", k)
		}
	}

	return attrs
}

package otel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// GinMiddlewares returns a slice of Gin middleware handlers for OpenTelemetry integration.
//
// Parameters:
//   - serviceName: The name of the service to be used in tracing spans
//
// Returns:
//   - []gin.HandlerFunc: A slice of middleware handlers to be registered with Gin router
func GinMiddlewares(serviceName string) []gin.HandlerFunc {
	mdws := []gin.HandlerFunc{}

	// Add middleware in order
	mdws = append(mdws, otelgin.Middleware(serviceName))

	return mdws
}

// HttpTransport creates and returns an HTTP transport wrapped with OpenTelemetry instrumentation.
// This transport automatically traces outgoing HTTP requests.
//
// Returns:
//   - *otelhttp.Transport: An HTTP transport with OpenTelemetry tracing enabled
func HttpTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(http.DefaultTransport)
}

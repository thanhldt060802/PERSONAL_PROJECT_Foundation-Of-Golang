package otel

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// GinMiddlewares returns a slice of Gin middleware handlers for OpenTelemetry integration.
// It includes middleware to inject client IP into context and OpenTelemetry tracing middleware.
//
// Parameters:
//   - serviceName: The name of the service to be used in tracing spans
//
// Returns:
//   - []gin.HandlerFunc: A slice of middleware handlers to be registered with Gin router
func GinMiddlewares(serviceName string) []gin.HandlerFunc {
	mdws := []gin.HandlerFunc{}

	// injectExtraInfoMdw injects the client IP address into the request context
	// so that downstream handlers can access this information
	injectExtraInfoMdw := func(c *gin.Context) {
		ctx := c.Request.Context()

		// Only inject client IP if it doesn't already exist in context
		if _, ok := ctx.Value(ClientIP).(string); !ok {
			ctx = context.WithValue(ctx, ClientIP, c.ClientIP())
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}

	// Add middleware in order: extra info injection first, then OpenTelemetry tracing
	mdws = append(mdws, injectExtraInfoMdw, otelgin.Middleware(serviceName))

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

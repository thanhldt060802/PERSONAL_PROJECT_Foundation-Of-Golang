package otel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// GinMiddlewares returns Gin middleware for automatic trace propagation.
// Adds tracing to all HTTP requests handled by Gin router.
//
// Example:
//
//	r := gin.New()
//	r.Use(otel.GinMiddlewares("api-service")...)
func GinMiddlewares(serviceName string) []gin.HandlerFunc {
	mdws := []gin.HandlerFunc{}

	// Add middleware in order
	mdws = append(mdws, otelgin.Middleware(serviceName))

	return mdws
}

// HttpTransport returns an HTTP transport with trace propagation.
// Use this with http.Client to propagate trace context in outbound requests.
//
// Example:
//
//	client := &http.Client{
//	    Transport: otel.HttpTransport(),
//	}
//	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
//	resp, _ := client.Do(req)
func HttpTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(http.DefaultTransport)
}

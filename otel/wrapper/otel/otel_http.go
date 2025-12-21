package otel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func GinMiddlewares(serviceName string) []gin.HandlerFunc {
	mdws := []gin.HandlerFunc{}

	// Add middleware in order
	mdws = append(mdws, otelgin.Middleware(serviceName))

	return mdws
}

func HttpTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(http.DefaultTransport)
}

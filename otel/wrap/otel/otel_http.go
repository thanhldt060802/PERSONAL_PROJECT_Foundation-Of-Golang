package otel

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type (
	clientIPKeyType struct{}
)

var (
	ClientIP = clientIPKeyType{}
)

func GinMiddleware(serviceName string) []gin.HandlerFunc {
	mdws := []gin.HandlerFunc{}

	injectExtraInfoMdw := func(c *gin.Context) {
		ctx := c.Request.Context()

		if _, ok := ctx.Value(ClientIP).(string); !ok {
			ctx = context.WithValue(ctx, ClientIP, c.ClientIP())
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}

	mdws = append(mdws, injectExtraInfoMdw, otelgin.Middleware(serviceName))

	return mdws
}

func HttpTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(http.DefaultTransport)
}

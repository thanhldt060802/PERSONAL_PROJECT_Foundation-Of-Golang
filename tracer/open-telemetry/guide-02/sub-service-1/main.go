package main

import (
	"fmt"
	"log"
	"net/http"
	"thanhldt060802/internal/opentelemetry"
	"thanhldt060802/middleware/auth"
	server "thanhldt060802/server/http"
	"thanhldt060802/service"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	apiV1 "thanhldt060802/api/v1"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read from config file failed: %v", err)
	}

	server.APP_NAME = viper.GetString("app.name")
	server.APP_VERSION = viper.GetString("app.version")
	server.APP_HOST = viper.GetString("app.host")
	server.APP_PORT = viper.GetInt("app.port")

	opentelemetry.ShutdownTracer = opentelemetry.NewTracer(opentelemetry.TracerEndPointConfig{
		ServiceName: server.APP_NAME,
		Host:        viper.GetString("jaeger.otlp_host"),
		Port:        viper.GetInt("jaeger.otlp_port"),
	})
}

func main() {
	defer opentelemetry.ShutdownTracer()

	router := server.NewHTTPServer()

	humaConfig := huma.Config{
		OpenAPI: &huma.OpenAPI{
			Components: &huma.Components{
				SecuritySchemes: map[string]*huma.SecurityScheme{
					"standard-auth": {
						Type:         "http",
						Scheme:       "bearer",
						In:           "header",
						Description:  "Authorization header using the Bearer scheme. Example: \"Authorization: Bearer {token}\"",
						BearerFormat: "Token String",
						Name:         "Authorization",
					},
				},
			},
			Servers: []*huma.Server{
				{
					URL:         fmt.Sprintf("http://%v:%v", server.APP_HOST, server.APP_PORT),
					Description: "Local Environment",
					Variables:   map[string]*huma.ServerVariable{},
				},
			},
		},
		OpenAPIPath:   fmt.Sprintf("/%v/openapi", server.APP_NAME),
		DocsPath:      "",
		Formats:       huma.DefaultFormats,
		DefaultFormat: "application/json",
	}

	router.GET(fmt.Sprintf("/%v/api-document", server.APP_NAME), func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<!doctype html>
		<html>
			<head>
				<title>MyService APIs</title>
				<meta charset="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
			</head>
			<body>
				<script id="api-reference" data-url="/`+server.APP_NAME+`/openapi.json"></script>
				<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
			</body>
		</html>
		`))
	})

	humaAPI := humagin.New(router, humaConfig)
	api := hureg.NewAPIGen(humaAPI)
	api = api.AddBasePath(fmt.Sprintf("%v/%v", server.APP_NAME, server.APP_VERSION[:2]))

	auth.AuthMdw = auth.NewSimpleAuthMiddleware()

	apiV1.RegisterAPIExample(api, service.NewPlayerService())

	server.Start(router)
}

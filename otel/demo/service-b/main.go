package main

import (
	"fmt"
	"net/http"
	"thanhldt060802/common/constant"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/lib/otel"
	"thanhldt060802/internal/redisclient"
	"thanhldt060802/internal/sqlclient"
	"thanhldt060802/middleware/auth"
	"thanhldt060802/model"
	"thanhldt060802/repository"
	"thanhldt060802/repository/db"
	server "thanhldt060802/server/http"
	"thanhldt060802/service"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	apiV1 "thanhldt060802/api/v1"
)

var shutdownObserver func()

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to load config/config.json: %v", err)
	}

	server.APP_NAME = viper.GetString("app.name")
	server.APP_VERSION = viper.GetString("app.version")
	server.APP_PORT = viper.GetInt("app.port")

	sqlclient.SqlClientConnInstance = sqlclient.NewSqlClient(sqlclient.SqlConfig{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetInt("db.port"),
		Database: viper.GetString("db.database"),
		Username: viper.GetString("db.username"),
		Password: viper.GetString("db.password"),
	})

	redisclient.RedisClientConnInstance = redisclient.NewRedisClient(redisclient.RedisConfig{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Database: viper.GetInt("redis.database"),
		Password: viper.GetString("redis.password"),
	})
	pubsub.RedisPubInstance = pubsub.NewRedisPub[*model.ExamplePubSubMessage](redisclient.RedisClientConnInstance.GetClient())

	shutdownObserver = otel.NewOtelObserver(
		otel.WithTracer(&otel.TracerConfig{
			ServiceName:    viper.GetString("app.name"),
			ServiceVersion: viper.GetString("app.version"),
			EndPoint:       viper.GetString("observer.tracer.end_point"),
			Insecure:       true,
			HttpHeader: map[string]string{
				"Authorization": "Bearer " + viper.GetString("observer.tracer.bearer_token"),
			},
		}),
		otel.WithLogger(&otel.LoggerConfig{
			ServiceName:    viper.GetString("app.name"),
			ServiceVersion: viper.GetString("app.version"),
			EndPoint:       viper.GetString("observer.logger.end_point"),
			Insecure:       true,
			HttpHeader: map[string]string{
				"Authorization": "Bearer " + viper.GetString("observer.logger.bearer_token"),
			},
			LocalLogFile:  viper.GetString("observer.logger.local_log_file"),
			LocalLogLevel: otel.LogLevel(viper.GetString("observer.logger.local_log_level")),
		}),
		otel.WithMeter(&otel.MeterConfig{
			ServiceName:    viper.GetString("app.name"),
			ServiceVersion: viper.GetString("app.version"),
			EndPoint:       viper.GetString("observer.meter.end_point"),
			Insecure:       true,
			HttpHeader: map[string]string{
				"Authorization": "Bearer " + viper.GetString("observer.meter.bearer_token"),
			},
			MetricCollectionInterval: time.Duration(viper.GetInt("observer.meter.metric_collection_interval_sec")) * time.Second,
			MetricDefs: []*otel.MetricDef{
				{
					Type:        otel.METRIC_TYPE_COUNTER,
					Name:        constant.HTTP_REQUESTS,
					Description: "Number of HTTP requests (count)",
					Unit:        "1",
				},
				{
					Type:        otel.METRIC_TYPE_UP_DOWN_COUNTER,
					Name:        constant.ACTIVE_JOBS,
					Description: "Current running job (count)",
					Unit:        "1",
				},
				{
					Type:        otel.METRIC_TYPE_HISTOGRAM,
					Name:        constant.JOB_PROCESS_DATA_SIZE,
					Description: "Job process data size (byte)",
					Unit:        "By",
				},
				{
					Type:        otel.METRIC_TYPE_GAUGE,
					Name:        constant.CPU_USAGE,
					Description: "CPU usage (%)",
					Unit:        "1",
				},
			},
		}),
	)
}

func main() {
	defer shutdownObserver()

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
					URL:         fmt.Sprintf("http://localhost:%v", server.APP_PORT),
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
				<title>B Service APIs</title>
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

	initRepository()

	apiV1.RegisterAPIExample(api, service.NewExampleService())

	startGaugeCollector()

	server.Start(router)
}

func initRepository() {
	repository.ExampleRepo = db.NewExampleRepo()
}

func startGaugeCollector() {
	service.StartGaugeCollector()
}

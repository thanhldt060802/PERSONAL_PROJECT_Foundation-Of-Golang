package main

import (
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/lib/otel"
	"thanhldt060802/internal/redisclient"
	"thanhldt060802/internal/sqlclient"
	"thanhldt060802/model"
	"thanhldt060802/repository"
	"thanhldt060802/repository/db"
	"thanhldt060802/service"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var shutdownObserver func()

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read from config file failed: %v", err)
	}

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
	pubsub.RedisSubInstance = pubsub.NewRedisSub[*model.ExamplePubSubMessage](redisclient.RedisClientConnInstance.GetClient())

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
			LocalLogLevel: otel.LogLevel(viper.GetString("observer.local_log_level")),
		}),
	)
}

func main() {
	defer shutdownObserver()

	initRepository()

	exampleService := service.NewExampleService()
	exampleService.InitSubscriber()

	log.Infof("Ready to consume message")
	select {}
}

func initRepository() {
	repository.ExampleRepo = db.NewExampleRepo()
}

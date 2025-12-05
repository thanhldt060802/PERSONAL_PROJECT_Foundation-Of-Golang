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

var ShutdownTracer func()

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

	ShutdownTracer = otel.NewTracer(otel.TracerEndPointConfig{
		ServiceName: viper.GetString("app.name"),
		Host:        viper.GetString("jaeger.otlp_host"),
		Port:        viper.GetInt("jaeger.otlp_port"),
	})

	initRepository()
}

func main() {
	defer ShutdownTracer()

	exampleService := service.NewExampleService()
	exampleService.InitSubscriber()

	log.Infof("Ready to consume message")
	select {}
}

func initRepository() {
	repository.ExampleRepo = db.NewExampleRepo()
}

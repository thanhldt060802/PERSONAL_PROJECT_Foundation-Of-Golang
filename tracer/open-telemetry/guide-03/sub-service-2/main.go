package main

import (
	"thanhldt060802/common/pubsub"
	"thanhldt060802/common/tracer"
	"thanhldt060802/internal/opentelemetry"
	"thanhldt060802/internal/redisclient"
	"thanhldt060802/internal/sqlclient"
	"thanhldt060802/repository"
	"thanhldt060802/repository/db"
	"thanhldt060802/service"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read from config file failed: %v", err)
	}

	appName := viper.GetString("app.name")

	switch viper.GetString("db.driver") {
	case "postgresql":
		{
			sqlclient.SqlClientConnInstance = sqlclient.NewSqlClient(sqlclient.SqlConfig{
				Host:     viper.GetString("db.host"),
				Port:     viper.GetInt("db.port"),
				Database: viper.GetString("db.database"),
				Username: viper.GetString("db.username"),
				Password: viper.GetString("db.password"),
			})
		}
	}

	opentelemetry.ShutdownTracer = opentelemetry.NewTracer(opentelemetry.TracerEndPointConfig{
		ServiceName: appName,
		Host:        viper.GetString("jaeger.otlp_host"),
		Port:        viper.GetInt("jaeger.otlp_port"),
	})

	redisclient.RedisClientConnInstance = redisclient.NewRedisClient(redisclient.RedisConfig{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Database: viper.GetInt("redis.database"),
		Password: viper.GetString("redis.password"),
	})
	pubsub.RedisSubInstance = pubsub.NewRedisSub[*tracer.MessageTracing](redisclient.RedisClientConnInstance.GetClient())

	initRepository()
}

func main() {
	defer opentelemetry.ShutdownTracer()

	playerService := service.NewPlayerService()
	playerService.InitSubscriber()

	select {}
}

func initRepository() {
	repository.PlayerRepo = db.NewPlayerRepo()
}

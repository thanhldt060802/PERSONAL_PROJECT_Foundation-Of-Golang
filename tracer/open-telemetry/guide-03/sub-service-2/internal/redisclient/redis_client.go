package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var RedisClientConnInstance IRedisClientConn

type IRedisClientConn interface {
	GetClient() *redis.Client
}

type RedisConfig struct {
	Host     string
	Port     int
	Database int
	Password string
}

type RedisClientConn struct {
	RedisConfig
	client *redis.Client
}

func NewRedisClient(config RedisConfig) IRedisClientConn {
	client := &RedisClientConn{}
	client.RedisConfig = config
	if err := client.Connect(); err != nil {
		log.Fatalf("Connect to redis failed: %v", err.Error())
	}

	return client
}

func (c *RedisClientConn) Connect() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", c.Host, c.Port),
		DB:       c.Database,
		Password: c.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return err
	}
	c.client = redisClient

	return nil
}

func (c *RedisClientConn) GetClient() *redis.Client {
	return c.client
}

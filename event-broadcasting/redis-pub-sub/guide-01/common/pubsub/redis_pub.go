package pubsub

import (
	"context"
	"encoding/json"
	"thanhldt060802/model"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var RedisPubInstance1 IRedisPub[string]
var RedisPubInstance2 IRedisPub[*model.DataStruct]

type IRedisPub[T any] interface {
	Publish(ctx context.Context, channel string, data T) error
}

type RedisPub[T any] struct {
	client *redis.Client
}

func NewRedisPub[T any](client *redis.Client) IRedisPub[T] {
	return &RedisPub[T]{
		client: client,
	}
}

func (redisPub *RedisPub[T]) Publish(ctx context.Context, channel string, data T) error {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Marshal data failed: %v", err.Error())
		return err
	}
	if err := redisPub.client.Publish(ctx, channel, payload).Err(); err != nil {
		log.Errorf("Publish %v to %v failed: %v", data, channel, err.Error())
		return err
	}

	log.Errorf("Publish %v to %v successful", data, channel)
	return nil
}

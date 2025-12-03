package pubsub

import (
	"context"
	"encoding/json"
	"reflect"
	"thanhldt060802/model"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var RedisSubInstance1 IRedisSub[string]
var RedisSubInstance2 IRedisSub[*model.DataStruct]

type IRedisSub[T any] interface {
	Subscribe(ctx context.Context, channel string, handler func(data T))
}

type RedisSub[T any] struct {
	client *redis.Client
}

func NewRedisSub[T any](client *redis.Client) IRedisSub[T] {
	return &RedisSub[T]{
		client: client,
	}
}

func (redisSub *RedisSub[T]) Subscribe(ctx context.Context, channel string, handler func(data T)) {
	sub := redisSub.client.Subscribe(ctx, channel)
	ch := sub.Channel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				sub.Close()
				return
			case message := <-ch:
				var value T
				t := reflect.TypeOf(value)

				var instance any
				if t.Kind() == reflect.Ptr {
					// T is pointer to struct: create *Struct
					instance = reflect.New(t.Elem()).Interface()
				} else {
					// T is value: create pointer to value (e.g., *int, *string)
					instance = reflect.New(t).Interface()
				}

				if err := json.Unmarshal([]byte(message.Payload), instance); err != nil {
					log.Errorf("Unmarshal %v failed: %v", message.Payload, err.Error())
					continue
				}

				var data T
				if t.Kind() == reflect.Ptr {
					// T is pointer already
					data = instance.(T)
				} else {
					// T is value, dereference pointer
					data = reflect.ValueOf(instance).Elem().Interface().(T)
				}

				handler(data)
			}
		}
	}()
}

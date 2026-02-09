package service

import (
	"context"
	"fmt"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal"
	"thanhldt060802/model"
	"thanhldt060802/repository"
)

type (
	IExampleService interface {
		InitSubscriber()
	}
	ExampleService struct {
	}
)

func NewExampleService() IExampleService {
	return &ExampleService{}
}

func (s *ExampleService) InitSubscriber() {
	pubsub.RedisSubInstance.Subscribe(context.Background(), "otel.pubsub.testing", func(message *model.ExamplePubSubMessage) {
		subCtx, span := internal.Observer.NewSpan(message.ExtractContext(), "SubscribeMessage")
		defer span.Done()

		span.AddEvent("Subscribe message from Redis", map[string]any{
			"redis.channel":              "otel.pubsub.testing",
			"redis.message.example_uuid": fmt.Sprintf("%v", message.ExampleUuid),
		})

		example, err := repository.ExampleRepo.GetById(subCtx, message.ExampleUuid)
		if err != nil {
			span.SetError(err)
			return
		}

		if example == nil {
			fmt.Println("Example example_uuid='" + message.ExampleUuid + "' not found")
		} else {
			fmt.Println(*example)
		}
	})
}

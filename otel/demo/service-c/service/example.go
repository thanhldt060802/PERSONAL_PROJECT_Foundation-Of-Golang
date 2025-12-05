package service

import (
	"context"
	"fmt"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/lib/otel"
	"thanhldt060802/model"
	"thanhldt060802/repository"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
		subCtx, span := otel.NewHybridSpan(message.ExtractContext())
		defer span.End()

		span.AddEvent("Subscribe message from Redis", trace.WithAttributes(
			attribute.String("redis.channel", "otel.pubsub.testing"),
			attribute.String("redis.message.example_uuid", fmt.Sprintf("%v", message.ExampleUuid)),
		))

		example, err := repository.ExampleRepo.GetById(subCtx, message.ExampleUuid)
		if err != nil {
			span.Error = err
			return
		}

		if example == nil {
			fmt.Println("Example example_uuid='" + message.ExampleUuid + "' not found")
		} else {
			fmt.Println(*example)
		}
	})
}

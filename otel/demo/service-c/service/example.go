package service

import (
	"context"
	"errors"
	"fmt"
	"thanhldt060802/common/observer"
	"thanhldt060802/common/pubsub"
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
	pubsub.RedisSubInstance.Subscribe(context.Background(), "observer.pubsub.testing", func(data *observer.MessageTracing) {
		subCtx, span := observer.StartSpanInternal(data.ExtractSpanContext())
		defer span.End()

		span.AddEvent("Subscribe message from Redis", trace.WithAttributes(
			attribute.String("redis.channel", "observer.pubsub.testing"),
			attribute.String("redis.message.payload", fmt.Sprintf("%v", data.Payload)),
		))

		exampleUuid, ok := data.Payload.(string)
		if !ok {
			span.Err = errors.New("invalid payload")
			return
		}

		example, err := repository.ExampleRepo.GetById(subCtx, exampleUuid)
		if err != nil {
			span.Err = err
			return
		}

		if example == nil {
			fmt.Println("Example example_uuid='" + exampleUuid + "' not found")
		} else {
			fmt.Println(*example)
		}
	})
}

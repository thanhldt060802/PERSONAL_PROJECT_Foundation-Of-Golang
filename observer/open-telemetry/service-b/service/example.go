package service

import (
	"context"
	"fmt"
	"thanhldt060802/common/apperror"
	"thanhldt060802/common/observer"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/model"
	"thanhldt060802/repository"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type (
	IExampleService interface {
		GetById(ctx context.Context, exampleUuid string) (*model.Example, error)
		PubSub_GetById(ctx context.Context, exampleUuid string) (string, error)
	}
	ExampleService struct {
	}
)

func NewExampleService() IExampleService {
	return &ExampleService{}
}

func (s *ExampleService) GetById(ctx context.Context, exampleUuid string) (*model.Example, error) {
	ctx, span := observer.StartSpanInternal(ctx)
	defer span.End()

	example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
	if err != nil {
		span.Err = err
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		return nil, apperror.ErrNotFound("Example example_uuid='"+exampleUuid+"' not found", "ERR_EXAMPLE_NOT_FOUND")
	}

	return example, nil
}

func (s *ExampleService) PubSub_GetById(ctx context.Context, exampleUuid string) (string, error) {
	ctx, span := observer.StartSpanInternal(ctx)
	defer span.End()

	msgTrace := observer.MessageTracing{
		SpanContext: propagation.MapCarrier{},
		Payload:     exampleUuid,
	}
	msgTrace.Inject(ctx)

	span.AddEvent("Publish message to Redis", trace.WithAttributes(
		attribute.String("redis.channel", "observer.pubsub.testing"),
		attribute.String("redis.message.payload", fmt.Sprintf("%v", msgTrace.Payload)),
	))

	if err := pubsub.RedisPubInstance.Publish(ctx, "observer.pubsub.testing", &msgTrace); err != nil {
		span.Err = err
		return "", apperror.ErrServiceUnavailable(err, "Failed to publish message to Redis")
	}

	return "success", nil
}

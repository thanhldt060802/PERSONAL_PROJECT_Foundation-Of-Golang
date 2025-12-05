package service

import (
	"context"
	"fmt"
	"thanhldt060802/common/apperror"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/lib/otel"
	"thanhldt060802/model"
	"thanhldt060802/repository"

	"go.opentelemetry.io/otel/attribute"
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
	ctx, span := otel.NewHybridSpan(ctx)
	defer span.End()

	example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
	if err != nil {
		span.Error = err
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		return nil, apperror.ErrNotFound("Example example_uuid='"+exampleUuid+"' not found", "ERR_EXAMPLE_NOT_FOUND")
	}

	return example, nil
}

func (s *ExampleService) PubSub_GetById(ctx context.Context, exampleUuid string) (string, error) {
	ctx, span := otel.NewHybridSpan(ctx)
	defer span.End()

	message := model.ExamplePubSubMessage{
		TraceCarrier: span.ExportTraceCarrier(),
		ExampleUuid:  exampleUuid,
	}

	span.AddEvent("Publish message to Redis", trace.WithAttributes(
		attribute.String("redis.channel", "otel.pubsub.testing"),
		attribute.String("redis.message.example_uuid", fmt.Sprintf("%v", message.ExampleUuid)),
	))

	if err := pubsub.RedisPubInstance.Publish(ctx, "otel.pubsub.testing", &message); err != nil {
		span.Error = err
		return "", apperror.ErrServiceUnavailable(err, "Failed to publish message to Redis")
	}

	return "success", nil
}

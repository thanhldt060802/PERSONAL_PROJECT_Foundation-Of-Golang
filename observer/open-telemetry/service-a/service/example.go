package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		CrossService_GetById(ctx context.Context, exampleUuid string) (*model.Example, error)
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
		return nil, err
	}

	return example, nil
}

func (s *ExampleService) CrossService_GetById(ctx context.Context, exampleUuid string) (*model.Example, error) {
	url := fmt.Sprintf("http://localhost:8002/service-b/v1/example/%v", exampleUuid)
	ctx, span, req, err := observer.StartSpanCrossService(ctx, "GET", url)
	if err != nil {
		return nil, err
	}
	span.End()

	span.AddEvent("Request HTTP to service-a", trace.WithAttributes(
		attribute.String("url", url),
	))

	client := http.Client{}
	req.Header.Set("Authorization", ctx.Value("auth_header").(string))

	res, err := client.Do(req)
	if err != nil {
		span.Err = err
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		span.Err = errors.New("request not OK")
		return nil, span.Err
	}

	resWrapper := new(struct {
		Data model.Example
	})
	if err := json.NewDecoder(res.Body).Decode(resWrapper); err != nil {
		span.Err = err
		return nil, err
	}
	example := &resWrapper.Data

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
		return "", err
	}

	return "success", nil
}

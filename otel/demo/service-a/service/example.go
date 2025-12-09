package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		CrossService_GetById(ctx context.Context, exampleUuid string) (*model.Example, error)
		PubSub_GetById(ctx context.Context, exampleUuid string) (string, error)
		Hybrid_GetById(ctx context.Context, exampleUuid string) (string, error)
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
		span.Err = err
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		return nil, apperror.ErrNotFound("Example example_uuid='"+exampleUuid+"' not found", "ERR_EXAMPLE_NOT_FOUND")
	}
	return example, nil
}

func (s *ExampleService) CrossService_GetById(ctx context.Context, exampleUuid string) (*model.Example, error) {
	ctx, span := otel.NewHybridSpan(ctx)
	defer span.End()

	url := fmt.Sprintf("http://localhost:8002/service-b/v1/example/%v", exampleUuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.Err = err
		return nil, apperror.ErrServiceUnavailable(err, "Failed to init cross-service")
	}
	span.InjectToRequestHeader(req.Header)

	span.AddEvent("Request HTTP to service-b", trace.WithAttributes(
		attribute.String("url", url),
	))

	client := http.Client{}
	req.Header.Set("Authorization", ctx.Value("auth_header").(string))

	res, err := client.Do(req)
	if err != nil {
		span.Err = err
		return nil, apperror.ErrServiceUnavailable(err, "Failed to request to service-b")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		span.Err = errors.New("response is not OK")
		return nil, apperror.ErrServiceUnavailable(span.Err, "Response is not OK from service-b")
	}

	resWrapper := new(struct {
		Data model.Example
	})
	if err := json.NewDecoder(res.Body).Decode(resWrapper); err != nil {
		span.Err = err
		return nil, apperror.ErrServiceUnavailable(err, "Failed to decode response from service-b")
	}
	example := &resWrapper.Data

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
		span.Err = err
		return "", apperror.ErrServiceUnavailable(err, "Failed to publish message to Redis")
	}

	return "success", nil
}

func (s *ExampleService) Hybrid_GetById(ctx context.Context, exampleUuid string) (string, error) {
	ctx, span := otel.NewHybridSpan(ctx)
	defer span.End()

	url := fmt.Sprintf("http://localhost:8002/service-b/v1/example/%v/pub-sub", exampleUuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.Err = err
		return "", apperror.ErrServiceUnavailable(err, "Failed to init cross-service")
	}
	span.InjectToRequestHeader(req.Header)

	span.AddEvent("Request HTTP to service-b", trace.WithAttributes(
		attribute.String("url", url),
	))

	client := http.Client{}
	req.Header.Set("Authorization", ctx.Value("auth_header").(string))

	res, err := client.Do(req)
	if err != nil {
		span.Err = err
		return "", apperror.ErrServiceUnavailable(err, "Failed to request to service-b")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		span.Err = errors.New("response is not OK")
		return "", apperror.ErrServiceUnavailable(span.Err, "Response is not OK from service-b")
	}

	resWrapper := new(struct {
		Data string
	})
	if err := json.NewDecoder(res.Body).Decode(resWrapper); err != nil {
		span.Err = err
		return "", apperror.ErrServiceUnavailable(err, "Failed to decode response from service-b")
	}
	result := resWrapper.Data

	return result, nil
}

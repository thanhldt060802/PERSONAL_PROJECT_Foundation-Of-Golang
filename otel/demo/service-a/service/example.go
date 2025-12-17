package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"thanhldt060802/common/apperror"
	"thanhldt060802/common/constant"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal/lib/otel"
	"thanhldt060802/model"
	"thanhldt060802/repository"
	"time"
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
	ctx, span := otel.NewSpan(ctx, "GetExampleById-Service")
	defer span.Done()

	otel.InfoLog(ctx, "[Service layer] Get Example by example_uuid='%s'", exampleUuid)

	otel.RecordCounter(ctx, constant.HTTP_REQUESTS, 1, nil)

	if rand.IntN(3) == 0 {
		err := errors.New("simulate error")
		otel.ErrorLog(ctx, "[Service layer] Failed to get Example by example_uuid='%s'", exampleUuid)
		span.SetError(err)
		return nil, apperror.ErrInternalServerError(err, "Failed to preprocess", "ERR_PREPROCESS")
	}

	go func(ctx context.Context) {
		ctx, span := otel.NewSpan(ctx, "AsyncJob")
		defer span.Done()

		otel.RecordUpDownCounter(span.Context(), constant.ACTIVE_JOBS, 1, nil)
		otel.InfoLog(ctx, "[Async job] Start process job")

		N := 3 + rand.IntN(3)
		for i := 0; i < N; i++ {
			time.Sleep(time.Duration(10+rand.IntN(10)) * time.Second)
			otel.RecordHistogram(ctx, constant.JOB_PROCESS_DATA_SIZE, float64(100+rand.IntN(100)), nil)
		}

		otel.RecordUpDownCounter(ctx, constant.ACTIVE_JOBS, -1, nil)
		otel.InfoLog(ctx, "[Async job] End process job")
	}(ctx)

	example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
	if err != nil {
		otel.ErrorLog(ctx, "[Service layer] Failed to get Example by example_uuid='%s': %v", exampleUuid, err)
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		otel.ErrorLog(ctx, "[Service layer] Failed to get Example by example_uuid='%s': Example not found", exampleUuid)
		return nil, apperror.ErrNotFound("Example example_uuid='"+exampleUuid+"' not found", "ERR_EXAMPLE_NOT_FOUND")
	}
	return example, nil
}

func (s *ExampleService) CrossService_GetById(ctx context.Context, exampleUuid string) (*model.Example, error) {
	ctx, span := otel.NewSpan(ctx, "CrossService_GetExampleById-Service")
	defer span.Done()

	url := fmt.Sprintf("http://localhost:8002/service-b/v1/example/%v", exampleUuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.SetError(err)
		return nil, apperror.ErrServiceUnavailable(err, "Failed to init cross-service")
	}
	req.Header.Set("Authorization", ctx.Value("auth_header").(string))

	span.AddEvent("Request HTTP to service-b", map[string]any{
		"url": url,
	})

	// span.InjectToRequestHeader(req.Header)
	client := http.Client{
		Transport: otel.HttpTransport(),
	}

	res, err := client.Do(req)
	if err != nil {
		span.SetError(err)
		return nil, apperror.ErrServiceUnavailable(err, "Failed to request to service-b")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := errors.New("response is not OK")
		span.SetError(err)
		return nil, apperror.ErrServiceUnavailable(err, "Response is not OK from service-b")
	}

	resWrapper := new(struct {
		Data model.Example
	})
	if err := json.NewDecoder(res.Body).Decode(resWrapper); err != nil {
		span.SetError(err)
		return nil, apperror.ErrServiceUnavailable(err, "Failed to decode response from service-b")
	}
	example := &resWrapper.Data

	return example, nil
}

func (s *ExampleService) PubSub_GetById(ctx context.Context, exampleUuid string) (string, error) {
	ctx, span := otel.NewSpan(ctx, "PubSub_GetExampleById-Service")
	defer span.Done()

	message := model.ExamplePubSubMessage{
		TraceCarrier: otel.ExportTraceCarrier(ctx),
		ExampleUuid:  exampleUuid,
	}

	span.AddEvent("Publish message to Redis", map[string]any{
		"redis.channel":              "otel.pubsub.testing",
		"redis.message.example_uuid": fmt.Sprintf("%v", message.ExampleUuid),
	})

	if err := pubsub.RedisPubInstance.Publish(ctx, "otel.pubsub.testing", &message); err != nil {
		span.SetError(err)
		return "", apperror.ErrServiceUnavailable(err, "Failed to publish message to Redis")
	}

	return "success", nil
}

func (s *ExampleService) Hybrid_GetById(ctx context.Context, exampleUuid string) (string, error) {
	ctx, span := otel.NewSpan(ctx, "Hybrid_GetExampleById-Service")
	defer span.Done()

	url := fmt.Sprintf("http://localhost:8002/service-b/v1/example/%v/pub-sub", exampleUuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.SetError(err)
		return "", apperror.ErrServiceUnavailable(err, "Failed to init cross-service")
	}
	req.Header.Set("Authorization", ctx.Value("auth_header").(string))

	span.AddEvent("Request HTTP to service-b", map[string]any{
		"url": url,
	})

	// span.InjectToRequestHeader(req.Header)
	client := http.Client{
		Transport: otel.HttpTransport(),
	}

	res, err := client.Do(req)
	if err != nil {
		span.SetError(err)
		return "", apperror.ErrServiceUnavailable(err, "Failed to request to service-b")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := errors.New("response is not OK")
		span.SetError(err)
		return "", apperror.ErrServiceUnavailable(err, "Response is not OK from service-b")
	}

	resWrapper := new(struct {
		Data string
	})
	if err := json.NewDecoder(res.Body).Decode(resWrapper); err != nil {
		span.SetError(err)
		return "", apperror.ErrServiceUnavailable(err, "Failed to decode response from service-b")
	}
	result := resWrapper.Data

	return result, nil
}

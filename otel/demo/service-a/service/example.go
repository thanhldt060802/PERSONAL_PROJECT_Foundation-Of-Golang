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
		BulkAsync_GetById(ctx context.Context, exampleUuid string) (string, error)
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

	otel.InfoLogWithCtx(ctx, "[Service layer] Get Example by example_uuid='%s'", exampleUuid)

	otel.RecordCounterWithCtx(ctx, constant.HTTP_REQUESTS, 1, nil)

	if rand.IntN(3) == 0 {
		err := errors.New("simulate error")
		otel.ErrorLogWithCtx(ctx, "[Service layer] Failed to get Example by example_uuid='%s'", exampleUuid)
		span.SetError(err)
		return nil, apperror.ErrInternalServerError(err, "Failed to preprocess", "ERR_PREPROCESS")
	}

	go func(ctx context.Context) {
		ctx, span := otel.NewSpan(ctx, "AsyncJob")
		defer span.Done()

		otel.RecordUpDownCounterWithCtx(span.Context(), constant.ACTIVE_JOBS, 1, nil)
		otel.InfoLogWithCtx(ctx, "[Async job] Start process job")

		N := 3 + rand.IntN(3)
		for i := 0; i < N; i++ {
			time.Sleep(time.Duration(3+rand.IntN(3)) * time.Second)
			otel.RecordHistogramWithCtx(ctx, constant.JOB_PROCESS_DATA_SIZE, rand.Float64()*float64(rand.IntN(10000)), nil)
		}

		otel.RecordUpDownCounterWithCtx(ctx, constant.ACTIVE_JOBS, -1, nil)
		otel.InfoLogWithCtx(ctx, "[Async job] End process job")
	}(ctx)

	example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
	if err != nil {
		otel.ErrorLogWithCtx(ctx, "[Service layer] Failed to get Example by example_uuid='%s': %v", exampleUuid, err)
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		otel.ErrorLogWithCtx(ctx, "[Service layer] Failed to get Example by example_uuid='%s': Example not found", exampleUuid)
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

func (s *ExampleService) BulkAsync_GetById(ctx context.Context, exampleUuid string) (string, error) {
	for i := 1; i <= 5; i++ {
		ctx, span := otel.NewSpan(context.Background(), "BulkAsync_GetExampleById-Service")
		defer span.Done()

		key := fmt.Sprintf("%s-%d", exampleUuid, i)
		if err := otel.SetCacheTraceCarrierFromGroup("my-job", key, otel.ExportTraceCarrier(ctx)); err != nil {
			otel.ErrorLogWithCtx(ctx, "Failed to set cache trace carrier: %v", err)
		}

		go func(exampleUuid string, count int) {
			ctx, span := otel.NewSpan(context.Background(), "BulkAsync_GetExampleById-Worker")

			key := fmt.Sprintf("%s-%d", exampleUuid, i)
			traceCarrier, err := otel.GetCacheTraceCarrierFromGroup("my-job", key)
			if err != nil {
				otel.ErrorLogWithCtx(ctx, "Failed to get cache trace carrier: %v", err)
			} else {
				ctx, span = otel.NewSpan(traceCarrier.ExtractContext(), "BulkAsync_GetExampleById-Worker")
			}

			defer span.Done()

			time.Sleep(5 * time.Second)

			example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
			if err != nil {
				span.SetError(err)
				return
			}

			if example == nil {
				fmt.Println("Example example_uuid='" + exampleUuid + "' not found")
			} else {
				fmt.Println(*example)
			}

			if err := otel.DeleteCacheTraceCarrierFromGroup("my-job", key); err != nil {
				otel.ErrorLogWithCtx(ctx, "Failed to delete cache trace carrier: %v", err)
			}
		}(exampleUuid, i)
	}

	return "success", nil
}

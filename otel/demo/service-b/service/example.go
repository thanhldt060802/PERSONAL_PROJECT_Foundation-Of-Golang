package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"thanhldt060802/common/apperror"
	"thanhldt060802/common/constant"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/internal"
	"thanhldt060802/internal/lib/otel"
	"thanhldt060802/model"
	"thanhldt060802/repository"
	"time"
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
	ctx, span := internal.Observer.NewSpan(ctx, "GetExampleById-Service")
	defer span.Done()

	internal.Observer.InfoLogWithCtx(ctx, "[Service layer] Get Example by example_uuid='%s'", exampleUuid)

	internal.Observer.RecordCounterWithCtx(ctx, constant.HTTP_REQUESTS, 1, nil)

	if rand.IntN(3) == 0 {
		err := errors.New("simulate error")
		internal.Observer.ErrorLogWithCtx(ctx, "[Service layer] Failed to get Example by example_uuid='%s'", exampleUuid)
		span.SetError(err)
		return nil, apperror.ErrInternalServerError(err, "Failed to preprocess", "ERR_PREPROCESS")
	}

	go func(ctx context.Context) {
		ctx, span := internal.Observer.NewSpan(ctx, "AsyncJob")
		defer span.Done()

		internal.Observer.RecordUpDownCounterWithCtx(span.Context(), constant.ACTIVE_JOBS, 1, nil)
		internal.Observer.InfoLogWithCtx(ctx, "[Async job] Start process job")

		N := 3 + rand.IntN(3)
		for i := 0; i < N; i++ {
			time.Sleep(time.Duration(3+rand.IntN(3)) * time.Second)
			internal.Observer.RecordHistogramWithCtx(ctx, constant.JOB_PROCESS_DATA_SIZE, rand.Float64()*float64(rand.IntN(10000)), nil)
		}

		internal.Observer.RecordUpDownCounterWithCtx(ctx, constant.ACTIVE_JOBS, -1, nil)
		internal.Observer.InfoLogWithCtx(ctx, "[Async job] End process job")
	}(ctx)

	example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
	if err != nil {
		internal.Observer.ErrorLogWithCtx(ctx, "[Service layer] Failed to get Example by example_uuid='%s': %v", exampleUuid, err)
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		internal.Observer.ErrorLogWithCtx(ctx, "[Service layer] Failed to get Example by example_uuid='%s': Example not found", exampleUuid)
		return nil, apperror.ErrNotFound("Example example_uuid='"+exampleUuid+"' not found", "ERR_EXAMPLE_NOT_FOUND")
	}
	return example, nil
}

func (s *ExampleService) PubSub_GetById(ctx context.Context, exampleUuid string) (string, error) {
	ctx, span := internal.Observer.NewSpan(ctx, "PubSub_GetExampleById-Service")
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

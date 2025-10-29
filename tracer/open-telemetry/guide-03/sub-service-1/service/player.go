package service

import (
	"context"
	"encoding/json"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/common/tracer"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type (
	IPlayerService interface {
		GetById(ctx context.Context, playUuid string) (string, error)
	}
	PlayerService struct {
	}
)

func NewPlayerService() IPlayerService {
	return &PlayerService{}
}

func (s *PlayerService) GetById(ctx context.Context, playUuid string) (string, error) {
	ctx, span := tracer.StartSpanInternal(ctx)
	defer span.End()

	msgTrace := tracer.MessageTracing{
		TraceContext: propagation.MapCarrier{},
		Payload:      playUuid,
	}

	otel.GetTextMapPropagator().Inject(ctx, msgTrace.TraceContext)

	payloadBytes, _ := json.Marshal(msgTrace.Payload)
	span.AddEvent("Publish to Redis", trace.WithAttributes(
		attribute.String("redis.pub_channel", "test.trace.pubsub"),
		attribute.String("test.trace.pubsub.payload", string(payloadBytes)),
	))
	if err := pubsub.RedisPubInstance.Publish(ctx, "test.trace.pubsub", &msgTrace); err != nil {
		span.Err = err
		return "", err
	}

	return "success", nil
}

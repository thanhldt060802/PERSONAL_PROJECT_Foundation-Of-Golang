package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"thanhldt060802/common/pubsub"
	"thanhldt060802/common/tracer"
	"thanhldt060802/repository"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	IPlayerService interface {
		InitSubscriber()
	}
	PlayerService struct {
	}
)

func NewPlayerService() IPlayerService {
	return &PlayerService{}
}

func (s *PlayerService) InitSubscriber() {
	pubsub.RedisSubInstance.Subscribe(context.Background(), "test.trace.pubsub", func(data *tracer.MessageTracing) {
		pubCtx := data.ExtractTraceContext()
		subCtx, span := tracer.StartSpanInternal(pubCtx)
		defer span.End()

		payloadBytes, _ := json.Marshal(data.Payload)
		span.AddEvent("Subscribe on Redis", trace.WithAttributes(
			attribute.String("redis.sub_channel", "test.trace.pubsub"),
			attribute.String("test.trace.pubsub.payload", string(payloadBytes)),
		))

		playerUuid, ok := data.Payload.(string)
		if !ok {
			span.Err = errors.New("invalid payload")
			return
		}

		player, err := repository.PlayerRepo.GetById(subCtx, playerUuid)
		if err != nil {
			span.Err = err
			return
		}
		fmt.Println(*player)
	})
}

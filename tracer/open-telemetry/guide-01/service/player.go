package service

import (
	"context"
	"thanhldt060802/common/tracer"
	"thanhldt060802/model"
	"thanhldt060802/repository"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	IPlayerService interface {
		GetById(ctx context.Context, playUuid string) (*model.Player, error)
	}
	PlayerService struct {
	}
)

func NewPlayerService() IPlayerService {
	return &PlayerService{}
}

func (s *PlayerService) GetById(ctx context.Context, playUuid string) (*model.Player, error) {
	ctx, span := tracer.StartSpanInternal(ctx)
	defer span.End()

	span.AddEvent("Call to PlayerRepo.GetById")
	time.Sleep(1 * time.Second)

	player, err := repository.PlayerRepo.GetById(ctx, playUuid)
	if err != nil {
		span.Err = err
		return nil, err
	}

	span.AddEvent("Call to Other.ActionName", trace.WithAttributes(
		attribute.String("param.a", "Data A"),
		attribute.String("param.b", "Data B"),
	))
	time.Sleep(1 * time.Second)

	return player, nil
}

package service

import (
	"context"
	"thanhldt060802/common/tracer"
	"thanhldt060802/model"
	"thanhldt060802/repository"
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

	player, err := repository.PlayerRepo.GetById(ctx, playUuid)
	if err != nil {
		span.Err = err
		return nil, err
	}

	return player, nil
}

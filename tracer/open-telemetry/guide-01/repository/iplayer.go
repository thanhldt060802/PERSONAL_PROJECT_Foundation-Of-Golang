package repository

import (
	"context"
	"thanhldt060802/model"
)

type IPlayerRepo interface {
	GetById(ctx context.Context, playerUuid string) (*model.Player, error)
}

var PlayerRepo IPlayerRepo

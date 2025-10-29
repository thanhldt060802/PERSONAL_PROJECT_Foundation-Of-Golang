package repository

import (
	"context"
	"thanhldt060802/dtos"
	"thanhldt060802/model"

	"github.com/google/uuid"
)

type ITaskRepo interface {
	GetsView(ctx context.Context, filter *dtos.GetsTaskFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error)
	GetsViewCustom(ctx context.Context, filter *dtos.GetsTaskCustomFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error)
	GetViewById(ctx context.Context, id uuid.UUID) (*model.TaskView, error)
	GetById(ctx context.Context, id uuid.UUID) (*model.Task, error)
	Create(ctx context.Context, feature *model.Task) error
	UpdateById(ctx context.Context, id uuid.UUID, feature *model.Task) error
	PatchById(ctx context.Context, id uuid.UUID, feature *model.Task) error
	DeleteById(ctx context.Context, id uuid.UUID) error
}

var TaskRepo ITaskRepo

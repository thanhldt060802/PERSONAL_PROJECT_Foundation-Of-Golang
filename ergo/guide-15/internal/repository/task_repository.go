package repository

import (
	"context"
	"thanhldt060802/infrastructure"
	"thanhldt060802/internal/model"
)

type taskRepository struct {
}

type TaskRepository interface {
	GetById(ctx context.Context, id int64) (*model.Task, error)
	Update(ctx context.Context, updatedTask *model.Task) error
}

func NewTaskRepository() TaskRepository {
	return &taskRepository{}
}

func (taskRepository *taskRepository) GetById(ctx context.Context, id int64) (*model.Task, error) {
	var task model.Task

	if err := infrastructure.PostgresDB.NewSelect().Model(&task).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}

	return &task, nil
}

func (taskRepository *taskRepository) Update(ctx context.Context, updatedTask *model.Task) error {
	_, err := infrastructure.PostgresDB.NewUpdate().Model(updatedTask).Where("id = ?", updatedTask.Id).Exec(ctx)
	return err
}

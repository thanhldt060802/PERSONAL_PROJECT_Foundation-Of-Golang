package repository

import (
	"math/rand"
	"thanhldt060802/internal/model"
)

type taskRepository struct {
}

type TaskRepository interface {
	GetById(id int64) *model.Task
}

func NewTaskRepository() TaskRepository {
	return &taskRepository{}
}

func (taskRepository *taskRepository) GetById(id int64) *model.Task {
	return &model.Task{
		Id:       id,
		Progress: 0,
		Target:   rand.Intn(20) + 1,
	}
}

package service

import (
	"context"

	"thanhldt060802/common/apperror"
	"thanhldt060802/dtos"
	"thanhldt060802/model"
	"thanhldt060802/repository"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type (
	ITaskService interface {
		Gets(ctx context.Context, filter *dtos.GetsTaskFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error)
		GetsCustom(ctx context.Context, filter *dtos.GetsTaskCustomFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error)
		GetById(ctx context.Context, id uuid.UUID) (*model.TaskView, error)
		Create(ctx context.Context, dto *dtos.CreateTaskDTO) (*model.Task, error)
		UpdateById(ctx context.Context, id uuid.UUID, dto *dtos.UpdateTaskDTO) (*model.Task, error)
		PatchById(ctx context.Context, id uuid.UUID, dto *dtos.PatchTaskDTO) (*model.Task, error)
		DeleteById(ctx context.Context, id uuid.UUID) error
	}
	TaskService struct {
		repo repository.ITaskRepo
	}
)

func NewTaskService(repo repository.ITaskRepo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Gets(ctx context.Context, filter *dtos.GetsTaskFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error) {
	tasks, total, err := s.repo.GetsView(ctx, filter, limit, offset, sorts)

	if err != nil {
		log.Error("Failed to get tasks:", err)
		return nil, 0, apperror.ErrServiceUnavailable(nil, err.Error())
	}

	return tasks, total, nil
}

func (s *TaskService) GetsCustom(ctx context.Context, filter *dtos.GetsTaskCustomFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error) {
	tasks, total, err := s.repo.GetsViewCustom(ctx, filter, limit, offset, sorts)

	if err != nil {
		log.Error("Failed to get tasks:", err)
		return nil, 0, apperror.ErrServiceUnavailable(nil, err.Error())
	}

	return tasks, total, nil
}

func (s *TaskService) GetById(ctx context.Context, id uuid.UUID) (*model.TaskView, error) {
	task, err := s.repo.GetViewById(ctx, id)

	if err != nil {
		log.Error("Task not found:", err)
		return nil, apperror.ErrNotFound("task not found", "")
	}

	return task, nil
}

func (s *TaskService) Create(ctx context.Context, dto *dtos.CreateTaskDTO) (*model.Task, error) {
	task := model.Task{}

	task.Id = uuid.New().String()

	if dto.Password != nil && *dto.Password != "" {
		task.Password = dto.Password
	}

	task.TaskName = dto.TaskName

	if dto.Description != nil && *dto.Description != "" {
		task.Description = dto.Description
	}

	task.State = dto.State

	task.Priority = dto.Priority

	task.Progress = dto.Progress

	task.CreatedBy = uuid.New().String()

	if err := s.repo.Create(ctx, &task); err != nil {
		log.Error("Failed to create task:", err)
		return nil, apperror.ErrServiceUnavailable(nil, err.Error())
	}

	return &task, nil
}

func (s *TaskService) UpdateById(ctx context.Context, id uuid.UUID, dto *dtos.UpdateTaskDTO) (*model.Task, error) {
	existingTask, err := s.repo.GetById(ctx, id)
	if err != nil {
		log.Error("Task not found:", err)
		return nil, apperror.ErrNotFound("task not found", "")
	}

	if dto.Password != "" {
		existingTask.Password = &dto.Password
	} else {
		existingTask.Password = nil
	}

	existingTask.TaskName = dto.TaskName

	if dto.Description != "" {
		existingTask.Description = &dto.Description
	} else {
		existingTask.Description = nil
	}

	existingTask.State = dto.State

	existingTask.Priority = dto.Priority

	existingTask.Progress = dto.Progress

	updatedBy := uuid.New().String()
	existingTask.UpdatedBy = &updatedBy

	if err := s.repo.UpdateById(ctx, id, existingTask); err != nil {
		log.Error("Failed to update task:", err)
		return nil, apperror.ErrServiceUnavailable(nil, err.Error())
	}

	return existingTask, nil
}

func (s *TaskService) PatchById(ctx context.Context, id uuid.UUID, dto *dtos.PatchTaskDTO) (*model.Task, error) {
	if _, err := s.repo.GetById(ctx, id); err != nil {
		log.Error("Task not found:", err)
		return nil, apperror.ErrNotFound("task not found", "")
	}

	task := model.Task{}

	if dto.Password != nil {
		task.Password = dto.Password
	}

	if dto.TaskName != nil {
		task.TaskName = *dto.TaskName
	}

	if dto.Description != nil {
		task.Description = dto.Description
	}

	if dto.State != nil {
		task.State = *dto.State
	}

	if dto.Priority != nil {
		task.Priority = *dto.Priority
	}

	if dto.Progress != nil {
		task.Progress = *dto.Progress
	}

	updatedBy := uuid.New().String()
	task.UpdatedBy = &updatedBy

	if err := s.repo.PatchById(ctx, id, &task); err != nil {
		log.Error("Failed to update task:", err)
		return nil, apperror.ErrServiceUnavailable(nil, err.Error())
	}

	existingTask, _ := s.repo.GetById(ctx, id)

	return existingTask, nil
}

func (s *TaskService) DeleteById(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetById(ctx, id); err != nil {
		log.Error("Task not found:", err)
		return apperror.ErrNotFound("task not found", "")
	}

	if err := s.repo.DeleteById(ctx, id); err != nil {
		log.Error("Failed to delete task:", err)
		return apperror.ErrServiceUnavailable(nil, err.Error())
	}

	return nil
}

package service

import (
	"context"
	"fmt"
	"thanhldt060802/internal/actormodel/types"
	"thanhldt060802/internal/dto"
	"thanhldt060802/internal/repository"

	"ergo.services/ergo/gen"
)

type taskService struct {
	taskRepository repository.TaskRepository
	node           gen.Node
	supervisorPID  gen.PID
}

type TaskService interface {
	GetExistedWorkers(ctx context.Context) (*dto.ExistedWorkers, error)
	DispatchTask(ctx context.Context, reqDTO *dto.DispatchTaskRequest) error
	RunTask(ctx context.Context, reqDTO *dto.RunTaskRequest) error
	RunTaskList(ctx context.Context, reqDTO *dto.RunTaskListRequest) error
}

func NewTaskService(taskRepository repository.TaskRepository, node gen.Node, supervisorPID gen.PID) TaskService {
	return &taskService{
		taskRepository: taskRepository,
		node:           node,
		supervisorPID:  supervisorPID,
	}
}

func (taskService *taskService) GetExistedWorkers(ctx context.Context) (*dto.ExistedWorkers, error) {
	workerNamesChan := make(chan []string)
	runningChan := make(chan []string)
	availableChan := make(chan []string)

	message := types.GetExistedWorkersMessage{
		WorkerNamesChan: workerNamesChan,
		RunningChan:     runningChan,
		AvailableChan:   availableChan,
	}

	if err := taskService.node.Send(taskService.supervisorPID, message); err != nil {
		return nil, fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	existedWorkers := &dto.ExistedWorkers{}
	existedWorkers.WorkerNames = <-workerNamesChan
	existedWorkers.Running = <-runningChan
	existedWorkers.Available = <-availableChan

	return existedWorkers, nil
}

func (taskService *taskService) DispatchTask(ctx context.Context, reqDTO *dto.DispatchTaskRequest) error {
	message := types.DispatchTaskMessage{
		WorkerName: reqDTO.Body.WorkerName,
		TaskId:     reqDTO.Body.TaskId,
	}
	if err := taskService.node.Send(taskService.supervisorPID, message); err != nil {
		return fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	return nil
}

func (taskService *taskService) RunTask(ctx context.Context, reqDTO *dto.RunTaskRequest) error {
	message := types.RunTaskMessage{
		TaskId: reqDTO.Body.TaskId,
	}
	if err := taskService.node.Send(taskService.supervisorPID, message); err != nil {
		return fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	return nil
}

func (taskService *taskService) RunTaskList(ctx context.Context, reqDTO *dto.RunTaskListRequest) error {
	message := types.RunTaskListMessage{
		TaskIdList: reqDTO.Body.TaskIdList,
	}
	if err := taskService.node.Send(taskService.supervisorPID, message); err != nil {
		return fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	return nil
}

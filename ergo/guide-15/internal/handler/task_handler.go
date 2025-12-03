package handler

import (
	"context"
	"net/http"
	"thanhldt060802/internal/dto"
	"thanhldt060802/internal/service"

	"github.com/danielgtaylor/huma/v2"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(api huma.API, taskService service.TaskService) *TaskHandler {
	taskHandler := &TaskHandler{
		taskService: taskService,
	}

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/get-existed-workers",
		Summary:     "/get-existed-workers",
		Description: "Get existed workers.",
		Tags:        []string{"Demo"},
	}, taskHandler.GetExistedWorkers)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/dispatch-task",
		Summary:     "/dispatch-task",
		Description: "Dispatch task.",
		Tags:        []string{"Demo"},
	}, taskHandler.DispatchTask)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/run-task",
		Summary:     "/run-task",
		Description: "Run task.",
		Tags:        []string{"Demo"},
	}, taskHandler.RunTask)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/run-task-list",
		Summary:     "/run-task-list",
		Description: "Run task list.",
		Tags:        []string{"Demo"},
	}, taskHandler.RunTaskList)

	return taskHandler
}

func (taskHandler *TaskHandler) GetExistedWorkers(ctx context.Context, _ *struct{}) (*dto.BodyResponse[dto.ExistedWorkers], error) {
	existedWorkers, err := taskHandler.taskService.GetExistedWorkers(ctx)
	if err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Get existed workers failed"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.BodyResponse[dto.ExistedWorkers]{}
	res.Body.Code = "OK"
	res.Body.Message = "Get existed workers successful"
	res.Body.Data = *existedWorkers
	return res, nil
}

func (taskHandler *TaskHandler) DispatchTask(ctx context.Context, reqDTO *dto.DispatchTaskRequest) (*dto.SuccessResponse, error) {
	if err := taskHandler.taskService.DispatchTask(ctx, reqDTO); err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Dispatch task failed"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.SuccessResponse{}
	res.Body.Code = "OK"
	res.Body.Message = "Dispatch task successful"
	return res, nil
}

func (taskHandler *TaskHandler) RunTask(ctx context.Context, reqDTO *dto.RunTaskRequest) (*dto.SuccessResponse, error) {
	if err := taskHandler.taskService.RunTask(ctx, reqDTO); err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Run task failed"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.SuccessResponse{}
	res.Body.Code = "OK"
	res.Body.Message = "Run task successful"
	return res, nil
}

func (taskHandler *TaskHandler) RunTaskList(ctx context.Context, reqDTO *dto.RunTaskListRequest) (*dto.SuccessResponse, error) {
	if err := taskHandler.taskService.RunTaskList(ctx, reqDTO); err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Run taks failed"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.SuccessResponse{}
	res.Body.Code = "OK"
	res.Body.Message = "Run tasks successful"
	return res, nil
}

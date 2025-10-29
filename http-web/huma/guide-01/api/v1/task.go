package v4

import (
	"context"
	"net/http"

	"thanhldt060802/common/response"
	"thanhldt060802/common/util"
	"thanhldt060802/dtos"
	authMdw "thanhldt060802/middleware/auth"
	"thanhldt060802/model"
	"thanhldt060802/service"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type apiTask struct {
	taskService service.ITaskService
}

func RegisterAPITask(api hureg.APIGen, taskService service.ITaskService) {
	handler := &apiTask{
		taskService: taskService,
	}

	apiGroup := api.AddBasePath("/task")

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "get-list-tasks",
			Method:      http.MethodGet,
			Path:        "",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get list tasks.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.Gets,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "custom-get-list-tasks",
			Method:      http.MethodGet,
			Path:        "/custom",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Custom get list tasks.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.GetsCustom,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "get-task-by-id",
			Method:      http.MethodGet,
			Path:        "/{id}",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get task by id.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.GetById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "create-task",
			Method:      http.MethodPost,
			Path:        "",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Create task.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.Create,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "custom-create-task",
			Method:      http.MethodPost,
			Path:        "/custom",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Custom create task.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.CreateCustom,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "update-task-by-id",
			Method:      http.MethodPut,
			Path:        "/{id}",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Update task by id.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.UpdateById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "patch-task-by-id",
			Method:      http.MethodPatch,
			Path:        "/{id}",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Patch task by id.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.PatchById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "custom-patch-task-by-id",
			Method:      http.MethodPatch,
			Path:        "/custom/{id}",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Custom patch task by id.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.PatchCustomById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			Tags:        []string{"Task"},
			OperationID: "delete-task-by-id",
			Method:      http.MethodDelete,
			Path:        "/{id}",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Delete task by id.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.DeleteById,
	)
}

func (handler *apiTask) Gets(ctx context.Context, req *struct {
	dtos.PagingCommon
	dtos.GetsTaskFilter
}) (resp *response.PaginationResponse[[]*model.TaskView], err error) {
	tasks, total, err := handler.taskService.Gets(ctx, &req.GetsTaskFilter, req.Limit, req.Offset, util.ParseSortBy(req.SortBy))
	if err != nil {
		log.Error("Failed to get tasks:", err)
		return
	}

	resp = response.Pagination(tasks, total, "success")
	return
}

func (handler *apiTask) GetsCustom(ctx context.Context, req *struct {
	dtos.PagingCommon
	dtos.GetsTaskCustomFilter
}) (resp *response.PaginationResponse[[]*model.TaskView], err error) {
	tasks, total, err := handler.taskService.GetsCustom(ctx, &req.GetsTaskCustomFilter, req.Limit, req.Offset, util.ParseSortBy(req.SortBy))
	if err != nil {
		log.Error("Failed to get tasks:", err)
		return
	}

	resp = response.Pagination(tasks, total, "success")
	return
}

func (handler *apiTask) GetById(ctx context.Context, req *struct {
	Id string `path:"id" format:"uuid"`
}) (resp *response.GenericResponse[*model.TaskView], err error) {
	id, _ := uuid.Parse(req.Id)

	task, err := handler.taskService.GetById(ctx, id)
	if err != nil {
		log.Error("Failed to get task by id:", err)
		return
	}

	resp = response.OK(task, "success")
	return
}

func (handler *apiTask) Create(ctx context.Context, req *struct {
	Body *dtos.CreateTaskDTO
}) (res *response.GenericResponse[*model.Task], err error) {
	task, err := handler.taskService.Create(ctx, req.Body)
	if err != nil {
		log.Error("Failed to create task:", err)
		return
	}

	res = response.OK(task, "sucess")
	return
}

func (handler *apiTask) CreateCustom(ctx context.Context, req *struct {
	Body *dtos.CreateTaskCustomDTO
}) (res *response.GenericResponse[*model.Task], err error) {
	convertDTO := dtos.CreateTaskDTO{
		Password:    req.Body.Password,
		TaskName:    req.Body.TaskName,
		Description: req.Body.Description,
		State:       "todo",
		Priority:    "medium",
		Progress:    0,
	}

	task, err := handler.taskService.Create(ctx, &convertDTO)
	if err != nil {
		log.Error("Failed to create task:", err)
		return
	}

	res = response.OK(task, "success")
	return
}

func (handler *apiTask) UpdateById(ctx context.Context, req *struct {
	Id   string `path:"id" format:"uuid"`
	Body *dtos.UpdateTaskDTO
}) (res *response.GenericResponse[map[string]any], err error) {
	id, _ := uuid.Parse(req.Id)

	updatedId, err := handler.taskService.UpdateById(ctx, id, req.Body)
	if err != nil {
		log.Error("Failed to update task:", err)
		return
	}

	res = response.OK(map[string]any{
		"task_uuid": updatedId,
	}, "success")
	return
}

func (handler *apiTask) PatchById(ctx context.Context, req *struct {
	Id   string `path:"id" format:"uuid"`
	Body *dtos.PatchTaskDTO
}) (res *response.GenericResponse[map[string]any], err error) {
	id, _ := uuid.Parse(req.Id)

	updatedId, err := handler.taskService.PatchById(ctx, id, req.Body)
	if err != nil {
		log.Error("Failed to update task:", err)
		return
	}

	res = response.OK(map[string]any{
		"task_uuid": updatedId,
	}, "success")
	return
}

func (handler *apiTask) PatchCustomById(ctx context.Context, req *struct {
	Id   string `path:"id" format:"uuid"`
	Body *dtos.PatchTaskCustomDTO
}) (res *response.GenericResponse[map[string]any], err error) {
	id, _ := uuid.Parse(req.Id)

	convertDTO := dtos.PatchTaskDTO{
		State:    &req.Body.State,
		Progress: &req.Body.Progress,
	}

	message, err := handler.taskService.PatchById(ctx, id, &convertDTO)
	if err != nil {
		log.Error("Failed to update task:", err)
		return
	}

	res = response.OK(map[string]any{
		"message": message,
	}, "success")
	return
}

func (handler *apiTask) DeleteById(ctx context.Context, req *struct {
	Id string `path:"id" format:"uuid"`
}) (res *response.GenericResponse[any], err error) {
	id, _ := uuid.Parse(req.Id)

	if err = handler.taskService.DeleteById(ctx, id); err != nil {
		log.Error("Failed to delete task:", err)
		return
	}

	res = response.OK_Only("success")
	return
}

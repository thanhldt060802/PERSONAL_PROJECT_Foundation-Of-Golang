package v1

import (
	"context"
	"net/http"
	"thanhldt060802/common/response"
	"thanhldt060802/internal"
	"thanhldt060802/model"
	"thanhldt060802/service"

	authMdw "thanhldt060802/middleware/auth"

	"github.com/cardinalby/hureg"
	"github.com/danielgtaylor/huma/v2"
)

type apiExample struct {
	exampleService service.IExampleService
}

func RegisterAPIExample(api hureg.APIGen, exampleService service.IExampleService) {
	handler := &apiExample{
		exampleService: exampleService,
	}

	apiGroup := api.AddBasePath("/example")

	hureg.Register(
		apiGroup,
		huma.Operation{
			OperationID: "get-example-by-id",
			Method:      http.MethodGet,
			Path:        "/{example_uuid}",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get example by id.",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.GetById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			OperationID: "cross-service-get-example-by-id",
			Method:      http.MethodGet,
			Path:        "/{example_uuid}/cross-service",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get example by id (cross-service).",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.CrossService_GetById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			OperationID: "pub-sub-get-example-by-id",
			Method:      http.MethodGet,
			Path:        "/{example_uuid}/pub-sub",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get example by id (pub-sub).",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.PubSub_GetById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			OperationID: "hybrid-get-example-by-id",
			Method:      http.MethodGet,
			Path:        "/{example_uuid}/hybrid",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get example by id (hybrid).",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.Hybrid_GetById,
	)

	hureg.Register(
		apiGroup,
		huma.Operation{
			OperationID: "bulk-async-get-example-by-id",
			Method:      http.MethodGet,
			Path:        "/{example_uuid}/bulk-async",
			Security:    authMdw.DefaultAuthSecurity,
			Description: "Get example by id (bulk async).",
			Middlewares: huma.Middlewares{authMdw.NewAuthMiddleware(api)},
		},
		handler.BulkAsync_GetById,
	)
}

func (handler *apiExample) GetById(ctx context.Context, req *struct {
	ExampleUuid string `path:"example_uuid" format:"uuid" doc:"Example uuid"`
}) (res *response.GenericResponse[*model.Example], err error) {
	ctx, span := internal.Observer.NewSpan(ctx, "GetExampleById-Handler")
	defer span.Done()

	internal.Observer.InfoLogWithCtx(ctx, "[Handler layer] - Get Example by example_uuid='%s'", req.ExampleUuid)

	example, err := handler.exampleService.GetById(ctx, req.ExampleUuid)
	if err != nil {
		internal.Observer.ErrorLogWithCtx(ctx, "[Handler layer] - Failed to get Example by example_uuid='%s': %v", req.ExampleUuid, err)
		return
	}

	res = response.Ok(example)
	return
}

func (handler *apiExample) CrossService_GetById(ctx context.Context, req *struct {
	ExampleUuid string `path:"example_uuid" format:"uuid" doc:"Example uuid"`
}) (res *response.GenericResponse[*model.Example], err error) {
	ctx, span := internal.Observer.NewSpan(ctx, "CrossService_GetExampleById-Handler")
	defer span.Done()

	example, err := handler.exampleService.CrossService_GetById(ctx, req.ExampleUuid)
	if err != nil {
		span.SetError(err)
		return
	}

	res = response.Ok(example)
	return
}

func (handler *apiExample) PubSub_GetById(ctx context.Context, req *struct {
	ExampleUuid string `path:"example_uuid" format:"uuid" doc:"Example uuid"`
}) (res *response.GenericResponse[*string], err error) {
	ctx, span := internal.Observer.NewSpan(ctx, "PubSub_GetExampleById-Handler")
	defer span.Done()

	result, err := handler.exampleService.PubSub_GetById(ctx, req.ExampleUuid)
	if err != nil {
		span.SetError(err)
		return
	}

	res = response.Ok(&result)
	return res, nil
}

func (handler *apiExample) Hybrid_GetById(ctx context.Context, req *struct {
	ExampleUuid string `path:"example_uuid" format:"uuid" doc:"Example uuid"`
}) (res *response.GenericResponse[*string], err error) {
	ctx, span := internal.Observer.NewSpan(ctx, "Hybrid_GetExampleById-Handler")
	defer span.Done()

	result, err := handler.exampleService.Hybrid_GetById(ctx, req.ExampleUuid)
	if err != nil {
		span.SetError(err)
		return
	}

	res = response.Ok(&result)
	return res, nil
}

func (handler *apiExample) BulkAsync_GetById(ctx context.Context, req *struct {
	ExampleUuid string `path:"example_uuid" format:"uuid" doc:"Example uuid"`
}) (res *response.GenericResponse[*string], err error) {
	ctx, span := internal.Observer.NewSpan(ctx, "BulkAsync_GetExampleById-Handler")
	defer span.Done()

	result, err := handler.exampleService.BulkAsync_GetById(ctx, req.ExampleUuid)
	if err != nil {
		span.SetError(err)
		return
	}

	res = response.Ok(&result)
	return res, nil
}

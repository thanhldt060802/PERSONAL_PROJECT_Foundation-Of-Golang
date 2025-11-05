package v1

import (
	"context"
	"net/http"
	"thanhldt060802/common/observer"
	"thanhldt060802/common/response"
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
}

func (handler *apiExample) GetById(ctx context.Context, req *struct {
	ExampleUuid string `path:"example_uuid" format:"uuid" doc:"Example uuid"`
}) (res *response.GenericResponse[*model.Example], err error) {
	ctx, span := observer.StartSpanInternal(ctx)
	defer span.End()

	example, err := handler.exampleService.GetById(ctx, req.ExampleUuid)
	if err != nil {
		span.Err = err
		return
	}

	res = response.Ok(example)
	return
}

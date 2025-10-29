package main

import (
	"math/rand"
	"net/http"
	"thanhldt060802/internal/actormodel/app"
	"thanhldt060802/internal/dto"
	"thanhldt060802/internal/handler"
	"thanhldt060802/internal/repository"
	"thanhldt060802/internal/service"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

var humaDocsEmbedded = `<!doctype html>
	<html>
	  <head>
	    <title>Example 12</title>
	    <meta charset="utf-8" />
	    <meta name="viewport" content="width=device-width, initial-scale=1" />
	  </head>
	  <body>
	    <script
	      id="api-reference"
	      data-url="/openapi.json"></script>
	    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	  </body>
	</html>`

func main() {

	rand.New(rand.NewSource(time.Now().UnixNano()))

	humaCfg := huma.DefaultConfig("Example 12", "v1.0.0")
	humaCfg.DocsPath = ""
	humaCfg.JSONSchemaDialect = ""
	humaCfg.CreateHooks = nil

	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]string, len(errs))
		for i, err := range errs {
			details[i] = err.Error()
		}
		res := &dto.ErrorResponse{}
		res.Status = status
		res.Message = msg
		res.Details = details
		return res
	}

	r := gin.Default()
	r.GET("/docs", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html", []byte(humaDocsEmbedded))
	})

	api := humagin.New(r, humaCfg)

	taskRepository := repository.NewTaskRepository()

	myApp := app.New(taskRepository, "mynode", 3)
	myApp.Start()

	taskService := service.NewTaskService(taskRepository, myApp.Node(), myApp.SupervisorPID())

	handler.NewTaskHandler(api, taskService)

	r.Run(":8080")

}

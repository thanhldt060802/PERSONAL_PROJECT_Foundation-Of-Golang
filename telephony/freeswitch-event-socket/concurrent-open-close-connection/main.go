package main

import (
	"math/rand"
	"net/http"
	"thanhldt060802/actormodel/app"
	"thanhldt060802/dto"
	"thanhldt060802/esl"
	"thanhldt060802/handler"
	"thanhldt060802/service"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

var humaDocsEmbedded = `<!doctype html>
	<html>
	  <head>
	    <title>FreeSwitchESLLab</title>
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

	humaCfg := huma.DefaultConfig("FreeSwitchESLLab", "v1.0.0")
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

	eslConfig := esl.ESLConfig{
		Address:  "103.72.97.156",
		Port:     7021,
		Password: "TEL4VN.COM",
	}
	myApp := app.New("mynode", eslConfig, 5)
	myApp.Start()

	cmdService := service.NewCmdService(myApp.Node(), myApp.SupervisorPID())

	handler.NewCmdHandler(api, cmdService)

	r.Run(":8080")

}

package handler

import (
	"context"
	"net/http"
	"thanhldt060802/dto"
	"thanhldt060802/service"

	"github.com/danielgtaylor/huma/v2"
)

type CmdHandler struct {
	cmdService service.CmdService
}

func NewCmdHandler(api huma.API, cmdService service.CmdService) *CmdHandler {
	cmdHandler := &CmdHandler{
		cmdService: cmdService,
	}

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/get-existed-workers",
		Summary:     "/get-existed-workers",
		Description: "Get existed workers.",
		Tags:        []string{"Demo"},
	}, cmdHandler.GetExistedWorkers)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/open-connection",
		Summary:     "/open-connection",
		Description: "Open connection.",
		Tags:        []string{"Demo"},
	}, cmdHandler.OpenConnection)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/close-connection",
		Summary:     "/close-connection",
		Description: "Close connection.",
		Tags:        []string{"Demo"},
	}, cmdHandler.CloseConnection)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/run-campaign-list",
		Summary:     "/run-campaign-list",
		Description: "Run campaign list.",
		Tags:        []string{"Demo"},
	}, cmdHandler.RunCampaignList)

	return cmdHandler
}

func (cmdHandler *CmdHandler) GetExistedWorkers(ctx context.Context, _ *struct{}) (*dto.BodyResponse[dto.ExistedWorkers], error) {
	existedWorkers, err := cmdHandler.cmdService.GetExistedWorkers(ctx)
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

func (cmdHandler *CmdHandler) OpenConnection(ctx context.Context, reqDTO *dto.OpenConnectionRequest) (*dto.SuccessResponse, error) {
	if err := cmdHandler.cmdService.OpenConnection(ctx, reqDTO); err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Open connection failed"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.SuccessResponse{}
	res.Body.Code = "OK"
	res.Body.Message = "Open connection successful"
	return res, nil
}

func (cmdHandler *CmdHandler) CloseConnection(ctx context.Context, reqDTO *dto.CloseConnectionRequest) (*dto.SuccessResponse, error) {
	if err := cmdHandler.cmdService.CloseConnection(ctx, reqDTO); err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Close connection failed"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.SuccessResponse{}
	res.Body.Code = "OK"
	res.Body.Message = "Close connection successful"
	return res, nil
}

func (cmdHandler *CmdHandler) RunCampaignList(ctx context.Context, reqDTO *dto.RunCampaignListRequest) (*dto.BodyResponse[[]dto.SummaryReport], error) {
	summaryReports, err := cmdHandler.cmdService.RunCampaignList(ctx, reqDTO)
	if err != nil {
		res := &dto.ErrorResponse{}
		res.Status = http.StatusInternalServerError
		res.Code = "ERR_INTERNAL_SERVER"
		res.Message = "Run campaign list"
		res.Details = []string{err.Error()}
		return nil, res
	}

	res := &dto.BodyResponse[[]dto.SummaryReport]{}
	res.Body.Code = "OK"
	res.Body.Message = "Run campaign list"
	res.Body.Data = summaryReports
	return res, nil
}

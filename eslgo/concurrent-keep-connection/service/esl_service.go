package service

import (
	"context"
	"fmt"
	"thanhldt060802/actormodel/types"
	"thanhldt060802/dto"

	"ergo.services/ergo/gen"
)

type cmdService struct {
	node          gen.Node
	supervisorPID gen.PID
}

type CmdService interface {
	GetExistedWorkers(ctx context.Context) (*dto.ExistedWorkers, error)
	OpenConnection(ctx context.Context, reqDTO *dto.OpenConnectionRequest) error
	CloseConnection(ctx context.Context, reqDTO *dto.CloseConnectionRequest) error
	RunCampaignList(ctx context.Context, reqDTO *dto.RunCampaignListRequest) ([]dto.SummaryReport, error)
}

func NewCmdService(node gen.Node, supervisorPID gen.PID) CmdService {
	return &cmdService{
		node:          node,
		supervisorPID: supervisorPID,
	}
}

func (cmdService *cmdService) GetExistedWorkers(ctx context.Context) (*dto.ExistedWorkers, error) {
	workerNamesChan := make(chan []string)
	runningChan := make(chan []string)
	availableChan := make(chan []string)
	sleepChan := make(chan []string)

	message := types.GetExistedWorkersMessage{
		WorkerNamesChan: workerNamesChan,
		RunningChan:     runningChan,
		AvailableChan:   availableChan,
		SleepChan:       sleepChan,
	}

	if err := cmdService.node.Send(cmdService.supervisorPID, message); err != nil {
		return nil, fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	existedWorkersDTO := &dto.ExistedWorkers{}
	existedWorkersDTO.WorkerNames = <-workerNamesChan
	existedWorkersDTO.Running = <-runningChan
	existedWorkersDTO.Available = <-availableChan
	existedWorkersDTO.Sleep = <-sleepChan

	return existedWorkersDTO, nil
}

func (cmdService *cmdService) OpenConnection(ctx context.Context, reqDTO *dto.OpenConnectionRequest) error {
	message := types.OpenConnectionMessage{
		WorkerName: reqDTO.Body.WorkerName,
	}
	if err := cmdService.node.Send(cmdService.supervisorPID, message); err != nil {
		return fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	return nil
}

func (cmdService *cmdService) CloseConnection(ctx context.Context, reqDTO *dto.CloseConnectionRequest) error {
	message := types.CloseConnectionMessage{
		WorkerName: reqDTO.Body.WorkerName,
	}
	if err := cmdService.node.Send(cmdService.supervisorPID, message); err != nil {
		return fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	return nil
}

func (cmdService *cmdService) RunCampaignList(ctx context.Context, reqDTO *dto.RunCampaignListRequest) ([]dto.SummaryReport, error) {
	maxSize := len(reqDTO.Body.CampaignList)

	summaryReportChan := make(chan types.SummaryReport, maxSize)

	campaignInfoList := []types.CampaignInfo{}
	for _, campaign := range reqDTO.Body.CampaignList {
		campaignInfoList = append(campaignInfoList, types.CampaignInfo{
			CampaignName: campaign.CampaignName,
			CmdList:      campaign.CmdList,
		})
	}

	message := types.RunCampaignListMessage{
		CampaignInfoList:  campaignInfoList,
		SummaryReportChan: summaryReportChan,
	}

	if err := cmdService.node.Send(cmdService.supervisorPID, message); err != nil {
		return nil, fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	summaryReportDTOs := []dto.SummaryReport{}
	i := 0
	for summaryReport := range summaryReportChan {
		summaryReportDTO := dto.SummaryReport{
			CampaignName:  summaryReport.CampaignName,
			Total:         summaryReport.Total,
			Complete:      summaryReport.Complete,
			TotalTime:     summaryReport.TotalTime,
			WorkerReports: []dto.WorkerReport{},
		}
		for _, workerReport := range summaryReport.WorkerReports {
			summaryReportDTO.WorkerReports = append(summaryReportDTO.WorkerReports, dto.WorkerReport{
				WorkerName: workerReport.WorkerName,
				DelayTime:  workerReport.DelayTime,
				Cmd:        workerReport.Cmd,
			})
		}
		summaryReportDTOs = append(summaryReportDTOs, summaryReportDTO)

		i++
		if i >= maxSize {
			break
		}
	}

	return summaryReportDTOs, nil
}

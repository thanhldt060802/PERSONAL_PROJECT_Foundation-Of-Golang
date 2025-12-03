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
	DispatchCmd(ctx context.Context, reqDTO *dto.DispatchCmdRequest) (*dto.WorkerReport, error)
	DispatchCmdList(ctx context.Context, reqDTO *dto.DispatchCmdListRequest) (*dto.SummaryReport, error)
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

func (cmdService *cmdService) DispatchCmd(ctx context.Context, reqDTO *dto.DispatchCmdRequest) (*dto.WorkerReport, error) {
	workerReportChan := make(chan types.WorkerReport)

	message := types.DispatchCmdMessage{
		WorkerName:       reqDTO.Body.WorkerName,
		Cmd:              reqDTO.Body.Cmd,
		WorkerReportChan: workerReportChan,
	}

	if err := cmdService.node.Send(cmdService.supervisorPID, message); err != nil {
		return nil, fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	workerReport := <-workerReportChan

	workerReportDTO := &dto.WorkerReport{}
	workerReportDTO.WorkerName = workerReport.WorkerName
	workerReportDTO.Cmd = workerReport.Cmd
	workerReportDTO.DelayTime = workerReport.DelayTime

	return workerReportDTO, nil
}

func (cmdService *cmdService) DispatchCmdList(ctx context.Context, reqDTO *dto.DispatchCmdListRequest) (*dto.SummaryReport, error) {
	summaryReportChan := make(chan types.SummaryReport)

	message := types.DispatchCmdListMessage{
		CmdList:           reqDTO.Body.CmdList,
		SummaryReportChan: summaryReportChan,
	}

	if err := cmdService.node.Send(cmdService.supervisorPID, message); err != nil {
		return nil, fmt.Errorf("some thing wrong on actor model: %v", err.Error())
	}

	summaryReport := <-summaryReportChan

	summaryReportDTO := &dto.SummaryReport{}
	summaryReportDTO.Total = summaryReport.Total
	summaryReportDTO.Complete = summaryReport.Complete
	summaryReportDTO.TotalTime = summaryReport.TotalTime
	summaryReportDTO.WorkerReports = []dto.WorkerReport{}
	for _, workerReport := range summaryReport.WorkerReports {
		workerReportDTO := dto.WorkerReport{}
		workerReportDTO.WorkerName = workerReport.WorkerName
		workerReportDTO.Cmd = workerReport.Cmd
		workerReportDTO.DelayTime = workerReport.DelayTime
		summaryReportDTO.WorkerReports = append(summaryReportDTO.WorkerReports, workerReportDTO)
	}

	return summaryReportDTO, nil
}

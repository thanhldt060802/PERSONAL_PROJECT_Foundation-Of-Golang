package app

import (
	"fmt"
	"sync"
	"thanhldt060802/actormodel/types"
	"thanhldt060802/esl"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	WORKER_RUNNING   = "running"
	WORKER_AVAILABLE = "available"
	WORKER_SLEEP     = "sleep"
)

type WorkerSupervisor struct {
	act.Supervisor

	eslConfig             esl.ESLConfig
	numberOfInitialWorker int

	workerStatusMap      map[string]string
	workerStatusMapMutex sync.Mutex

	campaignInfoMap      map[string]*types.CampaignInfo
	campaignInfoMapMutex sync.Mutex

	summaryReportChanMap      map[string]chan types.SummaryReport
	summaryReportChanMapMutex sync.Mutex
}

func FactoryWorkerSupervisor() gen.ProcessBehavior {
	return &WorkerSupervisor{}
}

func (workerSupervisor *WorkerSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	workerSupervisor.eslConfig = args[0].(esl.ESLConfig)
	workerSupervisor.numberOfInitialWorker = args[1].(int)
	workerSupervisor.workerStatusMap = map[string]string{}
	workerSupervisor.workerStatusMapMutex = sync.Mutex{}
	workerSupervisor.campaignInfoMap = map[string]*types.CampaignInfo{}
	workerSupervisor.campaignInfoMapMutex = sync.Mutex{}
	workerSupervisor.summaryReportChanMap = map[string]chan types.SummaryReport{}
	workerSupervisor.summaryReportChanMapMutex = sync.Mutex{}

	supervisorSpec := act.SupervisorSpec{}
	supervisorSpec.EnableHandleChild = true
	supervisorSpec.DisableAutoShutdown = true
	supervisorSpec.Type = act.SupervisorTypeOneForOne
	supervisorSpec.Restart.Strategy = act.SupervisorStrategyTransient
	supervisorSpec.Restart.Intensity = 100
	supervisorSpec.Restart.Period = 5

	supervisorSpec.Children = []act.SupervisorChildSpec{}
	for i := 1; i <= workerSupervisor.numberOfInitialWorker; i++ {
		supervisorSpec.Children = append(supervisorSpec.Children, act.SupervisorChildSpec{
			Name:    gen.Atom(fmt.Sprintf("worker_%v", i)),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args: []any{esl.ESLConfig{
				Address:  workerSupervisor.eslConfig.Address,
				Port:     workerSupervisor.eslConfig.Port,
				Password: workerSupervisor.eslConfig.Password,
			}},
		})
	}

	workerSupervisor.Log().Info("Started worker supervisor %v %v on %v", workerSupervisor.PID(), workerSupervisor.Name(), workerSupervisor.Node().Name())
	return supervisorSpec, nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildStart(childName gen.Atom, pid gen.PID) error {
	workerSupervisor.Log().Info("Actor start with name %v and PID %v", childName, pid)

	workerName := childName.String()
	workerName = workerName[1 : len(workerName)-1]

	workerSupervisor.campaignInfoMapMutex.Lock()
	defer workerSupervisor.campaignInfoMapMutex.Unlock()
	if campaignInfo, ok := workerSupervisor.campaignInfoMap[workerName]; ok {
		workerSupervisor.Send(childName, types.DoProcessCall{
			Cmd:              campaignInfo.CmdList[campaignInfo.Index],
			WorkerReportChan: campaignInfo.WorkerReportChans[campaignInfo.Index]},
		)
	} else {
		workerSupervisor.workerStatusMapMutex.Lock()
		defer workerSupervisor.workerStatusMapMutex.Unlock()
		workerSupervisor.workerStatusMap[workerName] = WORKER_AVAILABLE
	}

	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildTerminate(name gen.Atom, pid gen.PID, reason error) error {
	workerSupervisor.Log().Error("Actor %v terminated. Panic reason: %v", name, reason.Error())

	if reason.Error() == gen.TerminateReasonNormal.Error() {
		workerName := name.String()
		workerName = workerName[1 : len(workerName)-1]

		workerSupervisor.workerStatusMapMutex.Lock()
		defer workerSupervisor.workerStatusMapMutex.Unlock()
		workerSupervisor.workerStatusMap[workerName] = WORKER_SLEEP
	}

	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleMessage(from gen.PID, message any) error {
	switch receivedMessage := message.(type) {
	case types.CompleteMessage:
		{
			workerSupervisor.campaignInfoMapMutex.Lock()
			defer workerSupervisor.campaignInfoMapMutex.Unlock()
			campaignInfo := workerSupervisor.campaignInfoMap[receivedMessage.WorkerName]
			campaignInfo.Index++
			campaignInfo.Complete++
			if campaignInfo.Index < len(campaignInfo.CmdList) {
				workerSupervisor.Send(from, types.DoProcessCall{
					Cmd:              campaignInfo.CmdList[campaignInfo.Index],
					WorkerReportChan: campaignInfo.WorkerReportChans[campaignInfo.Index],
				})
			} else {
				campaignInfo.EndTime = time.Now()

				workerReports := []types.WorkerReport{}
				for _, workerReportChan := range campaignInfo.WorkerReportChans {
					workerReports = append(workerReports, <-workerReportChan)
				}

				summaryReport := types.SummaryReport{
					CampaignName:  campaignInfo.CampaignName,
					Total:         len(campaignInfo.CmdList),
					Complete:      campaignInfo.Complete,
					TotalTime:     fmt.Sprintf("%v", campaignInfo.EndTime.Sub(campaignInfo.StartTime)),
					WorkerReports: workerReports,
				}

				workerSupervisor.summaryReportChanMapMutex.Lock()
				defer workerSupervisor.summaryReportChanMapMutex.Unlock()
				workerSupervisor.summaryReportChanMap[receivedMessage.WorkerName] <- summaryReport
				delete(workerSupervisor.summaryReportChanMap, receivedMessage.WorkerName)

				delete(workerSupervisor.campaignInfoMap, receivedMessage.WorkerName)

				// Worker Done...!
				workerSupervisor.workerStatusMapMutex.Lock()
				defer workerSupervisor.workerStatusMapMutex.Unlock()
				workerSupervisor.workerStatusMap[receivedMessage.WorkerName] = WORKER_AVAILABLE
			}

			return nil
		}
	case types.GetExistedWorkersMessage:
		{
			workerSupervisor.getExistedWorkers(receivedMessage)
			return nil
		}
	case types.OpenConnectionMessage:
		{
			workerSupervisor.openConnection(receivedMessage)
			return nil
		}
	case types.CloseConnectionMessage:
		{
			workerSupervisor.closeConnection(receivedMessage)
			return nil
		}
	case types.RunCampaignListMessage:
		{
			workerSupervisor.runCampaignList(receivedMessage)
			return nil
		}
	}

	return nil
}

func (workerSupervisor *WorkerSupervisor) getExistedWorkers(message types.GetExistedWorkersMessage) {
	workerNames := []string{}
	running := []string{}
	available := []string{}
	sleep := []string{}

	workerSupervisor.workerStatusMapMutex.Lock()
	defer workerSupervisor.workerStatusMapMutex.Unlock()
	for _, supervisorChildSpec := range workerSupervisor.Children() {
		workerName := supervisorChildSpec.Name.String()
		workerName = workerName[1 : len(workerName)-1]

		workerNames = append(workerNames, workerName)
		switch workerSupervisor.workerStatusMap[workerName] {
		case WORKER_RUNNING:
			{
				running = append(running, workerName)
			}
		case WORKER_AVAILABLE:
			{
				available = append(available, workerName)
			}
		case WORKER_SLEEP:
			{
				sleep = append(sleep, workerName)
			}
		}
	}
	message.WorkerNamesChan <- workerNames
	message.RunningChan <- running
	message.AvailableChan <- available
	message.SleepChan <- sleep
}

func (workerSupervisor *WorkerSupervisor) openConnection(message types.OpenConnectionMessage) {
	workerSupervisor.workerStatusMapMutex.Lock()
	defer workerSupervisor.workerStatusMapMutex.Unlock()
	if status, ok := workerSupervisor.workerStatusMap[message.WorkerName]; ok {
		switch status {
		case WORKER_RUNNING:
			{
				workerSupervisor.Log().Warning("Actor %v is running", message.WorkerName)
			}
		case WORKER_AVAILABLE:
			{
				workerSupervisor.Log().Warning("Actor %v is available", message.WorkerName)
			}
		case WORKER_SLEEP:
			{
				workerSupervisor.Log().Info("Restart existed actor %v", message.WorkerName)
				workerSupervisor.StartChild(gen.Atom(message.WorkerName), esl.ESLConfig{
					Address:  workerSupervisor.eslConfig.Address,
					Port:     workerSupervisor.eslConfig.Port,
					Password: workerSupervisor.eslConfig.Password,
				})
				workerSupervisor.Log().Info("Restart existed actor %v successful", message.WorkerName)
			}
		}
	} else {
		workerSupervisor.Log().Info("Start new actor %v", message.WorkerName)
		if err := workerSupervisor.AddChild(act.SupervisorChildSpec{
			Name:    gen.Atom(message.WorkerName),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args: []any{esl.ESLConfig{
				Address:  workerSupervisor.eslConfig.Address,
				Port:     workerSupervisor.eslConfig.Port,
				Password: workerSupervisor.eslConfig.Password,
			}},
		}); err != nil {
			workerSupervisor.Log().Error("Start new actor %v failed: %v", message.WorkerName, err.Error())
		}
		workerSupervisor.Log().Info("Start new actor %v successful", message.WorkerName)
	}
}

func (workerSupervisor *WorkerSupervisor) closeConnection(message types.CloseConnectionMessage) {
	workerSupervisor.workerStatusMapMutex.Lock()
	defer workerSupervisor.workerStatusMapMutex.Unlock()
	if status, ok := workerSupervisor.workerStatusMap[message.WorkerName]; ok {
		switch status {
		case WORKER_RUNNING:
			{
				workerSupervisor.Log().Warning("Actor %v is running", message.WorkerName)
			}
		case WORKER_AVAILABLE:
			{
				workerSupervisor.Log().Info("Sending termanate signal to %v", message.WorkerName)
				workerSupervisor.Send(gen.Atom(message.WorkerName), types.DoSafetyTerminate{})
			}
		case WORKER_SLEEP:
			{
				workerSupervisor.Log().Warning("Actor %v si sleeping", message.WorkerName)
			}
		}
	} else {
		workerSupervisor.Log().Warning("Actor %v is not valid", message.WorkerName)
	}
}

func (workerSupervisor *WorkerSupervisor) runCampaignList(message types.RunCampaignListMessage) {
	workerSupervisor.workerStatusMapMutex.Lock()
	defer workerSupervisor.workerStatusMapMutex.Unlock()
	for _, campaignInfo := range message.CampaignInfoList {
		for workerName, status := range workerSupervisor.workerStatusMap {
			if status == WORKER_AVAILABLE {
				workerSupervisor.workerStatusMap[workerName] = WORKER_RUNNING
				campaignInfo.Index = 0
				campaignInfo.Complete = 0
				campaignInfo.StartTime = time.Now()
				campaignInfo.WorkerReportChans = []chan types.WorkerReport{}
				for i := 0; i < len(campaignInfo.CmdList); i++ {
					campaignInfo.WorkerReportChans = append(campaignInfo.WorkerReportChans, make(chan types.WorkerReport, 1))
				}
				workerSupervisor.campaignInfoMap[workerName] = &campaignInfo
				workerSupervisor.summaryReportChanMap[workerName] = message.SummaryReportChan
				workerSupervisor.Log().Info("Dispatch campaign %v for Worker %v", campaignInfo.CampaignName, workerName)
				workerSupervisor.Send(gen.Atom(workerName), types.DoProcessCall{
					Cmd:              campaignInfo.CmdList[campaignInfo.Index],
					WorkerReportChan: campaignInfo.WorkerReportChans[campaignInfo.Index]},
				)
				break
			}
		}
	}
}

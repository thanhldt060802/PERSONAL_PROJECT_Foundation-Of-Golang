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

	multiProcess bool
	complete     int
}

func FactoryWorkerSupervisor() gen.ProcessBehavior {
	return &WorkerSupervisor{}
}

func (workerSupervisor *WorkerSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	workerSupervisor.eslConfig = args[0].(esl.ESLConfig)
	workerSupervisor.numberOfInitialWorker = args[1].(int)
	workerSupervisor.workerStatusMap = map[string]string{}
	workerSupervisor.workerStatusMapMutex = sync.Mutex{}
	workerSupervisor.multiProcess = false

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
	workerName := childName.String()
	workerName = workerName[1 : len(workerName)-1]

	workerSupervisor.workerStatusMapMutex.Lock()
	workerSupervisor.workerStatusMap[workerName] = WORKER_AVAILABLE
	workerSupervisor.workerStatusMapMutex.Unlock()

	workerSupervisor.Log().Info("Actor start with name %v and PID %v", childName, pid)

	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildTerminate(name gen.Atom, pid gen.PID, reason error) error {
	if reason.Error() == gen.TerminateReasonNormal.Error() {
		workerName := name.String()
		workerName = workerName[1 : len(workerName)-1]

		workerSupervisor.workerStatusMapMutex.Lock()
		workerSupervisor.workerStatusMap[workerName] = WORKER_SLEEP
		workerSupervisor.workerStatusMapMutex.Unlock()
	}
	workerSupervisor.Log().Error("Actor %v terminated. Panic reason: %v", name, reason.Error())
	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleMessage(from gen.PID, message any) error {
	switch receivedMessage := message.(type) {
	case types.CompleteMessage:
		{
			workerSupervisor.workerStatusMapMutex.Lock()
			workerSupervisor.workerStatusMap[receivedMessage.WorkerName] = WORKER_AVAILABLE
			workerSupervisor.workerStatusMapMutex.Unlock()

			if workerSupervisor.multiProcess {
				workerSupervisor.complete++
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
	case types.DispatchCmdMessage:
		{
			workerSupervisor.dispatchCmd(receivedMessage)
			return nil
		}
	case types.DispatchCmdListMessage:
		{
			workerSupervisor.dispatchCmdList(receivedMessage)
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

func (workerSupervisor *WorkerSupervisor) dispatchCmd(message types.DispatchCmdMessage) {
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
				workerSupervisor.workerStatusMap[message.WorkerName] = WORKER_RUNNING
				workerSupervisor.Log().Info("Sending cmd=%v to %v", message.Cmd, message.WorkerName)
				workerSupervisor.Send(gen.Atom(message.WorkerName), types.DoProcessCall{Cmd: message.Cmd, WorkerReportChan: message.WorkerReportChan})
			}
		case WORKER_SLEEP:
			{
				workerSupervisor.Log().Warning("Actor %v is sleeping", message.WorkerName)
			}
		}
	} else {
		workerSupervisor.Log().Error("Actor %v is not valid", message.WorkerName)
	}
}

func (workerSupervisor *WorkerSupervisor) dispatchCmdList(message types.DispatchCmdListMessage) {
	go func() {
		workerSupervisor.multiProcess = true
		workerSupervisor.complete = 0

		workerReportChans := []chan types.WorkerReport{}
		for i := 0; i < len(message.CmdList); i++ {
			workerReportChans = append(workerReportChans, make(chan types.WorkerReport, 1))
		}

		startTime := time.Now()
		index := 0
		for index < len(message.CmdList) {
			workerSupervisor.workerStatusMapMutex.Lock()
			for workerName, status := range workerSupervisor.workerStatusMap {
				if status == WORKER_AVAILABLE {
					workerSupervisor.workerStatusMap[workerName] = WORKER_RUNNING
					workerSupervisor.Log().Info("Sending cmd=%v to %v", message.CmdList[index], workerName)
					workerSupervisor.Send(gen.Atom(workerName), types.DoProcessCall{Cmd: message.CmdList[index], WorkerReportChan: workerReportChans[index]})
					index++

					if index >= len(message.CmdList) {
						break
					}
				}
			}
			workerSupervisor.workerStatusMapMutex.Unlock()
		}

		workerReports := []types.WorkerReport{}
		for _, workerReportChan := range workerReportChans {
			workerReports = append(workerReports, <-workerReportChan)
		}

		endTime := time.Now()

		message.SummaryReportChan <- types.SummaryReport{
			Total:         len(message.CmdList),
			Complete:      workerSupervisor.complete,
			TotalTime:     fmt.Sprintf("%v", endTime.Sub(startTime)),
			WorkerReports: workerReports,
		}

		workerSupervisor.multiProcess = false
	}()
}

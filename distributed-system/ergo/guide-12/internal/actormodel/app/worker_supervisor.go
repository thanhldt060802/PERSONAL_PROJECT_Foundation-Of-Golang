package app

import (
	"fmt"
	"sync"
	"thanhldt060802/internal/actormodel/types"
	"thanhldt060802/internal/repository"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerSupervisor struct {
	act.Supervisor

	taskRepository         repository.TaskRepository
	numberOfInitialWorkers int

	availableWorkerMap      map[string]bool
	availableWorkerMapMutex sync.Mutex
}

func FactoryWorkerSupervisor() gen.ProcessBehavior {
	return &WorkerSupervisor{}
}

func (workerSupervisor *WorkerSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	workerSupervisor.taskRepository = args[0].(repository.TaskRepository)
	workerSupervisor.numberOfInitialWorkers = args[1].(int)
	workerSupervisor.availableWorkerMap = map[string]bool{}
	workerSupervisor.availableWorkerMapMutex = sync.Mutex{}

	supervisorSpec := act.SupervisorSpec{}
	supervisorSpec.EnableHandleChild = true
	supervisorSpec.DisableAutoShutdown = true
	supervisorSpec.Type = act.SupervisorTypeOneForOne
	supervisorSpec.Restart.Strategy = act.SupervisorStrategyTransient
	supervisorSpec.Restart.Intensity = 100
	supervisorSpec.Restart.Period = 5

	supervisorSpec.Children = []act.SupervisorChildSpec{}
	for i := 1; i <= workerSupervisor.numberOfInitialWorkers; i++ {
		supervisorSpec.Children = append(supervisorSpec.Children, act.SupervisorChildSpec{
			Name:    gen.Atom(fmt.Sprintf("worker_%v", i)),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args:    nil,
		})
	}

	workerSupervisor.Log().Info("Started worker supervisor %s %s on %s", workerSupervisor.PID(), workerSupervisor.Name(), workerSupervisor.Node().Name())
	workerSupervisor.Send(workerSupervisor.PID(), types.DoStart{})

	return supervisorSpec, nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildStart(childName gen.Atom, pid gen.PID) error {
	workerName := childName.String()
	workerName = workerName[1 : len(workerName)-1]

	workerSupervisor.availableWorkerMapMutex.Lock()
	defer workerSupervisor.availableWorkerMapMutex.Unlock()
	workerSupervisor.availableWorkerMap[workerName] = false

	workerSupervisor.Log().Info("Actor start with name %v and PID %v", childName, pid)

	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildTerminate(name gen.Atom, pid gen.PID, reason error) error {
	if reason.Error() == gen.TerminateReasonNormal.Error() {
		workerName := name.String()
		workerName = workerName[1 : len(workerName)-1]

		workerSupervisor.availableWorkerMapMutex.Lock()
		defer workerSupervisor.availableWorkerMapMutex.Unlock()
		workerSupervisor.availableWorkerMap[workerName] = true
	} else {
		workerSupervisor.Log().Error("Actor %v terminated. Panic reason: %v", name, reason.Error())
	}

	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleMessage(from gen.PID, message any) error {
	switch receivedMessage := message.(type) {
	case types.DoStart:
		{
			for _, supervisorChildSpec := range workerSupervisor.Children() {
				workerSupervisor.Send(supervisorChildSpec.Name, types.DoStart{})
			}

			return nil
		}
	case types.GetExistedWorkersMessage:
		{
			workerSupervisor.getExistedWorkers(receivedMessage)
			return nil
		}
	case types.RunTaskMessage:
		{
			workerSupervisor.runTask(receivedMessage)
			return nil
		}
	}

	return nil
}

func (workerSupervisor *WorkerSupervisor) getExistedWorkers(message types.GetExistedWorkersMessage) {
	workerNames := []string{}
	running := []string{}
	available := []string{}

	workerSupervisor.availableWorkerMapMutex.Lock()
	defer workerSupervisor.availableWorkerMapMutex.Unlock()
	for _, supervisorChildSpec := range workerSupervisor.Children() {
		workerName := supervisorChildSpec.Name.String()
		workerName = workerName[1 : len(workerName)-1]

		workerNames = append(workerNames, workerName)
		if workerSupervisor.availableWorkerMap[workerName] {
			available = append(available, workerName)
		} else {
			running = append(running, workerName)
		}
	}
	message.WorkerNamesChan <- workerNames
	message.RunningChan <- running
	message.AvailableChan <- available
}

func (workerSupervisor *WorkerSupervisor) runTask(message types.RunTaskMessage) {
	workerSupervisor.availableWorkerMapMutex.Lock()
	defer workerSupervisor.availableWorkerMapMutex.Unlock()
	for workerName := range workerSupervisor.availableWorkerMap {
		if workerSupervisor.availableWorkerMap[workerName] {
			workerSupervisor.Log().Info("Restart existed actor %v", workerName)
			if err := workerSupervisor.StartChild(gen.Atom(workerName), workerSupervisor.taskRepository, message.TaskId); err != nil {
				workerSupervisor.Log().Error("Restart existed actor %v failed: %v", workerName, err.Error())
			}

			workerSupervisor.Log().Info("Restart existed actor %v successful", workerName)

			return
		}
	}

	workerName := fmt.Sprintf("worker_%v", len(workerSupervisor.availableWorkerMap)+1)
	workerSupervisor.Log().Info("Start new actor %v", workerName)
	if err := workerSupervisor.AddChild(act.SupervisorChildSpec{
		Name:    gen.Atom(workerName),
		Factory: FactoryWorkerActor,
		Options: gen.ProcessOptions{},
		Args:    []any{workerSupervisor.taskRepository, message.TaskId},
	}); err != nil {
		workerSupervisor.Log().Info("Start new actor %v failed: %v", workerName, err.Error())
	}

	workerSupervisor.Log().Info("Start new actor %v successful", workerName)
}

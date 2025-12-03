package app

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerSupervisor struct {
	act.Supervisor
}

func FactoryWorkerSupervisor() gen.ProcessBehavior {
	return &WorkerSupervisor{}
}

func (workerSupervisor *WorkerSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	supervisorSpec := act.SupervisorSpec{}
	supervisorSpec.EnableHandleChild = true
	supervisorSpec.DisableAutoShutdown = true

	// supervisorSpec.Type = act.SupervisorTypeOneForOne
	// supervisorSpec.Restart.Strategy = act.SupervisorStrategyPermanent

	// supervisorSpec.Type = act.SupervisorTypeOneForOne
	// supervisorSpec.Restart.Strategy = act.SupervisorStrategyTemporary

	supervisorSpec.Type = act.SupervisorTypeOneForOne
	supervisorSpec.Restart.Strategy = act.SupervisorStrategyTransient

	// supervisorSpec.Type = act.SupervisorTypeAllForOne
	// supervisorSpec.Restart.Strategy = act.SupervisorStrategyTransient

	// supervisorSpec.Type = act.SupervisorTypeRestForOne
	// supervisorSpec.Restart.Strategy = act.SupervisorStrategyTransient

	supervisorSpec.Restart.Intensity = 100
	supervisorSpec.Restart.Period = 5

	supervisorSpec.Children = []act.SupervisorChildSpec{
		{
			Name:    gen.Atom("worker_1"),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args:    []any{3, false},
		},
		{
			Name:    gen.Atom("worker_2"),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args:    []any{6, false},
		},
		{
			Name:    gen.Atom("worker_3"),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args:    []any{9, true},
		},
		{
			Name:    gen.Atom("worker_4"),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args:    []any{12, false},
		},
		{
			Name:    gen.Atom("worker_5"),
			Factory: FactoryWorkerActor,
			Options: gen.ProcessOptions{},
			Args:    []any{15, false},
		},
	}

	workerSupervisor.Log().Info("Started worker supervisor %v %v on %v", workerSupervisor.PID(), workerSupervisor.Name(), workerSupervisor.Node().Name())
	return supervisorSpec, nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildStart(childName gen.Atom, pid gen.PID) error {
	workerSupervisor.Log().Info("Actor start with name %v and PID %v", childName, pid)
	return nil
}

func (workerSupervisor *WorkerSupervisor) HandleChildTerminate(name gen.Atom, pid gen.PID, reason error) error {
	workerSupervisor.Log().Error("Actor %v terminated. Panic reason: %v", name, reason.Error())
	return nil
}

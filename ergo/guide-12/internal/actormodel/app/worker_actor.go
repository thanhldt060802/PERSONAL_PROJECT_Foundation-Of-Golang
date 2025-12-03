package app

import (
	"math/rand"
	"thanhldt060802/internal/actormodel/types"
	"thanhldt060802/internal/model"
	"thanhldt060802/internal/repository"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerActor struct {
	act.Actor

	taskRepository repository.TaskRepository
	taskId         int64

	task *model.Task
}

func FactoryWorkerActor() gen.ProcessBehavior {
	return &WorkerActor{}
}

func (workerActor *WorkerActor) Init(args ...any) error {
	if args == nil {
		workerActor.Log().Info("Started worker %v %v on %v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name())
	} else {
		workerActor.taskRepository = args[0].(repository.TaskRepository)
		workerActor.taskId = args[1].(int64)
		workerActor.Log().Info("Started worker %v %v on %v with init value taskId=%v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name(), workerActor.taskId)

		workerActor.Send(workerActor.PID(), types.DoProcessTask{})
	}

	return nil
}

func (workerActor *WorkerActor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case types.DoStart:
		{
			return gen.TerminateReasonNormal
		}
	case types.DoProcessTask:
		{
			workerActor.Log().Info("RECEIVED task taskId=%v ...", workerActor.taskId)

			workerActor.task = workerActor.taskRepository.GetById(workerActor.taskId)
			workerActor.Log().Info("GET task taskId=%v from database successful", workerActor.task.Id)

			workerActor.Log().Info("UPDATE IN PROGRESS task taskId=%v on database successful", workerActor.task.Id)

			for workerActor.task.Progress < workerActor.task.Target {
				if rand.Intn(10) == 0 {
					return gen.ErrProcessTerminated
				}
				workerActor.task.Progress++
				workerActor.Log().Info("PRCOESSING task taskId=%v (%v/%v)", workerActor.task.Id, workerActor.task.Progress, workerActor.task.Target)
				time.Sleep(500 * time.Millisecond)
			}

			workerActor.Log().Info("UPDATE COMPLETED task taskId=%v on database successful", workerActor.task.Id)

			workerActor.Log().Info("COMPLETED task taskId=%v", workerActor.task.Id)

			return gen.TerminateReasonNormal
		}
	}

	return nil
}

func (workerActor *WorkerActor) Terminate(reason error) {
	if reason.Error() != gen.TerminateReasonNormal.Error() {
		workerActor.Log().Info("--- UPDATE CANCEL task taskId=%v on database successful", workerActor.task.Id)
	}
}

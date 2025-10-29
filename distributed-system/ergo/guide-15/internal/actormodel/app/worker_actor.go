package app

import (
	"context"
	"fmt"
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
		workerActor.Log().Info("Started worker %v %v on %v with init task taskId=%v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name(), workerActor.taskId)

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

			for {
				foundTask, err := workerActor.taskRepository.GetById(context.Background(), workerActor.taskId)
				if err != nil {
					workerActor.Log().Error("GET task taskId=%v from postgresql failed: %v", workerActor.taskId, err.Error())
					return fmt.Errorf("id of task is not valid: %v", err.Error())
				}
				if foundTask.Status != "IN PROGRESS" {
					workerActor.task = foundTask
					break
				}
				time.Sleep(1 * time.Second)
			}
			workerActor.Log().Info("GET task taskId=%v from postgresql successful", workerActor.task.Id)

			workerActor.task.Status = "IN PROGRESS"
			if err := workerActor.taskRepository.Update(context.Background(), workerActor.task); err != nil {
				workerActor.Log().Error("UPDATE IN PROGRESS task taskId=%v on postgresql failed: %v", workerActor.task.Id, err.Error())
				return fmt.Errorf("update task on postgresql failed: %v", err.Error())
			}
			workerActor.Log().Info("UPDATE IN PROGRESS task taskId=%v on postgresql successful", workerActor.task.Id)

			for workerActor.task.Progress < workerActor.task.Target {
				if rand.Intn(100) < workerActor.task.ErrorRate {
					return gen.ErrProcessTerminated
				}

				workerActor.task.Progress++
				workerActor.Log().Info("PROCESSING task taskId=%v (%v/%v)", workerActor.task.Id, workerActor.task.Progress, workerActor.task.Target)
				time.Sleep(1 * time.Second)
			}

			workerActor.task.Status = "COMPLETED"
			if err := workerActor.taskRepository.Update(context.Background(), workerActor.task); err != nil {
				workerActor.Log().Error("UPDATE COMPLETED task taskId=%v on postgresql failed: %v", workerActor.task.Id, err.Error())
				return fmt.Errorf("update task on postgresql failed: %v", err.Error())
			}
			workerActor.Log().Info("UPDATE COMPLETED task taskId=%v on postgresql successful", workerActor.task.Id)

			workerActor.Log().Info("COMPLETED task taskId=%v ...", workerActor.task.Id)
			return gen.TerminateReasonNormal
		}
	}

	return nil
}

func (workerActor *WorkerActor) Terminate(reason error) {
	if reason.Error() != gen.TerminateReasonNormal.Error() {
		workerActor.task.Status = "CANCEL"
		if err := workerActor.taskRepository.Update(context.Background(), workerActor.task); err != nil {
			workerActor.Log().Error("UPDATE CANCEL task taskId=%v on postgresql failed: %v", workerActor.task.Id, err.Error())
		}
		workerActor.Log().Info("UPDATE CANCEL task taskId=%v on postgresql successful", workerActor.task.Id)
	}
}

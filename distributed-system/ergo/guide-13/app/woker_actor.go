package app

import (
	"fmt"
	"thanhldt060802/repository"
	"thanhldt060802/types"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerActor struct {
	act.Actor

	delaySecond int
}

func FactoryWorkerActor() gen.ProcessBehavior {
	return &WorkerActor{}
}

func (workerActor *WorkerActor) Init(args ...any) error {
	workerActor.delaySecond = args[0].(int)
	workerActor.Log().Info("Started process %v %v on %v with init value delaySecond=%v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name(), workerActor.delaySecond)

	return nil
}

func (workerActor *WorkerActor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case types.Run:
		{
			repository.SharedDataSourceMutex.Lock()
			workerActor.Log().Info("ACCESS to shared data source")
			for i := 1; i <= workerActor.delaySecond; i++ {
				workerActor.Log().Info("PROCESSING (%v/%v)", i, workerActor.delaySecond)
				time.Sleep(1 * time.Second)
			}

			repository.SharedDataSource = append(repository.SharedDataSource, fmt.Sprintf("%v accessed", workerActor.Name()))
			fmt.Println(repository.SharedDataSource)
			repository.SharedDataSourceMutex.Unlock()

			workerActor.Log().Info("COMPLETED access to shared data source")

			return nil
		}
	}

	return fmt.Errorf("unknown message")
}

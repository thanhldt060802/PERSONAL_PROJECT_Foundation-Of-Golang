package app

import (
	"fmt"
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
	switch receivedMessage := message.(type) {
	case types.SimpleMessage:
		{
			workerActor.Log().Info("RECEIVED %v from %v", receivedMessage, from)
			for i := 1; i <= workerActor.delaySecond; i++ {
				workerActor.Log().Info("PROCESSING %v from %v (%v/%v)", receivedMessage, from, i, workerActor.delaySecond)
				time.Sleep(1 * time.Second)
			}
			workerActor.Log().Info("COMPLETED %v from %v", receivedMessage, from)

			return nil
		}
	}

	return fmt.Errorf("unknown message")
}

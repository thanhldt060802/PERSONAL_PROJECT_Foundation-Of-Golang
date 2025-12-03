package app

import (
	"fmt"
	"math/rand"
	"thanhldt060802/types"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerActor struct {
	act.Actor

	delaySecond int
	maybeError  bool
}

func FactoryWorkerActor() gen.ProcessBehavior {
	return &WorkerActor{}
}

func (workerActor *WorkerActor) Init(args ...any) error {
	workerActor.delaySecond = args[0].(int)
	workerActor.maybeError = args[1].(bool)
	workerActor.Log().Info("Started process %v %v on %v with init value delaySecond=%v, maybeError=%v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name(), workerActor.delaySecond, workerActor.maybeError)

	return nil
}

func (workerActor *WorkerActor) HandleMessage(from gen.PID, message any) error {
	switch receivedMessage := message.(type) {
	case types.SimpleMessage:
		{
			workerActor.Log().Info("RECEIVED %v from %v", receivedMessage, from)
			for i := 1; i <= workerActor.delaySecond; i++ {
				if workerActor.maybeError {
					if rand.Intn(3) == 0 {
						return gen.ErrProcessTerminated
					}
				}
				workerActor.Log().Info("PROCESSING %v from %v (%v/%v)", receivedMessage, from, i, workerActor.delaySecond)
				time.Sleep(1 * time.Second)
			}
			workerActor.Log().Info("COMPLETED %v from %v", receivedMessage, from)

			return nil
			// return gen.TerminateReasonNormal // Sử dụng return này để Terminate Actor với signal bình thường (dùng với SupervisorStrategyTransient sẽ không được restart, với SupervisorStrategyPermanent sẽ được restart)
		}
	}

	return fmt.Errorf("unknown message")
}

func (workerActor *WorkerActor) Terminate(reason error) {
	workerActor.Log().Warning("CANCEL by reason: %v", reason.Error())
}

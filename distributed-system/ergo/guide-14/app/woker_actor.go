package app

import (
	"fmt"
	"thanhldt060802/types"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerActor struct {
	act.Actor
}

func FactoryWorkerActor() gen.ProcessBehavior {
	return &WorkerActor{}
}

func (workerActor *WorkerActor) Init(args ...any) error {
	workerActor.Log().Info("Started process %v %v on %v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name())
	return nil
}

func (workerActor *WorkerActor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case types.Run:
		{
			eventChan := make(chan string)
			requestMessage := types.RequestMessage{
				EventChan: eventChan,
			}

			workerActor.Log().Info("SEND request to Event Dispatcher")
			workerActor.Send("event_dispatcher", requestMessage)
			for {
				event := <-eventChan
				if event == "completed" {
					break
				} else {
					workerActor.Log().Info("RECEIVED event from Event Dispatcher: %v", event)
				}
			}

			workerActor.Log().Info("COMPLETED request to Event Dispatcher")

			return nil
		}
	}

	return fmt.Errorf("unknown message")
}

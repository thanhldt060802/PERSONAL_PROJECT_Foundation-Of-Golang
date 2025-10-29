package app

import (
	"fmt"
	"math/rand"
	"thanhldt060802/types"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

var eventList []string = []string{
	"event_1",
	"event_2",
	"event_3",
	"event_4",
	"event_5",
}

type EventDispatcherActor struct {
	act.Actor
}

func FactoryEventDispatcherActor() gen.ProcessBehavior {
	return &EventDispatcherActor{}
}

func (eventDispatcherActor *EventDispatcherActor) Init(args ...any) error {
	eventDispatcherActor.Log().Info("Started process %s %s on %s", eventDispatcherActor.PID(), eventDispatcherActor.Name(), eventDispatcherActor.Node().Name())
	return nil
}

func (eventDispatcherActor *EventDispatcherActor) HandleMessage(from gen.PID, message any) error {
	switch receivedMessage := message.(type) {
	case types.RequestMessage:
		{
			eventDispatcherActor.Log().Info("ACCEPT request from %v", from)
			go func() {
				for i := 1; i <= 3; i++ {
					time.Sleep(1 * time.Second)
					event := eventList[rand.Intn(len(eventList))]
					eventDispatcherActor.Log().Info("SEND EVENT to %v: %v", from, event)
					receivedMessage.EventChan <- event
				}
				receivedMessage.EventChan <- "completed"

				eventDispatcherActor.Log().Info("COMPLETED request from %v", from)
			}()

			return nil
		}
	}

	return fmt.Errorf("unknown message")
}

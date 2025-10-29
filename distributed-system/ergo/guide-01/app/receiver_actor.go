package app

import (
	"fmt"
	"math/rand"
	"thanhldt060802/types"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type ReceiverActor struct {
	act.Actor
}

func FactoryReceiverActor() gen.ProcessBehavior {
	return &ReceiverActor{}
}

func (receiverActor *ReceiverActor) Init(args ...any) error {
	receiverActor.Log().Info("Started process %v %v on %v", receiverActor.PID(), receiverActor.Name(), receiverActor.Node().Name())
	return nil
}

func (receiverActor *ReceiverActor) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch receivedMessage := request.(type) {
	case types.SimpleMessage:
		{
			delayTime := time.Duration(rand.Intn(int(3000*time.Millisecond-1000*time.Millisecond))) + 1000*time.Millisecond

			receiverActor.Log().Info("RECEIVED %v from %v", receivedMessage, from)
			receiverActor.Log().Info("PROCESSING %v from %v", receivedMessage, from)
			time.Sleep(delayTime)
			receiverActor.Log().Info("COMPLETED %v from %v in %v", receivedMessage, from, delayTime)

			return fmt.Sprintf("%v completed in %v", receivedMessage.Data, delayTime), nil
		}
	}

	return nil, fmt.Errorf("unknown message")
}

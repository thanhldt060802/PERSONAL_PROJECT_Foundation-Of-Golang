package app

import (
	"fmt"
	"thanhldt060802/types"

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
			receiverActor.Log().Info("RECEIVED %v from %v", receivedMessage, from)
			receiverActor.Log().Info("PROCESSING %v from %v", receivedMessage, from)
			receiverActor.Log().Info("COMPLETED %v from %v", receivedMessage, from)

			return fmt.Sprintf("%v completed", receivedMessage.Data), nil
		}
	}

	return nil, fmt.Errorf("unknown message")
}

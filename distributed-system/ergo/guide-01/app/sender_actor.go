package app

import (
	"fmt"
	"thanhldt060802/types"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type SenderActor struct {
	act.Actor

	count int
}

func FactorySenderActor() gen.ProcessBehavior {
	return &SenderActor{}
}

func (senderActor *SenderActor) Init(args ...any) error {
	senderActor.count = 1
	senderActor.Log().Info("Started process %v %v on %v", senderActor.PID(), senderActor.Name(), senderActor.Node().Name())

	return nil
}

func (senderActor *SenderActor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case types.Run:
		{
			sendingMessage := types.SimpleMessage{Data: fmt.Sprintf("Data %v", senderActor.count)}
			senderActor.Log().Info("SENDING %v to receiver", sendingMessage)
			resp, err := senderActor.Call(gen.Atom("receiver"), sendingMessage)
			if err != nil {
				senderActor.Log().Error("ERROR from receiver: %v", err.Error())
			} else {
				senderActor.Log().Info("RECEIVED %v from receiver", resp)
				senderActor.count++
			}

			return nil
		}
	}

	return fmt.Errorf("unknown message")
}

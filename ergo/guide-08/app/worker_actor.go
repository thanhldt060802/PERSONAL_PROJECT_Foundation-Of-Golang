package app

import (
	"fmt"
	"strings"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
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
	switch receivedMessage := message.(type) {
	case meta.MessageTCPConnect:
		workerActor.Log().Info("--- New connection with: %v (serving meta-process: %v)", receivedMessage.RemoteAddr, receivedMessage.ID)
		return nil
	case meta.MessageTCPDisconnect:
		workerActor.Log().Info("--- Terminated connection (serving meta-process: %v)", receivedMessage.ID)
		return nil
	case meta.MessageTCP:
		data := string(receivedMessage.Data)
		workerActor.Log().Info("<-- Got TCP packet from %v: %v ", receivedMessage.ID, strings.TrimRight(data, "\r\n"))
		receivedMessage.Data = []byte("OK: " + data)
		if err := workerActor.SendAlias(receivedMessage.ID, receivedMessage); err != nil {
			workerActor.Log().Error("--- Send to %v failed: %v", receivedMessage.ID, err.Error())
		}

		return nil
	}

	return fmt.Errorf("unknown message")
}

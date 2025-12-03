package app

import (
	"fmt"
	"strings"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/meta/websocket"
)

type WorkerActor struct {
	act.Actor
}

func FactoryWorkerActor() gen.ProcessBehavior {
	return &WorkerActor{}
}

func (workerActor *WorkerActor) Init(args ...any) error {
	workerActor.Log().Info("Started process successful")
	return nil
}

func (workerActor *WorkerActor) HandleMessage(from gen.PID, message any) error {
	switch m := message.(type) {

	case websocket.MessageConnect:
		workerActor.Log().Info("%s new websocket connection with %s, meta-process %s", workerActor.Name(), m.RemoteAddr.String(), m.ID)
		reply := websocket.Message{
			Body: []byte("hello from " + workerActor.PID().String()),
		}
		workerActor.SendAlias(m.ID, reply)

	case websocket.MessageDisconnect:
		workerActor.Log().Info("%s disconnected with %s", workerActor.Name(), m.ID)

	case websocket.Message:
		received := string(m.Body)
		strip := strings.TrimRight(received, "\r\n")
		workerActor.Log().Info("%s got message (meta-process: %s): %s", workerActor.Name(), m.ID, strip)
		reply := fmt.Sprintf("OK %s", strip)
		m.Body = []byte(reply)
		workerActor.SendAlias(m.ID, m)

	default:
		workerActor.Log().Error("uknown message from %s %#v", from, message)
	}
	return nil
}

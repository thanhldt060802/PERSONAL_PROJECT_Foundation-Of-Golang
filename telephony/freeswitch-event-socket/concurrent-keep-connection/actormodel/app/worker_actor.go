package app

import (
	"fmt"
	"thanhldt060802/actormodel/types"
	"thanhldt060802/esl"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"github.com/luandnh/eslgo"
)

type WorkerActor struct {
	act.Actor

	eslConn *eslgo.Conn

	workerReportChan chan types.WorkerReport
	workerReport     types.WorkerReport
}

func FactoryWorkerActor() gen.ProcessBehavior {
	return &WorkerActor{}
}

func (workerActor *WorkerActor) Init(args ...any) error {
	eslConfig := args[0].(esl.ESLConfig)
	eslConn, err := eslConfig.Connect()
	if err != nil {
		workerActor.Log().Info("Connect to ESL failed: %s", err.Error())
		return gen.ErrProcessTerminated
	}
	workerActor.Log().Info("Connect to ESL successful")
	workerActor.eslConn = eslConn
	workerActor.Log().Info("Started worker %v %v on %v", workerActor.PID(), workerActor.Name(), workerActor.Node().Name())

	return nil
}

func (workerActor *WorkerActor) HandleMessage(from gen.PID, message any) error {
	switch receivedMessage := message.(type) {
	case types.DoProcessCall:
		{
			workerActor.workerReportChan = receivedMessage.WorkerReportChan
			workerActor.workerReport = types.WorkerReport{}

			workerName := workerActor.Name().String()
			workerName = workerName[1 : len(workerName)-1]

			workerActor.workerReport.WorkerName = workerName
			workerActor.workerReport.Cmd = receivedMessage.Cmd

			workerActor.Log().Info("--- Call api(%v) ...", receivedMessage.Cmd)

			startTime := time.Now()
			if _, err := esl.API(workerActor.eslConn, receivedMessage.Cmd); err != nil {
				return gen.ErrProcessTerminated
			}
			time.Sleep(40 * time.Millisecond)
			endTime := time.Now()

			workerActor.Log().Info("--- Received response from api(%v) ...", receivedMessage.Cmd)

			workerActor.workerReport.DelayTime = fmt.Sprintf("%v", endTime.Sub(startTime))

			workerActor.workerReportChan <- workerActor.workerReport

			workerActor.Send(from, types.CompleteMessage{WorkerName: workerName})
			return nil
		}
	case types.DoSafetyTerminate:
		{
			return gen.TerminateReasonNormal
		}
	}

	return nil
}

func (workerActor *WorkerActor) Terminate(reason error) {
	workerActor.workerReportChan <- workerActor.workerReport
	workerActor.eslConn.Close()
}

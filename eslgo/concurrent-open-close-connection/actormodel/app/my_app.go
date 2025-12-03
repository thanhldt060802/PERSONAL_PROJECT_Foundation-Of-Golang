package app

import (
	"fmt"
	"thanhldt060802/esl"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
)

type ActorModel struct {
	nodeName              string
	eslConfig             esl.ESLConfig
	numberOfInitialWorker int

	node          gen.Node
	supervisorPID gen.PID
}

func New(nodeName string, eslConfig esl.ESLConfig, numberOfInitialWorker int) *ActorModel {
	return &ActorModel{
		nodeName:              nodeName,
		eslConfig:             eslConfig,
		numberOfInitialWorker: numberOfInitialWorker,
	}
}

func (actorModel *ActorModel) Start() {
	var nodeOptions gen.NodeOptions

	optionColored := colored.Options{TimeFormat: time.DateTime, IncludeName: true}
	loggerColored, err := colored.CreateLogger(optionColored)
	if err != nil {
		panic(err)
	}

	nodeOptions.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{}),
	}
	nodeOptions.Log.DefaultLogger.Disable = true
	nodeOptions.Log.Loggers = append(nodeOptions.Log.Loggers, gen.Logger{Name: actorModel.nodeName, Logger: loggerColored})

	node, err := ergo.StartNode(gen.Atom(fmt.Sprintf("%v@localhost", actorModel.nodeName)), nodeOptions)
	if err != nil {
		panic(err)
	}
	actorModel.node = node

	supervisorPID, _ := actorModel.node.SpawnRegister(gen.Atom("worker_supervisor"), FactoryWorkerSupervisor, gen.ProcessOptions{}, actorModel.eslConfig, actorModel.numberOfInitialWorker)
	actorModel.supervisorPID = supervisorPID
}

func (actorModel *ActorModel) Node() gen.Node {
	return actorModel.node
}

func (actorModel *ActorModel) SupervisorPID() gen.PID {
	return actorModel.supervisorPID
}

package app

import (
	"fmt"
	"thanhldt060802/internal/actormodel/types"
	"thanhldt060802/internal/repository"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
)

type ActorModel struct {
	taskRepository repository.TaskRepository

	nodeName               string
	numberOfInitialWorkers int

	node          gen.Node
	supervisorPID gen.PID
}

func New(taskRepository repository.TaskRepository, nodeName string, numberOfInitialWorkers int) *ActorModel {
	return &ActorModel{
		taskRepository:         taskRepository,
		nodeName:               nodeName,
		numberOfInitialWorkers: numberOfInitialWorkers,
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

	supervisorPID, _ := actorModel.node.Spawn(FactoryWorkerSupervisor, gen.ProcessOptions{}, actorModel.taskRepository, actorModel.numberOfInitialWorkers)
	actorModel.supervisorPID = supervisorPID

	actorModel.node.Send(supervisorPID, types.DoStart{})
}

func (actorModel *ActorModel) Node() gen.Node {
	return actorModel.node
}

func (actorModel *ActorModel) SupervisorPID() gen.PID {
	return actorModel.supervisorPID
}

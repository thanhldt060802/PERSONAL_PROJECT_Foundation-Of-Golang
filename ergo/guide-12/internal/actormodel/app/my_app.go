package app

import (
	"fmt"
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
	optionColored := colored.Options{TimeFormat: time.DateTime, IncludeName: true}
	loggerColored, err := colored.CreateLogger(optionColored)
	if err != nil {
		panic(err)
	}

	var myNodeOptions gen.NodeOptions
	myNodeOptions.Log.DefaultLogger.Disable = true
	myNodeOptions.Log.Loggers = []gen.Logger{
		{Name: "mynode", Logger: loggerColored},
	}
	myNodeOptions.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{}),
	}

	myNode, err := ergo.StartNode(gen.Atom(fmt.Sprintf("%v@localhost", actorModel.nodeName)), myNodeOptions)
	if err != nil {
		panic(err)
	}
	actorModel.node = myNode

	supervisorPID, _ := actorModel.node.SpawnRegister(gen.Atom("worker_supervisor"), FactoryWorkerSupervisor, gen.ProcessOptions{}, actorModel.taskRepository, actorModel.numberOfInitialWorkers)
	actorModel.supervisorPID = supervisorPID
}

func (actorModel *ActorModel) Node() gen.Node {
	return actorModel.node
}

func (actorModel *ActorModel) SupervisorPID() gen.PID {
	return actorModel.supervisorPID
}

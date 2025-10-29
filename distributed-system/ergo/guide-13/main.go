package main

import (
	"fmt"
	"math/rand"
	"sync"
	"thanhldt060802/app"
	"thanhldt060802/repository"
	"thanhldt060802/types"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
)

func main() {

	rand.New(rand.NewSource(time.Now().UnixNano()))

	repository.SharedDataSource = []string{}
	repository.SharedDataSourceMutex = sync.Mutex{}

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

	myNode, err := ergo.StartNode(gen.Atom("mynode@localhost"), myNodeOptions)
	if err != nil {
		panic(err)
	}

	myNode.SpawnRegister(gen.Atom("worker_1"), app.FactoryWorkerActor, gen.ProcessOptions{}, 5)
	myNode.SpawnRegister(gen.Atom("worker_2"), app.FactoryWorkerActor, gen.ProcessOptions{}, 3)
	myNode.SpawnRegister(gen.Atom("worker_3"), app.FactoryWorkerActor, gen.ProcessOptions{}, 4)
	myNode.SpawnRegister(gen.Atom("worker_4"), app.FactoryWorkerActor, gen.ProcessOptions{}, 7)
	myNode.SpawnRegister(gen.Atom("worker_5"), app.FactoryWorkerActor, gen.ProcessOptions{}, 6)

	fmt.Println()
	fmt.Println()
	time.Sleep(2 * time.Second)

	myNode.Send(gen.Atom("worker_1"), types.Run{})
	myNode.Send(gen.Atom("worker_2"), types.Run{})
	myNode.Send(gen.Atom("worker_3"), types.Run{})
	myNode.Send(gen.Atom("worker_4"), types.Run{})
	myNode.Send(gen.Atom("worker_5"), types.Run{})

	select {}

}

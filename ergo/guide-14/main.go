package main

import (
	"fmt"
	"math/rand"
	"thanhldt060802/app"
	"thanhldt060802/types"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
)

func main() {

	rand.New(rand.NewSource(time.Now().UnixNano()))

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

	myNode.SpawnRegister(gen.Atom("event_dispatcher"), app.FactoryEventDispatcherActor, gen.ProcessOptions{})
	myNode.SpawnRegister(gen.Atom("worker_1"), app.FactoryWorkerActor, gen.ProcessOptions{})
	myNode.SpawnRegister(gen.Atom("worker_2"), app.FactoryWorkerActor, gen.ProcessOptions{})
	myNode.SpawnRegister(gen.Atom("worker_3"), app.FactoryWorkerActor, gen.ProcessOptions{})

	fmt.Println()
	fmt.Println()
	time.Sleep(2 * time.Second)

	myNode.Send(gen.Atom("worker_1"), types.Run{})
	myNode.Send(gen.Atom("worker_2"), types.Run{})
	myNode.Send(gen.Atom("worker_3"), types.Run{})

	select {}

}

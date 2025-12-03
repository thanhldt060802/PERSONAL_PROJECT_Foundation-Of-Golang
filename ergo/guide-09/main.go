package main

import (
	"thanhldt060802/app"
	"time"

	"ergo.services/application/observer"
	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/logger/colored"
)

func main() {

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

	myNode.SpawnRegister("my_web_worker", app.FactoryMyWebWorker, gen.ProcessOptions{})
	myNode.SpawnRegister("my_web", app.FactoryMyWeb, gen.ProcessOptions{})

	select {}

}

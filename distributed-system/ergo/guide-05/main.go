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

	var node1Options gen.NodeOptions
	node1Options.Log.DefaultLogger.Disable = true
	node1Options.Log.Loggers = []gen.Logger{
		{Name: "node1", Logger: loggerColored},
	}
	node1Options.Network.Cookie = "123"
	node1Options.Applications = []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{}),
	}

	node1, err := ergo.StartNode(gen.Atom("node1@localhost"), node1Options)
	if err != nil {
		panic(err)
	}

	node1.SpawnRegister(gen.Atom("sender_1"), app.FactorySenderActor, gen.ProcessOptions{})
	node1.SpawnRegister(gen.Atom("sender_2"), app.FactorySenderActor, gen.ProcessOptions{})
	node1.SpawnRegister(gen.Atom("receiver"), app.FactoryReceiverActor, gen.ProcessOptions{})

	var node2Options gen.NodeOptions
	node2Options.Log.DefaultLogger.Disable = true
	node2Options.Log.Loggers = []gen.Logger{
		{Name: "node2", Logger: loggerColored},
	}
	node2Options.Network.Cookie = "123"

	node2, err := ergo.StartNode(gen.Atom("node2@localhost"), node2Options)
	if err != nil {
		panic(err)
	}

	node2.SpawnRegister(gen.Atom("receiver"), app.FactoryReceiverActor, gen.ProcessOptions{})

	fmt.Println()
	fmt.Println()
	time.Sleep(2 * time.Second)

	node1.Send(gen.Atom("sender_1"), types.RunLocal{})
	node1.Send(gen.Atom("sender_2"), types.RunLocal{})

	select {}

}

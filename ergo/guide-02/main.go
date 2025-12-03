package main

import (
	"fmt"
	"math/rand"
	"thanhldt060802/app"
	"thanhldt060802/types"
	"time"

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

	mynode, err := ergo.StartNode(gen.Atom("mynode@localhost"), myNodeOptions)
	if err != nil {
		panic(err)
	}

	mynode.SpawnRegister(gen.Atom("sender_1"), app.FactorySenderActor, gen.ProcessOptions{})
	mynode.SpawnRegister(gen.Atom("sender_2"), app.FactorySenderActor, gen.ProcessOptions{})
	mynode.SpawnRegister(gen.Atom("receiver"), app.FactoryReceiverActor, gen.ProcessOptions{})

	fmt.Println()
	fmt.Println()
	time.Sleep(2 * time.Second)

	mynode.Send(gen.Atom("sender_1"), types.Run{})
	mynode.Send(gen.Atom("sender_2"), types.Run{})
	mynode.Send(gen.Atom("sender_1"), types.Run{})
	mynode.Send(gen.Atom("sender_2"), types.Run{})

	select {}

}

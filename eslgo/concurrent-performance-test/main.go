package main

import (
	"thanhldt060802/esl"
	"thanhldt060802/testautocall"
)

func main() {

	eslConfig := esl.ESLConfig{
		Address:  "103.72.97.156",
		Port:     7021,
		Password: "TEL4VN.COM",
	}

	testCall := &testautocall.TestCall{
		ESLConfig:             &eslConfig,
		NumberOfThread:        5,
		NumberOfCallPerThread: 500,
		Command:               "show calls",
		MinDelayTimeMillis:    10,
		MaxDelayTimeMillis:    20,
	}
	testCall.Start()

	select {}

}

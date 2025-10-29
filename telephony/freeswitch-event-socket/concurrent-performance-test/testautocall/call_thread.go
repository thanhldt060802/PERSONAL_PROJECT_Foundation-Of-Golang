package testautocall

import (
	"fmt"
	"math/rand"
	"thanhldt060802/esl"
	"time"
)

type CallThread struct {
	name               string
	eslConfig          *esl.ESLConfig
	numberOfCall       int
	command            string
	minDelayTimeMillis int
	maxDelayTimeMillis int
}

func (callThread *CallThread) Run() {
	defer CallReportManagerWG.Done()

	callReport := CallReport{
		numberOfCall:         callThread.numberOfCall,
		numberOfCallComplete: 0,
		totalTime:            0,
		avgTime:              0,
	}

	fmt.Printf("Call thread %v start\n", callThread.name)
	callThreadStartTime := time.Now()
	for i := 1; i <= callThread.numberOfCall; i++ {
		fmt.Printf("Connection %v-%v\n", callThread.name, i)
		conn, err := callThread.eslConfig.Connect()
		if err != nil {
			fmt.Printf("Connection %v-%v failed: %v\n", callThread.name, i, err.Error())
			continue
		}
		fmt.Printf("Connection %v-%v call api(%v)\n", callThread.name, i, callThread.command)
		if _, err := esl.API(conn, callThread.command); err != nil {
			fmt.Printf("Connection %v-%v call api(%v) failed: %s\n", callThread.name, i, callThread.command, err.Error())
		} else {
			fmt.Printf("Connection %v-%v received response from api(%v)\n", callThread.name, i, callThread.command)
			callReport.numberOfCallComplete++
		}

		delay := rand.Intn(callThread.maxDelayTimeMillis-callThread.minDelayTimeMillis+1) + callThread.minDelayTimeMillis
		time.Sleep(time.Duration(delay) * time.Millisecond)

		conn.Close()
	}
	totalTime := time.Since(callThreadStartTime)
	fmt.Printf("Call thread %v end, total time: %v\n", callThread.name, totalTime)

	callReport.totalTime = totalTime
	callReport.avgTime = totalTime / time.Duration(callThread.numberOfCall)

	CallReportChanManager[callThread.name] <- callReport
}

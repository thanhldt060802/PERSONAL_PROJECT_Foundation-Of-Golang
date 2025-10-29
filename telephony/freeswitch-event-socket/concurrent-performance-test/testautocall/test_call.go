package testautocall

import (
	"fmt"
	"sync"
	"thanhldt060802/esl"
	"time"
)

type TestCall struct {
	ESLConfig             *esl.ESLConfig
	NumberOfThread        int
	NumberOfCallPerThread int
	Command               string
	MinDelayTimeMillis    int
	MaxDelayTimeMillis    int

	callThreads []*CallThread
}

func (testCall *TestCall) Start() {
	CallReportChanManager = map[string]chan CallReport{}
	CallReportManagerWG = &sync.WaitGroup{}
	testCall.callThreads = []*CallThread{}

	for i := 1; i <= testCall.NumberOfThread; i++ {
		testCall.callThreads = append(testCall.callThreads, &CallThread{
			name:               fmt.Sprintf("A%v", i),
			eslConfig:          testCall.ESLConfig,
			numberOfCall:       testCall.NumberOfCallPerThread,
			command:            testCall.Command,
			minDelayTimeMillis: testCall.MinDelayTimeMillis,
			maxDelayTimeMillis: testCall.MaxDelayTimeMillis,
		})
		CallReportChanManager[fmt.Sprintf("A%v", i)] = make(chan CallReport, 1)
	}
	CallReportManagerWG.Add(testCall.NumberOfThread)

	for _, callThread := range testCall.callThreads {
		go callThread.Run()
	}

	CallReportManagerWG.Wait()

	time.Sleep(1 * time.Second)
	fmt.Println()
	fmt.Println("Call report:")
	for callThreadName, callReportChan := range CallReportChanManager {
		callReport := <-callReportChan
		fmt.Println("-------------------------")
		fmt.Printf("Call thread name: %v\n", callThreadName)
		callReport.Show()
	}
	fmt.Println()
}

package testautocall

import (
	"fmt"
	"sync"
	"time"
)

type CallReport struct {
	numberOfCall         int
	numberOfCallComplete int
	totalTime            time.Duration
	avgTime              time.Duration
}

func (callReport *CallReport) Show() {
	fmt.Printf("Number of call: %v\n", callReport.numberOfCall)
	fmt.Printf("Number of call complete: %v\n", callReport.numberOfCallComplete)
	fmt.Printf("Total time: %v\n", callReport.totalTime)
	fmt.Printf("Avg time: %v\n", callReport.avgTime)
}

var CallReportChanManager map[string]chan CallReport
var CallReportManagerWG *sync.WaitGroup

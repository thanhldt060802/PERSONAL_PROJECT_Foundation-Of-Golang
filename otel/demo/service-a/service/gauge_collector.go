package service

import (
	"runtime"
	"thanhldt060802/common/constant"
	"thanhldt060802/internal"
	"time"
)

func StartGaugeCollector() {
	go func() {
		for {
			internal.Observer.RecordGauge(constant.CPU_USAGE, 0.5*float64(runtime.NumGoroutine()), nil)
			time.Sleep(1 * time.Second)
		}
	}()
}

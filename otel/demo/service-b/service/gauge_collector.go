package service

import (
	"runtime"
	"thanhldt060802/common/constant"
	"thanhldt060802/internal/lib/otel"
	"time"
)

func StartGaugeCollector() {
	go func() {
		for {
			otel.RecordGauge(constant.CPU_USAGE_PERCENT, 0.5*float64(runtime.NumGoroutine()), nil)
			time.Sleep(1 * time.Second)
		}
	}()
}

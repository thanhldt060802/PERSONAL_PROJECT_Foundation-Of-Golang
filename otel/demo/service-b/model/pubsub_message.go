package model

import "thanhldt060802/internal/lib/otel"

type ExamplePubSubMessage struct {
	otel.TraceCarrier `json:"trace_carrier"`

	ExampleUuid string `json:"example_uuid"`
}

package repository

import (
	"context"
	"thanhldt060802/model"
)

type IExampleRepo interface {
	GetById(ctx context.Context, exampleUuid string) (*model.Example, error)
}

var ExampleRepo IExampleRepo

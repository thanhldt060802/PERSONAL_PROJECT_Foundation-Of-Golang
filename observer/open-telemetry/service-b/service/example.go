package service

import (
	"context"
	"thanhldt060802/common/observer"
	"thanhldt060802/model"
	"thanhldt060802/repository"
)

type (
	IExampleService interface {
		GetById(ctx context.Context, exampleUuid string) (*model.Example, error)
	}
	ExampleService struct {
	}
)

func NewExampleService() IExampleService {
	return &ExampleService{}
}

func (s *ExampleService) GetById(ctx context.Context, exampleUuid string) (*model.Example, error) {
	ctx, span := observer.StartSpanInternal(ctx)
	defer span.End()

	example, err := repository.ExampleRepo.GetById(ctx, exampleUuid)
	if err != nil {
		span.Err = err
		return nil, err
	}

	return example, nil
}

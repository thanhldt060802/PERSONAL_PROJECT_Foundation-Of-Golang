package service

import (
	"context"
	"thanhldt060802/common/apperror"
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
		return nil, apperror.ErrServiceUnavailable(err, "Failed to get example")
	} else if example == nil {
		return nil, apperror.ErrNotFound("Example example_uuid='"+exampleUuid+"' not found", "ERR_EXAMPLE_NOT_FOUND")
	}

	return example, nil
}

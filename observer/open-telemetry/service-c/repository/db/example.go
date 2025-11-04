package db

import (
	"context"
	"database/sql"
	"fmt"
	"thanhldt060802/common/observer"
	"thanhldt060802/internal/sqlclient"
	"thanhldt060802/model"
	"thanhldt060802/repository"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ExampleRepo struct {
}

func NewExampleRepo() repository.IExampleRepo {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	repo := &ExampleRepo{}
	repo.DeleteTable(ctx)
	repo.InitTable(ctx)
	repo.GenerateData(ctx)

	return repo
}

func (repo *ExampleRepo) DeleteTable(ctx context.Context) {
	if err := repository.DropTable(sqlclient.SqlClientConnInstance, ctx, (*model.Example)(nil)); err != nil {
		panic(err)
	}
}

func (repo *ExampleRepo) InitTable(ctx context.Context) {
	if err := repository.CreateTable(sqlclient.SqlClientConnInstance, ctx, (*model.Example)(nil)); err != nil {
		panic(err)
	}
}

func (repo *ExampleRepo) GenerateData(ctx context.Context) {
	if err := sqlclient.SqlClientConnInstance.GetDB().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		for i := 1; i <= 30; i++ {
			user := model.Example{
				ExampleUuid: uuid.New().String(),
				Name:        fmt.Sprintf("Example %v", i),
				Description: fmt.Sprintf("Description %v", i),
			}
			if _, err := tx.NewInsert().Model(&user).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (repo *ExampleRepo) GetById(ctx context.Context, exampleUuid string) (*model.Example, error) {
	ctx, span := observer.StartSpanInternal(ctx)
	defer span.End()

	example := new(model.Example)

	query := sqlclient.SqlClientConnInstance.GetDB().NewSelect().Model(example).
		Where("example_uuid = ?", exampleUuid)

	span.AddEvent("Start query", trace.WithAttributes(
		attribute.String("sql", query.String()),
	))

	err := query.Scan(ctx)
	if err == sql.ErrNoRows {
		span.SetAttributes(
			attribute.String("data", "null"),
		)
		return nil, nil
	} else if err != nil {
		span.Err = err
		return nil, err
	} else {
		span.SetAttributes(
			attribute.String("data", fmt.Sprintf(`{ "example_uuid": "%v"}`, example.ExampleUuid)),
		)
		return example, nil
	}
}

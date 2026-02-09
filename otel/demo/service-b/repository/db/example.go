package db

import (
	"context"
	"database/sql"
	"fmt"
	"thanhldt060802/internal"
	"thanhldt060802/internal/sqlclient"
	"thanhldt060802/model"
	"thanhldt060802/repository"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ExampleRepo struct {
}

func NewExampleRepo() repository.IExampleRepo {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()

	repo := &ExampleRepo{}
	// repo.DeleteTable(ctx)
	// repo.InitTable(ctx)
	// repo.GenerateData(ctx)

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
	ctx, span := internal.Observer.NewSpan(ctx, "GetExampleById-Repository")
	defer span.Done()

	internal.Observer.InfoLogWithCtx(ctx, "[Repository layer] Get Example by example_uuid='%s'", exampleUuid)

	example := new(model.Example)

	query := sqlclient.SqlClientConnInstance.GetDB().NewSelect().Model(example).
		Where("example_uuid = ?", exampleUuid)

	span.AddEvent("Execute SQL", map[string]any{
		"sql": query.String(),
	})

	err := query.Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		internal.Observer.ErrorLogWithCtx(ctx, "[Repository layer] Failed to get Example by example_uuid='%s'", exampleUuid)
		span.SetError(err)
		return nil, err
	} else {
		return example, nil
	}
}

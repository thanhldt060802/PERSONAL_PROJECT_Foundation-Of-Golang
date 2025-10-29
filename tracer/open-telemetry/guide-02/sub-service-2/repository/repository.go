package repository

import (
	"context"
	"strings"
	"thanhldt060802/internal/sqlclient"

	"github.com/uptrace/bun/schema"
)

func CreateTable(client sqlclient.ISqlClientConn, ctx context.Context, table interface{}) error {
	query := client.GetDB().NewCreateTable().Model(table).IfNotExists()
	value, _ := query.AppendQuery(schema.NewFormatter(query.Dialect()), nil)
	queryStr := string(value)
	queryStr = strings.ReplaceAll(queryStr, " char(36)", " uuid")
	queryStr = strings.ReplaceAll(queryStr, " timestamp", " timestamptz")
	queryStr = strings.ReplaceAll(queryStr, " timestamptz_only", " timestamp")

	_, err := client.GetDB().QueryContext(ctx, queryStr)
	return err
}

func DropTable(client sqlclient.ISqlClientConn, ctx context.Context, table interface{}) error {
	_, err := client.GetDB().NewDropTable().Model(table).IfExists().Exec(ctx)
	return err
}

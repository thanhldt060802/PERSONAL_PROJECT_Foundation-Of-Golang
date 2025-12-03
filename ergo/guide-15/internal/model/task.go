package model

import (
	"github.com/uptrace/bun"
)

type Task struct {
	bun.BaseModel `bun:"table:tasks"`

	Id        int64  `bun:"id,pk,autoincrement"`
	Progress  int    `bun:"progress,notnull"`
	Target    int    `bun:"target,notnull"`
	Status    string `bun:"status,notnull"`
	ErrorRate int    `bun:"error_rate,notnull"`
}

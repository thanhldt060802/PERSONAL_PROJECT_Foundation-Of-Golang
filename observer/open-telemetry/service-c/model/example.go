package model

import "github.com/uptrace/bun"

type Example struct {
	bun.BaseModel `json:"-" bun:"tb_example"`

	ExampleUuid string `json:"example_uuid" bun:"example_uuid,pk,type:uuid"`
	Name        string `json:"name" bun:"name,type:varchar(100),notnull"`
	Description string `json:"description" bun:"description,type:text,notnull"`
}

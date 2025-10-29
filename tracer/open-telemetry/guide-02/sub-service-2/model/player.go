package model

import "github.com/uptrace/bun"

type Player struct {
	bun.BaseModel `json:"-" bun:"tb_player"`

	PlayerUuid string `json:"player_uuid" bun:"player_uuid,pk,type:uuid"`
	Name       string `json:"name" bun:"name,type:varchar(100),notnull"`
	Class      string `json:"class" bun:"class,type:varchar(50),notnull"`
	Level      int    `json:"level" bun:"level,type:integer,notnull"`
}

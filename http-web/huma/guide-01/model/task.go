package model

import (
	"time"
)

type Task struct {
	Id          string    `json:"task_uuid" gorm:"column:task_uuid;primaryKey;type:uuid"`
	Password    *string   `json:"-" gorm:"column:password;type:text;default:NULL"`
	TaskName    string    `json:"task_name" gorm:"column:password;type:varchar(100);not null"`
	Description *string   `json:"description" gorm:"column:description;type:text;default:NULL"`
	State       string    `json:"state" gorm:"column:state;type:varchar(100);not null"`
	Priority    string    `json:"priority" gorm:"column:priority;type:varchar(50);not null"`
	Progress    int       `json:"progress" gorm:"column:progress;type:integer;not null"`
	CreatedBy   string    `json:"created_by" gorm:"column:created_by;type:uuid;not null"`
	UpdatedBy   *string   `json:"updated_by" gorm:"column:updated_by;type:uuid;default:NULL"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
}

type TaskView struct {
	Id          string    `json:"task_uuid" gorm:"column:task_uuid;primaryKey;type:uuid"`
	TaskName    string    `json:"task_name" gorm:"column:password;type:varchar(100);not null"`
	Description *string   `json:"description" gorm:"column:description;type:text;default:NULL"`
	State       string    `json:"state" gorm:"column:state;type:varchar(100);not null"`
	Priority    string    `json:"priority" gorm:"column:priority;type:varchar(50);not null"`
	Progress    int       `json:"progress" gorm:"column:progress;type:integer;not null"`
	CreatedBy   string    `json:"created_by" gorm:"column:created_by;type:uuid;not null"`
	UpdatedBy   *string   `json:"updated_by" gorm:"column:updated_by;type:uuid;default:NULL"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
}

// Mapping với table được tào từ struct gốc
func (TaskView) TableName() string {
	return "tasks"
}

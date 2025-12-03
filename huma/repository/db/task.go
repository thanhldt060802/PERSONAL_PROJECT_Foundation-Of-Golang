package db

import (
	"context"
	"fmt"
	"math/rand"
	"thanhldt060802/common/util"
	"thanhldt060802/dtos"
	"thanhldt060802/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepo struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) *TaskRepo {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	repo := &TaskRepo{
		db: db,
	}
	repo.DeleteTable(ctx)
	repo.InitTable(ctx)
	repo.GenerateData(ctx)

	return repo
}

func (repo *TaskRepo) DeleteTable(ctx context.Context) {
	if err := repo.db.Migrator().DropTable(&model.Task{}); err != nil {
		panic(err)
	}
}

func (repo *TaskRepo) InitTable(ctx context.Context) {
	if err := repo.db.AutoMigrate(&model.Task{}); err != nil {
		panic(err)
	}
}

func (repo *TaskRepo) GenerateData(ctx context.Context) {
	states := []string{"todo", "in progress", "done"}
	priorities := []string{"low", "medium", "high"}

	if err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 1; i <= 30; i++ {
			task := model.Task{}
			task.Id = uuid.New().String()
			password := fmt.Sprintf("pass_%v", i)
			task.Password = &password
			task.TaskName = fmt.Sprintf("Task %v", i)
			description := fmt.Sprintf("Description %v", i)
			task.Description = &description
			task.State = states[rand.Intn(len(states))]
			task.Priority = priorities[rand.Intn(len(priorities))]
			task.Progress = rand.Intn(101)
			task.CreatedBy = uuid.New().String()
			updatedBy := uuid.New().String()
			task.UpdatedBy = &updatedBy
			if err := tx.Create(&task).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func (repo *TaskRepo) GetsView(ctx context.Context, filter *dtos.GetsTaskFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error) {
	var tasks []*model.TaskView

	query := repo.db.WithContext(ctx).Model(&model.TaskView{})
	query = util.BuildQuery(query, filter, &model.TaskView{})
	query = query.Limit(limit).Offset(offset)
	query = query.Order(util.GetOrderExpr(sorts, "created_at desc"))

	results := query.Find(&tasks)
	if results.Error != nil {
		return nil, 0, results.Error
	} else {
		totalRecord := results.RowsAffected
		return tasks, int(totalRecord), nil
	}
}

func (repo *TaskRepo) GetsViewCustom(ctx context.Context, filter *dtos.GetsTaskCustomFilter, limit, offset int, sorts []string) ([]*model.TaskView, int, error) {
	var tasks []*model.TaskView

	query := repo.db.WithContext(ctx).Model(&model.TaskView{})
	query = util.BuildQuery(query, filter, &model.TaskView{})
	query = query.Limit(limit).Offset(offset)
	query = query.Order(util.GetOrderExpr(sorts, "created_at desc"))

	results := query.Find(&tasks)
	if results.Error != nil {
		return nil, 0, results.Error
	} else {
		totalRecord := results.RowsAffected
		return tasks, int(totalRecord), nil
	}
}

func (repo *TaskRepo) GetViewById(ctx context.Context, id uuid.UUID) (*model.TaskView, error) {
	task := new(model.TaskView)

	query := repo.db.WithContext(ctx).Model(&model.TaskView{}).
		Where("task_uuid = ?", id)

	err := query.First(task).Error
	if err != nil {
		return nil, err
	} else {
		return task, nil
	}
}

func (repo *TaskRepo) GetById(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	task := new(model.Task)

	query := repo.db.WithContext(ctx).Model(&model.Task{}).
		Where("task_uuid = ?", id)

	err := query.First(task).Error
	if err != nil {
		return nil, err
	} else {
		return task, nil
	}
}

func (repo *TaskRepo) Create(ctx context.Context, task *model.Task) error {
	return repo.db.WithContext(ctx).Create(task).Error
}

func (repo *TaskRepo) UpdateById(ctx context.Context, id uuid.UUID, task *model.Task) error {
	return repo.db.WithContext(ctx).Where("task_uuid = ?", id).Save(task).Error
}

func (repo *TaskRepo) PatchById(ctx context.Context, id uuid.UUID, task *model.Task) error {
	return repo.db.WithContext(ctx).Where("task_uuid = ?", id).Updates(task).Error
}

func (repo *TaskRepo) DeleteById(ctx context.Context, id uuid.UUID) error {
	return repo.db.WithContext(ctx).Delete(&model.Task{}, id).Error
}

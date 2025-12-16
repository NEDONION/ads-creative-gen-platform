package repository

import (
	"context"
	"fmt"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
)

// TaskRepository 定义 copywriting 模块需要的任务操作
type TaskRepository interface {
	Create(ctx context.Context, task *models.CreativeTask) error
	GetByUUID(ctx context.Context, uuid string) (*models.CreativeTask, error)
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
}

type gormTaskRepo struct{}

// NewTaskRepository 默认实现
func NewTaskRepository() TaskRepository {
	return &gormTaskRepo{}
}

func (r *gormTaskRepo) Create(ctx context.Context, task *models.CreativeTask) error {
	return database.DB.WithContext(ctx).Create(task).Error
}

func (r *gormTaskRepo) GetByUUID(ctx context.Context, uuid string) (*models.CreativeTask, error) {
	var task models.CreativeTask
	if err := database.DB.WithContext(ctx).Where("uuid = ?", uuid).First(&task).Error; err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return &task, nil
}

func (r *gormTaskRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return database.DB.WithContext(ctx).Model(&models.CreativeTask{}).Where("id = ?", id).Updates(fields).Error
}

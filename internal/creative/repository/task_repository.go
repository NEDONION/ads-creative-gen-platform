package repository

import (
	"context"
	"fmt"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"
	"ads-creative-gen-platform/internal/shared"

	"gorm.io/gorm"
)

// taskRepository 任务仓储实现
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(db *gorm.DB) ports.TaskRepository {
	return &taskRepository{db: db}
}

// Create 创建任务
func (r *taskRepository) Create(ctx context.Context, task *models.CreativeTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取任务
func (r *taskRepository) GetByID(ctx context.Context, id uint) (*models.CreativeTask, error) {
	var task models.CreativeTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUUID 根据UUID获取任务
func (r *taskRepository) GetByUUID(ctx context.Context, uuid string) (*models.CreativeTask, error) {
	var task models.CreativeTask
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUUIDWithAssets 根据UUID获取任务及素材
func (r *taskRepository) GetByUUIDWithAssets(ctx context.Context, uuid string) (*models.CreativeTask, error) {
	var task models.CreativeTask
	if err := r.db.WithContext(ctx).Preload("Assets").Where("uuid = ?", uuid).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateStatus 更新任务状态
func (r *taskRepository) UpdateStatus(ctx context.Context, id uint, status models.TaskStatus, progress int) error {
	return r.db.WithContext(ctx).Model(&models.CreativeTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   status,
			"progress": progress,
		}).Error
}

// UpdateProgress 更新任务进度
func (r *taskRepository) UpdateProgress(ctx context.Context, id uint, progress int) error {
	return r.db.WithContext(ctx).Model(&models.CreativeTask{}).
		Where("id = ?", id).
		Update("progress", progress).Error
}

// UpdateFields 更新任务字段
func (r *taskRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.CreativeTask{}).
		Where("id = ?", id).
		Updates(fields).Error
}

// List 查询任务列表
func (r *taskRepository) List(ctx context.Context, query shared.ListTasksQuery) ([]models.CreativeTask, int64, error) {
	var tasks []models.CreativeTask
	var total int64

	// 构建查询
	dbQuery := r.db.WithContext(ctx).Model(&models.CreativeTask{}).Preload("Assets")

	// 应用筛选条件
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	} else {
		// 默认不展示草稿
		dbQuery = dbQuery.Where("status <> ?", models.TaskDraft)
	}
	if query.UserID > 0 {
		dbQuery = dbQuery.Where("user_id = ?", query.UserID)
	}

	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count tasks failed: %w", err)
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := dbQuery.Offset(offset).Limit(query.PageSize).
		Order("created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("list tasks failed: %w", err)
	}

	return tasks, total, nil
}

// Delete 删除任务
func (r *taskRepository) Delete(ctx context.Context, task *models.CreativeTask) error {
	return r.db.WithContext(ctx).Delete(task).Error
}

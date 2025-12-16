package repository

import (
	"context"
	"fmt"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"
	"ads-creative-gen-platform/internal/shared"

	"gorm.io/gorm"
)

// assetRepository 素材仓储实现
type assetRepository struct {
	db *gorm.DB
}

// NewAssetRepository 创建素材仓储
func NewAssetRepository(db *gorm.DB) ports.AssetRepository {
	return &assetRepository{db: db}
}

// Create 创建素材
func (r *assetRepository) Create(ctx context.Context, asset *models.CreativeAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// List 查询素材列表
func (r *assetRepository) List(ctx context.Context, query shared.ListAssetsQuery) ([]models.CreativeAsset, int64, error) {
	var assets []models.CreativeAsset
	var total int64

	// 构建查询
	dbQuery := r.db.WithContext(ctx).Model(&models.CreativeAsset{}).Preload("Task")

	// 应用筛选条件
	if query.Format != "" {
		dbQuery = dbQuery.Where("format = ?", query.Format)
	}
	if query.TaskID != "" {
		dbQuery = dbQuery.Joins("JOIN creative_tasks ON creative_assets.task_id = creative_tasks.id").
			Where("creative_tasks.uuid = ?", query.TaskID)
	}

	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count assets failed: %w", err)
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := dbQuery.Offset(offset).Limit(query.PageSize).
		Order("creative_assets.created_at DESC").
		Find(&assets).Error; err != nil {
		return nil, 0, fmt.Errorf("list assets failed: %w", err)
	}

	return assets, total, nil
}

// DeleteByTaskID 根据任务删除素材
func (r *assetRepository) DeleteByTaskID(ctx context.Context, taskID uint) error {
	return r.db.WithContext(ctx).Where("task_id = ?", taskID).Delete(&models.CreativeAsset{}).Error
}

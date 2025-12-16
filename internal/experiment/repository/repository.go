package repository

import (
	"fmt"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
)

// ExperimentRepository 定义实验相关的数据访问
type ExperimentRepository interface {
	ListExperiments(status string, page, pageSize int) ([]models.Experiment, int64, error)
	CreateExperiment(exp *models.Experiment) error
	CreateVariants(variants []models.ExperimentVariant) error
	FindAssetByID(id uint) (*models.CreativeAsset, error)
	FindAssetByUUID(uuid string) (*models.CreativeAsset, error)
	FindAssetWithTaskByID(id uint) (*models.CreativeAsset, error)
	GetExperimentByUUID(uuid string) (*models.Experiment, error)
	GetExperimentWithVariants(uuid string) (*models.Experiment, error)
	UpdateExperimentFields(uuid string, fields map[string]interface{}) error
	GetMetric(expID uint, creativeID uint) (*models.ExperimentMetric, error)
	SaveMetric(metric *models.ExperimentMetric) error
	ListMetrics(expID uint) ([]models.ExperimentMetric, error)
}

type gormExperimentRepo struct{}

// NewExperimentRepository 默认实现
func NewExperimentRepository() ExperimentRepository {
	return &gormExperimentRepo{}
}

func (r *gormExperimentRepo) ListExperiments(status string, page, pageSize int) ([]models.Experiment, int64, error) {
	db := database.DB.Model(&models.Experiment{})
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count experiments failed: %w", err)
	}

	var experiments []models.Experiment
	if err := db.Preload("Variants").
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&experiments).Error; err != nil {
		return nil, 0, fmt.Errorf("list experiments failed: %w", err)
	}
	return experiments, total, nil
}

func (r *gormExperimentRepo) CreateExperiment(exp *models.Experiment) error {
	return database.DB.Create(exp).Error
}

func (r *gormExperimentRepo) CreateVariants(variants []models.ExperimentVariant) error {
	return database.DB.Create(&variants).Error
}

func (r *gormExperimentRepo) FindAssetByID(id uint) (*models.CreativeAsset, error) {
	var asset models.CreativeAsset
	if err := database.DB.Where("id = ?", id).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *gormExperimentRepo) FindAssetByUUID(uuid string) (*models.CreativeAsset, error) {
	var asset models.CreativeAsset
	if err := database.DB.Where("uuid = ?", uuid).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *gormExperimentRepo) FindAssetWithTaskByID(id uint) (*models.CreativeAsset, error) {
	var asset models.CreativeAsset
	if err := database.DB.Preload("Task").Where("id = ?", id).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *gormExperimentRepo) GetExperimentByUUID(uuid string) (*models.Experiment, error) {
	var exp models.Experiment
	if err := database.DB.Where("uuid = ?", uuid).First(&exp).Error; err != nil {
		return nil, err
	}
	return &exp, nil
}

func (r *gormExperimentRepo) GetExperimentWithVariants(uuid string) (*models.Experiment, error) {
	var exp models.Experiment
	if err := database.DB.Preload("Variants").Where("uuid = ?", uuid).First(&exp).Error; err != nil {
		return nil, err
	}
	return &exp, nil
}

func (r *gormExperimentRepo) UpdateExperimentFields(uuid string, fields map[string]interface{}) error {
	return database.DB.Model(&models.Experiment{}).Where("uuid = ?", uuid).Updates(fields).Error
}

func (r *gormExperimentRepo) GetMetric(expID uint, creativeID uint) (*models.ExperimentMetric, error) {
	var metric models.ExperimentMetric
	if err := database.DB.Where("experiment_id = ? AND creative_id = ?", expID, creativeID).First(&metric).Error; err != nil {
		return nil, err
	}
	return &metric, nil
}

func (r *gormExperimentRepo) SaveMetric(metric *models.ExperimentMetric) error {
	if metric.ID == 0 {
		return database.DB.Create(metric).Error
	}
	return database.DB.Save(metric).Error
}

func (r *gormExperimentRepo) ListMetrics(expID uint) ([]models.ExperimentMetric, error) {
	var metrics []models.ExperimentMetric
	if err := database.DB.Where("experiment_id = ?", expID).Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}

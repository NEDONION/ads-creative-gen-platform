package repository

import (
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
	"fmt"
)

type TraceRepository interface {
	List(page, pageSize int, status, modelName, traceID, productName string) ([]models.ModelTrace, int64, error)
	Detail(traceID string) (*models.ModelTrace, error)
	CreateTrace(trace *models.ModelTrace) error
	UpdateTrace(traceID string, updates map[string]interface{}) error
	AddStep(step *models.ModelTraceStep) error
}

type gormTraceRepository struct{}

func NewTraceRepository() TraceRepository {
	return &gormTraceRepository{}
}

func (r *gormTraceRepository) List(page, pageSize int, status, modelName, traceID, productName string) ([]models.ModelTrace, int64, error) {
	query := database.DB.Model(&models.ModelTrace{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if modelName != "" {
		query = query.Where("model_name = ?", modelName)
	}
	if traceID != "" {
		query = query.Where("trace_id = ?", traceID)
	}
	if productName != "" {
		query = query.Where("product_name = ?", productName)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count traces failed: %w", err)
	}

	var traces []models.ModelTrace
	if err := query.Order("start_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&traces).Error; err != nil {
		return nil, 0, fmt.Errorf("list traces failed: %w", err)
	}
	return traces, total, nil
}

func (r *gormTraceRepository) Detail(traceID string) (*models.ModelTrace, error) {
	var trace models.ModelTrace
	if err := database.DB.Where("trace_id = ?", traceID).First(&trace).Error; err != nil {
		return nil, fmt.Errorf("trace not found: %w", err)
	}
	var steps []models.ModelTraceStep
	if err := database.DB.Where("trace_id = ?", traceID).Order("start_at asc").Find(&steps).Error; err != nil {
		return nil, fmt.Errorf("load steps failed: %w", err)
	}
	trace.Steps = steps
	return &trace, nil
}

func (r *gormTraceRepository) CreateTrace(trace *models.ModelTrace) error {
	return database.DB.Create(trace).Error
}

func (r *gormTraceRepository) UpdateTrace(traceID string, updates map[string]interface{}) error {
	return database.DB.Model(&models.ModelTrace{}).Where("trace_id = ?", traceID).Updates(updates).Error
}

func (r *gormTraceRepository) AddStep(step *models.ModelTraceStep) error {
	return database.DB.Create(step).Error
}

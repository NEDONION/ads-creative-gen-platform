package repository

import (
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
	"fmt"
	"time"
)

type TraceRepository interface {
	List(page, pageSize int, status, modelName, traceID, productName string) ([]models.ModelTrace, int64, error)
	Detail(traceID string) (*models.ModelTrace, error)
	CreateTrace(trace *models.ModelTrace) error
	UpdateTrace(traceID string, updates map[string]interface{}) error
	AddStep(step *models.ModelTraceStep) error
	RecoverStuckRunning(maxAge time.Duration, markMessage string) (int64, error)
	FailRunningBySource(source, markMessage string) (int64, error)
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

// RecoverStuckRunning 将超过 maxAge 仍 running 的 trace 标记为 failed
func (r *gormTraceRepository) RecoverStuckRunning(maxAge time.Duration, markMessage string) (int64, error) {
	if maxAge <= 0 {
		return 0, nil
	}
	cutoff := time.Now().Add(-maxAge)

	var stuck []models.ModelTrace
	if err := database.DB.Where("status = ? AND start_at < ?", "running", cutoff).Find(&stuck).Error; err != nil {
		return 0, fmt.Errorf("query stuck traces failed: %w", err)
	}

	var affected int64
	now := time.Now()
	for _, t := range stuck {
		duration := int(now.Sub(t.StartAt).Milliseconds())
		updates := map[string]interface{}{
			"status":        "failed",
			"end_at":        now,
			"duration_ms":   duration,
			"error_message": markMessage,
			"updated_at":    now,
		}
		if err := database.DB.Model(&models.ModelTrace{}).Where("id = ?", t.ID).Updates(updates).Error; err != nil {
			return affected, fmt.Errorf("update trace %s failed: %w", t.TraceID, err)
		}
		affected++
	}
	return affected, nil
}

// FailRunningBySource 将指定 source 的 running trace 标记失败
func (r *gormTraceRepository) FailRunningBySource(source, markMessage string) (int64, error) {
	if source == "" {
		return 0, nil
	}
	var traces []models.ModelTrace
	if err := database.DB.Where("status = ? AND source = ?", "running", source).Find(&traces).Error; err != nil {
		return 0, fmt.Errorf("query running traces by source failed: %w", err)
	}
	var affected int64
	now := time.Now()
	for _, t := range traces {
		duration := int(now.Sub(t.StartAt).Milliseconds())
		updates := map[string]interface{}{
			"status":        "failed",
			"end_at":        now,
			"duration_ms":   duration,
			"error_message": markMessage,
			"updated_at":    now,
		}
		if err := database.DB.Model(&models.ModelTrace{}).Where("id = ?", t.ID).Updates(updates).Error; err != nil {
			return affected, fmt.Errorf("update trace %s failed: %w", t.TraceID, err)
		}
		affected++
	}
	return affected, nil
}

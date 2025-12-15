package tracing

import (
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TraceListResult struct {
	Traces   []models.ModelTrace
	Total    int64
	Page     int
	PageSize int
}

type TraceService struct{}

func NewTraceService() *TraceService {
	return &TraceService{}
}

// List traces with filters
func (s *TraceService) List(page, pageSize int, status, modelName, traceID, productName string) (*TraceListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

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
		return nil, fmt.Errorf("count traces failed: %w", err)
	}

	var traces []models.ModelTrace
	if err := query.Order("start_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&traces).Error; err != nil {
		return nil, fmt.Errorf("list traces failed: %w", err)
	}

	return &TraceListResult{
		Traces:   traces,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Detail returns trace with steps
func (s *TraceService) Detail(traceID string) (*models.ModelTrace, error) {
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

// StartTrace 创建主 trace
func (s *TraceService) StartTrace(modelName, modelVersion, source, inputPreview, productName string) (string, error) {
	traceID := uuid.New().String()
	mt := models.ModelTrace{
		TraceID:       traceID,
		ModelName:     modelName,
		ModelVersion:  modelVersion,
		ProductName:   productName,
		Status:        "running",
		StartAt:       time.Now(),
		Source:        source,
		InputPreview:  inputPreview,
		OutputPreview: "",
	}
	if err := database.DB.Create(&mt).Error; err != nil {
		return "", fmt.Errorf("create trace failed: %w", err)
	}
	return traceID, nil
}

// FinishTrace 更新状态与耗时
func (s *TraceService) FinishTrace(traceID, status, outputPreview, errorMessage string) error {
	var trace models.ModelTrace
	if err := database.DB.Where("trace_id = ?", traceID).First(&trace).Error; err != nil {
		return fmt.Errorf("trace not found: %w", err)
	}
	now := time.Now()
	duration := int(now.Sub(trace.StartAt).Milliseconds())
	update := map[string]interface{}{
		"status":         status,
		"end_at":         now,
		"duration_ms":    duration,
		"output_preview": outputPreview,
		"error_message":  errorMessage,
		"updated_at":     now,
	}
	return database.DB.Model(&models.ModelTrace{}).Where("trace_id = ?", traceID).Updates(update).Error
}

// AddStep 写入步骤
func (s *TraceService) AddStep(traceID, stepName, component, status, inputPreview, outputPreview, errorMessage string, startAt, endAt time.Time) error {
	step := models.ModelTraceStep{
		TraceID:       traceID,
		StepName:      stepName,
		Component:     component,
		Status:        status,
		DurationMs:    int(endAt.Sub(startAt).Milliseconds()),
		StartAt:       startAt,
		EndAt:         endAt,
		InputPreview:  inputPreview,
		OutputPreview: outputPreview,
		ErrorMessage:  errorMessage,
	}
	return database.DB.Create(&step).Error
}

package tracing

import (
	"fmt"
	"time"

	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/infra/cache"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/tracing/repository"

	"github.com/google/uuid"
)

type TraceListResult struct {
	Traces   []models.ModelTrace
	Total    int64
	Page     int
	PageSize int
}

type TraceService struct {
	repo repository.TraceRepository
}

func NewTraceService() *TraceService {
	cfg := config.CacheConfig
	ttl := time.Minute
	if cfg != nil && cfg.DefaultTTL > 0 {
		ttl = cfg.DefaultTTL
	}
	return &TraceService{
		repo: repository.NewCachedTraceRepository(
			repository.NewTraceRepository(),
			cache.NewConfiguredCache(cfg),
			ttl,
			cfg != nil && cfg.DisableTracing,
		),
	}
}

// List traces with filters
func (s *TraceService) List(page, pageSize int, status, modelName, traceID, productName string) (*TraceListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

	traces, total, err := s.repo.List(page, pageSize, status, modelName, traceID, productName)
	if err != nil {
		return nil, err
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
	return s.repo.Detail(traceID)
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
	if err := s.repo.CreateTrace(&mt); err != nil {
		return "", fmt.Errorf("create trace failed: %w", err)
	}
	return traceID, nil
}

// FinishTrace 更新状态与耗时
func (s *TraceService) FinishTrace(traceID, status, outputPreview, errorMessage string) error {
	trace, err := s.repo.Detail(traceID)
	if err != nil || trace == nil {
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
	if err := s.repo.UpdateTrace(traceID, update); err != nil {
		return err
	}
	return nil
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
	return s.repo.AddStep(&step)
}

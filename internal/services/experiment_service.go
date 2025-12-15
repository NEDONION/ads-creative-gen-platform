package services

import (
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// ExperimentService 实验服务
type ExperimentService struct{}

func NewExperimentService() *ExperimentService {
	return &ExperimentService{}
}

// CreateExperimentInput 创建实验输入
type CreateExperimentInput struct {
	Name        string                   `json:"name"`
	ProductName string                   `json:"product_name,omitempty"`
	Variants    []ExperimentVariantInput `json:"variants"`
}

type ExperimentVariantInput struct {
	CreativeID string  `json:"creative_id"` // uuid 或 数字字符串
	Weight     float64 `json:"weight"`      // 0-1
}

// AssignedVariant 分流结果
type AssignedVariant struct {
	Variant models.ExperimentVariant
	Asset   *models.CreativeAsset
}

// ListExperimentsResult 实验列表结果
type ListExperimentsResult struct {
	Experiments []models.Experiment
	Total       int64
	Page        int
	PageSize    int
}

// ListExperiments 获取实验列表
func (s *ExperimentService) ListExperiments(page int, pageSize int, status string) (*ListExperimentsResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

	query := database.DB.Model(&models.Experiment{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count experiments failed: %w", err)
	}

	var experiments []models.Experiment
	if err := query.Preload("Variants").
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&experiments).Error; err != nil {
		return nil, fmt.Errorf("list experiments failed: %w", err)
	}

	return &ListExperimentsResult{
		Experiments: experiments,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// CreateExperiment 创建实验并计算桶
func (s *ExperimentService) CreateExperiment(input CreateExperimentInput) (*models.Experiment, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if len(input.Variants) < 2 {
		return nil, errors.New("need at least 2 variants")
	}
	totalWeight := 0.0
	for _, v := range input.Variants {
		if v.Weight <= 0 {
			return nil, errors.New("variant weight must be > 0")
		}
		totalWeight += v.Weight
	}
	if totalWeight <= 0 {
		return nil, errors.New("total weight invalid")
	}

	exp := models.Experiment{
		UUIDModel:   models.UUIDModel{UUID: uuid.New().String()},
		Name:        input.Name,
		ProductName: input.ProductName,
		Status:      models.ExpDraft,
	}

	if err := database.DB.Create(&exp).Error; err != nil {
		return nil, fmt.Errorf("create experiment failed: %w", err)
	}

	// 计算桶
	acc := 0
	var variants []models.ExperimentVariant
	for _, v := range input.Variants {
		if v.CreativeID == "" {
			return nil, errors.New("creative_id required")
		}
		var asset models.CreativeAsset
		var creativeNumericID uint
		// 尝试数字ID
		if parsed, err := strconv.ParseUint(v.CreativeID, 10, 64); err == nil {
			creativeNumericID = uint(parsed)
			_ = database.DB.Where("id = ?", creativeNumericID).First(&asset).Error
		} else {
			// 尝试 UUID
			if err := database.DB.Where("uuid = ?", v.CreativeID).First(&asset).Error; err == nil {
				creativeNumericID = asset.ID
			}
		}
		if creativeNumericID == 0 {
			return nil, fmt.Errorf("creative_id %s not found", v.CreativeID)
		}

		width := int(math.Round((v.Weight / totalWeight) * 10000))
		if width <= 0 {
			width = 1
		}
		start := acc
		end := acc + width - 1
		if end > 9999 {
			end = 9999
		}
		acc = end + 1

		variants = append(variants, models.ExperimentVariant{
			ExperimentID:  exp.ID,
			CreativeID:    creativeNumericID,
			Weight:        v.Weight,
			BucketStart:   start,
			BucketEnd:     end,
			Title:         asset.Title,
			ProductName:   asset.ProductName,
			ImageURL:      asset.PublicURL,
			CTAText:       asset.CTAText,
			SellingPoints: asset.SellingPoints,
		})
	}
	// 调整最后一个覆盖到 9999
	if len(variants) > 0 {
		variants[len(variants)-1].BucketEnd = 9999
	}

	if err := database.DB.Create(&variants).Error; err != nil {
		return nil, fmt.Errorf("create variants failed: %w", err)
	}

	exp.Variants = variants
	return &exp, nil
}

// UpdateStatus 更新实验状态
func (s *ExperimentService) UpdateStatus(expUUID string, status models.ExperimentStatus) error {
	if status != models.ExpActive && status != models.ExpPaused && status != models.ExpArchived && status != models.ExpDraft {
		return errors.New("invalid status")
	}
	update := map[string]interface{}{
		"status": status,
	}
	now := time.Now()
	if status == models.ExpActive {
		update["start_at"] = now
	}
	if status == models.ExpArchived {
		update["end_at"] = now
	}
	return database.DB.Model(&models.Experiment{}).Where("uuid = ?", expUUID).Updates(update).Error
}

// Assign 分流（返回变体与创意信息）
func (s *ExperimentService) Assign(expUUID string, userKey string) (*AssignedVariant, error) {
	var exp models.Experiment
	if err := database.DB.Preload("Variants").Where("uuid = ?", expUUID).First(&exp).Error; err != nil {
		return nil, fmt.Errorf("experiment not found: %w", err)
	}
	if exp.Status != models.ExpActive {
		return nil, errors.New("experiment not active")
	}

	bucket := randBucket(userKey)
	for _, v := range exp.Variants {
		if bucket >= v.BucketStart && bucket <= v.BucketEnd {
			var asset models.CreativeAsset
			if err := database.DB.Preload("Task").Where("id = ?", v.CreativeID).First(&asset).Error; err != nil {
				// 如果找不到资产，仍返回分流结果，只是不带素材信息
				return &AssignedVariant{Variant: v, Asset: nil}, nil
			}
			return &AssignedVariant{Variant: v, Asset: &asset}, nil
		}
	}
	return nil, errors.New("no variant matched")
}

// Hit 上报曝光
func (s *ExperimentService) Hit(expUUID string, creativeID uint) error {
	return s.incMetric(expUUID, creativeID, true)
}

// Click 上报点击
func (s *ExperimentService) Click(expUUID string, creativeID uint) error {
	return s.incMetric(expUUID, creativeID, false)
}

func (s *ExperimentService) incMetric(expUUID string, creativeID uint, isImpression bool) error {
	var exp models.Experiment
	if err := database.DB.Where("uuid = ?", expUUID).First(&exp).Error; err != nil {
		return fmt.Errorf("experiment not found: %w", err)
	}

	var metric models.ExperimentMetric
	if err := database.DB.Where("experiment_id = ? AND creative_id = ?", exp.ID, creativeID).First(&metric).Error; err != nil {
		metric = models.ExperimentMetric{
			ExperimentID: exp.ID,
			CreativeID:   creativeID,
			Impressions:  0,
			Clicks:       0,
		}
		if err := database.DB.Create(&metric).Error; err != nil {
			return err
		}
	}

	update := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if isImpression {
		update["impressions"] = metric.Impressions + 1
	} else {
		update["clicks"] = metric.Clicks + 1
	}
	if err := database.DB.Model(&metric).Updates(update).Error; err != nil {
		return err
	}
	return nil
}

// Metrics 查询实验指标
type ExperimentMetricsDTO struct {
	ExperimentID string `json:"experiment_id"`
	Variants     []struct {
		CreativeID  uint    `json:"creative_id"`
		Impressions int64   `json:"impressions"`
		Clicks      int64   `json:"clicks"`
		CTR         float64 `json:"ctr"`
	} `json:"variants"`
}

func (s *ExperimentService) GetMetrics(expUUID string) (*ExperimentMetricsDTO, error) {
	var exp models.Experiment
	if err := database.DB.Where("uuid = ?", expUUID).First(&exp).Error; err != nil {
		return nil, fmt.Errorf("experiment not found: %w", err)
	}

	var metrics []models.ExperimentMetric
	if err := database.DB.Where("experiment_id = ?", exp.ID).Find(&metrics).Error; err != nil {
		return nil, err
	}

	dto := ExperimentMetricsDTO{ExperimentID: expUUID}
	for _, m := range metrics {
		ctr := 0.0
		if m.Impressions > 0 {
			ctr = float64(m.Clicks) / float64(m.Impressions)
		}
		dto.Variants = append(dto.Variants, struct {
			CreativeID  uint    `json:"creative_id"`
			Impressions int64   `json:"impressions"`
			Clicks      int64   `json:"clicks"`
			CTR         float64 `json:"ctr"`
		}{
			CreativeID:  m.CreativeID,
			Impressions: m.Impressions,
			Clicks:      m.Clicks,
			CTR:         ctr,
		})
	}
	return &dto, nil
}

// randBucket 将 userKey hash 到 [0,9999]
func randBucket(userKey string) int {
	if userKey == "" {
		userKey = uuid.New().String()
	}
	h := uuid.NewSHA1(uuid.NameSpaceOID, []byte(userKey))
	b := h[:]
	val := int(b[0])<<8 + int(b[1])
	return val % 10000
}

package service

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"ads-creative-gen-platform/internal/experiment/repository"
	"ads-creative-gen-platform/internal/models"

	"github.com/google/uuid"
)

// ExperimentService 实验服务
type ExperimentService struct {
	repo repository.ExperimentRepository
}

func NewExperimentService() *ExperimentService {
	return &ExperimentService{repo: repository.NewExperimentRepository()}
}

// NewExperimentServiceWithRepo 支持依赖注入
func NewExperimentServiceWithRepo(repo repository.ExperimentRepository) *ExperimentService {
	return &ExperimentService{repo: repo}
}

// CreateExperimentInput 创建实验输入
type CreateExperimentInput struct {
	Name        string                   `json:"name"`
	ProductName string                   `json:"product_name,omitempty"`
	Variants    []ExperimentVariantInput `json:"variants"`
}

type ExperimentVariantInput struct {
	CreativeID    string   `json:"creative_id"`              // uuid 或 数字字符串
	Weight        float64  `json:"weight"`                   // 0-1
	Title         string   `json:"title,omitempty"`          // 覆盖标题
	ProductName   string   `json:"product_name,omitempty"`   // 覆盖产品名
	ImageURL      string   `json:"image_url,omitempty"`      // 覆盖图片URL
	CTAText       string   `json:"cta_text,omitempty"`       // 覆盖CTA文案
	SellingPoints []string `json:"selling_points,omitempty"` // 覆盖卖点（用户选择的）
}

// AssignedVariant 分流结果
type AssignedVariant struct {
	Variant models.ExperimentVariant
	Asset   *models.CreativeAsset
}

// Metrics DTO
type ExperimentMetricsDTO struct {
	ExperimentID string `json:"experiment_id"`
	Variants     []struct {
		CreativeID  uint    `json:"creative_id"`
		Impressions int64   `json:"impressions"`
		Clicks      int64   `json:"clicks"`
		CTR         float64 `json:"ctr"`
	} `json:"variants"`
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

	experiments, total, err := s.repo.ListExperiments(status, page, pageSize)
	if err != nil {
		return nil, err
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

	if err := s.repo.CreateExperiment(&exp); err != nil {
		return nil, fmt.Errorf("create experiment failed: %w", err)
	}

	// 计算桶
	acc := 0
	var variants []models.ExperimentVariant
	for _, v := range input.Variants {
		if v.CreativeID == "" {
			return nil, errors.New("creative_id required")
		}
		asset, creativeNumericID, err := s.lookupAsset(v.CreativeID)
		if err != nil {
			return nil, err
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

		// 使用用户覆盖的值，如果没有则使用 asset 的默认值
		title := v.Title
		if title == "" {
			title = asset.Title
		}
		productName := v.ProductName
		if productName == "" {
			productName = asset.ProductName
		}
		imageURL := v.ImageURL
		if imageURL == "" {
			imageURL = asset.PublicURL
		}
		ctaText := v.CTAText
		if ctaText == "" {
			ctaText = asset.CTAText
		}
		sellingPoints := v.SellingPoints
		if len(sellingPoints) == 0 {
			sellingPoints = asset.SellingPoints
		}

		variants = append(variants, models.ExperimentVariant{
			ExperimentID:  exp.ID,
			CreativeID:    creativeNumericID,
			Weight:        v.Weight,
			BucketStart:   start,
			BucketEnd:     end,
			Title:         title,
			ProductName:   productName,
			ImageURL:      imageURL,
			CTAText:       ctaText,
			SellingPoints: sellingPoints,
		})
	}
	// 调整最后一个覆盖到 9999
	if len(variants) > 0 {
		variants[len(variants)-1].BucketEnd = 9999
	}

	if err := s.repo.CreateVariants(variants); err != nil {
		return nil, fmt.Errorf("create variants failed: %w", err)
	}

	exp.Variants = variants
	return &exp, nil
}

func (s *ExperimentService) lookupAsset(creativeID string) (*models.CreativeAsset, uint, error) {
	var asset *models.CreativeAsset
	var numericID uint

	if parsed, err := strconv.ParseUint(creativeID, 10, 64); err == nil {
		numericID = uint(parsed)
		a, _ := s.repo.FindAssetByID(numericID)
		asset = a
	} else {
		a, err := s.repo.FindAssetByUUID(creativeID)
		if err == nil && a != nil {
			numericID = a.ID
			asset = a
		}
	}

	if numericID == 0 {
		return nil, 0, fmt.Errorf("creative_id %s not found", creativeID)
	}

	return asset, numericID, nil
}

// UpdateStatus 更新实验状态
func (s *ExperimentService) UpdateStatus(id string, status models.ExperimentStatus) error {
	if status != models.ExpActive && status != models.ExpPaused && status != models.ExpArchived && status != models.ExpDraft {
		return fmt.Errorf("invalid status")
	}
	fields := map[string]interface{}{"status": status}
	now := time.Now()
	if status == models.ExpActive {
		fields["start_at"] = now
	}
	if status == models.ExpArchived {
		fields["end_at"] = now
	}
	return s.repo.UpdateExperimentFields(id, fields)
}

// Assign 分流（返回变体与创意信息）
func (s *ExperimentService) Assign(id string, userKey string) (*AssignedVariant, error) {
	exp, err := s.repo.GetExperimentWithVariants(id)
	if err != nil {
		return nil, fmt.Errorf("experiment not found: %w", err)
	}
	if exp.Status != models.ExpActive {
		return nil, fmt.Errorf("experiment not active")
	}

	bucket := randBucket(userKey)
	for _, v := range exp.Variants {
		if bucket >= v.BucketStart && bucket <= v.BucketEnd {
			asset, err := s.repo.FindAssetWithTaskByID(v.CreativeID)
			if err != nil {
				return &AssignedVariant{Variant: v, Asset: nil}, nil
			}
			return &AssignedVariant{Variant: v, Asset: asset}, nil
		}
	}
	return nil, fmt.Errorf("no variant matched")
}

// Hit 记录曝光
func (s *ExperimentService) Hit(id string, creativeID uint) error {
	return s.incMetric(id, creativeID, true)
}

// Click 记录点击
func (s *ExperimentService) Click(id string, creativeID uint) error {
	return s.incMetric(id, creativeID, false)
}

// GetMetrics 获取实验指标
func (s *ExperimentService) GetMetrics(id string) (interface{}, error) {
	exp, err := s.repo.GetExperimentByUUID(id)
	if err != nil {
		return nil, fmt.Errorf("experiment not found: %w", err)
	}

	metrics, err := s.repo.ListMetrics(exp.ID)
	if err != nil {
		return nil, err
	}

	dto := ExperimentMetricsDTO{ExperimentID: id}
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

func (s *ExperimentService) incMetric(expUUID string, creativeID uint, isImpression bool) error {
	exp, err := s.repo.GetExperimentByUUID(expUUID)
	if err != nil {
		return fmt.Errorf("experiment not found: %w", err)
	}

	metric, err := s.repo.GetMetric(exp.ID, creativeID)
	if err != nil || metric == nil {
		metric = &models.ExperimentMetric{
			ExperimentID: exp.ID,
			CreativeID:   creativeID,
			Impressions:  0,
			Clicks:       0,
		}
	}

	if isImpression {
		metric.Impressions++
	} else {
		metric.Clicks++
	}
	metric.UpdatedAt = time.Now()

	return s.repo.SaveMetric(metric)
}

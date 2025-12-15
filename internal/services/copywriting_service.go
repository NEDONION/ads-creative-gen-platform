package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"

	"github.com/google/uuid"
)

// CopywritingService 负责文案生成与确认
type CopywritingService struct {
	qwenClient *QwenClient
}

// NewCopywritingService 构造服务
func NewCopywritingService() *CopywritingService {
	return &CopywritingService{
		qwenClient: NewQwenClient(),
	}
}

// GenerateCopywritingInput 文案生成输入
type GenerateCopywritingInput struct {
	UserID      uint   `json:"user_id"`
	ProductName string `json:"product_name"`
	Language    string `json:"language,omitempty"`
}

// GenerateCopywritingOutput 文案生成输出
type GenerateCopywritingOutput struct {
	TaskID                 string   `json:"task_id"`
	CTACandidates          []string `json:"cta_candidates"`
	SellingPointCandidates []string `json:"selling_point_candidates"`
}

// ConfirmCopywritingInput 用户确认输入
type ConfirmCopywritingInput struct {
	TaskID            string   `json:"task_id"`
	SelectedCTAIndex  int      `json:"selected_cta_index"`
	SelectedSPIndexes []int    `json:"selected_sp_indexes"`
	EditedCTA         string   `json:"edited_cta,omitempty"`
	EditedSPs         []string `json:"edited_sps,omitempty"`
	ProductImageURL   string   `json:"product_image_url,omitempty"`
	Style             string   `json:"style,omitempty"`
	NumVariants       int      `json:"num_variants,omitempty"`
	Formats           []string `json:"formats,omitempty"`
}

// GenerateCopywriting 调用 LLM 并创建任务
func (s *CopywritingService) GenerateCopywriting(input GenerateCopywritingInput) (*GenerateCopywritingOutput, error) {
	if input.ProductName == "" {
		return nil, errors.New("product_name is required")
	}

	targetLanguage := resolveLanguage(input.ProductName, input.Language)

	result, err := s.qwenClient.GenerateCopywriting(input.ProductName, targetLanguage)
	if err != nil {
		return nil, err
	}

	task := models.CreativeTask{
		UUIDModel: models.UUIDModel{
			UUID: uuid.New().String(),
		},
		UserID:                 input.UserID,
		Title:                  input.ProductName,
		ProductName:            input.ProductName,
		CTACandidates:          models.StringArray(result.CTAOptions),
		SellingPointCandidates: models.StringArray(result.SellingPointOptions),
		RequestedFormats:       models.StringArray{"1:1"},
		RequestedStyles:        models.StringArray{""},
		NumVariants:            3,
		Status:                 models.TaskDraft,
		CopywritingGenerated:   true,
		CopywritingRaw:         result.RawResponse,
		PromptUsed:             fmt.Sprintf("copywriting_language=%s", targetLanguage),
	}

	if err := database.DB.Create(&task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &GenerateCopywritingOutput{
		TaskID:                 task.UUID,
		CTACandidates:          result.CTAOptions,
		SellingPointCandidates: result.SellingPointOptions,
	}, nil
}

// ConfirmCopywriting 选择/编辑文案并更新任务
func (s *CopywritingService) ConfirmCopywriting(input ConfirmCopywritingInput) (*models.CreativeTask, error) {
	if input.TaskID == "" {
		return nil, errors.New("task_id is required")
	}

	var task models.CreativeTask
	if err := database.DB.Where("uuid = ?", input.TaskID).First(&task).Error; err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if len(task.CTACandidates) == 0 || len(task.SellingPointCandidates) == 0 {
		return nil, errors.New("task has no copywriting candidates")
	}

	if input.SelectedCTAIndex < 0 || input.SelectedCTAIndex >= len(task.CTACandidates) {
		return nil, errors.New("selected_cta_index out of range")
	}

	if len(input.SelectedSPIndexes) == 0 && len(input.EditedSPs) == 0 {
		return nil, errors.New("at least one selling point is required")
	}

	finalCTA := input.EditedCTA
	if finalCTA == "" {
		finalCTA = task.CTACandidates[input.SelectedCTAIndex]
	}

	var finalSPs []string
	if len(input.EditedSPs) > 0 {
		finalSPs = input.EditedSPs
	} else {
		for _, idx := range input.SelectedSPIndexes {
			if idx < 0 || idx >= len(task.SellingPointCandidates) {
				return nil, errors.New("selected_sp_indexes out of range")
			}
			finalSPs = append(finalSPs, task.SellingPointCandidates[idx])
		}
	}

	formats := input.Formats
	if len(formats) == 0 {
		formats = []string{"1:1"}
	}

	if input.NumVariants <= 0 {
		input.NumVariants = 2
	}

	selectedSPIndexes := make(models.StringArray, 0, len(input.SelectedSPIndexes))
	for _, idx := range input.SelectedSPIndexes {
		selectedSPIndexes = append(selectedSPIndexes, strconv.Itoa(idx))
	}

	updates := map[string]interface{}{
		"cta_text":            finalCTA,
		"selling_points":      models.StringArray(finalSPs),
		"selected_cta_index":  input.SelectedCTAIndex,
		"selected_sp_indexes": selectedSPIndexes,
		"product_image_url":   input.ProductImageURL,
		"requested_styles":    models.StringArray{input.Style},
		"requested_formats":   models.StringArray(formats),
		"num_variants":        input.NumVariants,
	}

	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update task failed: %w", err)
	}

	// 重新查询最新任务
	if err := database.DB.Where("uuid = ?", input.TaskID).First(&task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// resolveLanguage 决定生成语言（显式选择优先，其次自动检测）
func resolveLanguage(productName, language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "zh" || lang == "en" {
		return lang
	}

	// 简单检测：存在中文字符则用中文，否则英文
	hasHan := false
	alphaCount := 0
	hanCount := 0
	for _, r := range productName {
		if unicode.Is(unicode.Han, r) {
			hasHan = true
			hanCount++
		} else if unicode.IsLetter(r) {
			alphaCount++
		}
	}

	if hasHan && hanCount >= alphaCount {
		return "zh"
	}

	if alphaCount > 0 {
		return "en"
	}

	// 默认中文
	return "zh"
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/creative/repository"
	"ads-creative-gen-platform/internal/infra/cache"
	"ads-creative-gen-platform/internal/infra/llm"
	"ads-creative-gen-platform/internal/infra/storage"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"
	"ads-creative-gen-platform/internal/settings"
	"ads-creative-gen-platform/internal/shared"
	"ads-creative-gen-platform/internal/tracing"
	"ads-creative-gen-platform/pkg/database"

	"github.com/google/uuid"
)

// CreativeService 创意生成服务
type CreativeService struct {
	taskRepo    ports.TaskRepository
	assetRepo   ports.AssetRepository
	processor   *TaskProcessor
	enqueueFunc func(taskID uint) error
	traceSvc    *tracing.TraceService
}

// NewCreativeService 创建服务
func NewCreativeService() *CreativeService {
	llmClient := llm.NewTongyiClient()
	storageClient := storage.NewQiniuClient()
	baseTaskRepo := repository.NewTaskRepository(database.DB)
	baseAssetRepo := repository.NewAssetRepository(database.DB)

	poller := Poller{Interval: settings.PollInterval, MaxAttempts: settings.MaxPollAttempts}

	cacheCfg := config.CacheConfig
	dataCache := cache.NewConfiguredCache(cacheCfg)
	if cacheCfg != nil && cacheCfg.DisableCreative {
		dataCache = cache.NoopCache{}
	}
	ttl := time.Minute
	if cacheCfg != nil && cacheCfg.DefaultTTL > 0 {
		ttl = cacheCfg.DefaultTTL
	}

	taskRepo := repository.NewCachedTaskRepository(baseTaskRepo, dataCache, ttl)
	assetRepo := repository.NewCachedAssetRepository(baseAssetRepo, dataCache, ttl)

	return &CreativeService{
		taskRepo:  taskRepo,
		assetRepo: assetRepo,
		processor: NewTaskProcessor(llmClient, storageClient, taskRepo, assetRepo, poller),
		traceSvc:  tracing.NewTraceService(),
	}
}

// NewCreativeServiceWithDeps 支持依赖注入
func NewCreativeServiceWithDeps(
	taskRepo ports.TaskRepository,
	assetRepo ports.AssetRepository,
	processor *TaskProcessor,
	enqueue func(taskID uint) error,
	traceSvc *tracing.TraceService,
) *CreativeService {
	if traceSvc == nil {
		traceSvc = tracing.NewTraceService()
	}
	return &CreativeService{
		taskRepo:    taskRepo,
		assetRepo:   assetRepo,
		processor:   processor,
		enqueueFunc: enqueue,
		traceSvc:    traceSvc,
	}
}

// SetEnqueuer 设置任务入队方法（便于外部注入 Runner）
func (s *CreativeService) SetEnqueuer(enqueue func(taskID uint) error) {
	s.enqueueFunc = enqueue
}

// CreateTaskInput 创建任务输入
type CreateTaskInput struct {
	UserID          uint
	Title           string
	SellingPoints   []string
	ProductImageURL string
	Formats         []string
	Style           string
	CTAText         string
	NumVariants     int
	VariantPrompts  []string
	VariantStyles   []string
}

// CreateTask 创建创意生成任务
func (s *CreativeService) CreateTask(input CreateTaskInput) (*models.CreativeTask, error) {
	// 默认值
	if len(input.Formats) == 0 {
		input.Formats = []string{settings.DefaultFormat}
	}
	if input.NumVariants <= 0 {
		input.NumVariants = settings.DefaultNumVariants
	}

	// 创建任务
	task := models.CreativeTask{
		UUIDModel: models.UUIDModel{
			UUID: uuid.New().String(),
		},
		UserID:           input.UserID,
		Title:            input.Title,
		SellingPoints:    models.StringArray(input.SellingPoints),
		ProductImageURL:  input.ProductImageURL,
		RequestedFormats: models.StringArray(input.Formats),
		RequestedStyles:  models.StringArray{input.Style},
		NumVariants:      input.NumVariants,
		CTAText:          input.CTAText,
		VariantPrompts:   models.StringArray(input.VariantPrompts),
		VariantStyles:    models.StringArray(input.VariantStyles),
		Status:           models.TaskPending,
		Progress:         0,
	}

	// 保存到数据库
	if err := s.taskRepo.Create(context.Background(), &task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.enqueueOrProcess(task.ID)

	return &task, nil
}

// StartCreativeOptions 启动创意生成选项
type StartCreativeOptions struct {
	ProductImageURL string
	Style           string
	NumVariants     int
	Formats         []string
	VariantPrompts  []string
	VariantStyles   []string
}

// StartCreativeGeneration 根据已有任务启动生成
func (s *CreativeService) StartCreativeGeneration(taskUUID string, opts *StartCreativeOptions) error {
	if taskUUID == "" {
		return errors.New("task_id is required")
	}
	ctx := context.Background()

	task, err := s.taskRepo.GetByUUID(ctx, taskUUID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":                models.TaskQueued,
		"progress":              0,
		"error_message":         "",
		"queued_at":             &now,
		"started_at":            nil,
		"completed_at":          nil,
		"processing_duration":   nil,
		"first_asset_url":       "",
		"retry_to":              "",
		"retry_from":            task.RetryFrom,
		"prompt_used":           "",
		"copywriting_generated": task.CopywritingGenerated,
	}
	if opts != nil {
		if opts.ProductImageURL != "" {
			updates["product_image_url"] = opts.ProductImageURL
		}
		if opts.Style != "" {
			updates["requested_styles"] = models.StringArray{opts.Style}
		}
		if opts.NumVariants > 0 {
			updates["num_variants"] = opts.NumVariants
		}
		if len(opts.Formats) > 0 {
			updates["requested_formats"] = models.StringArray(opts.Formats)
		}
		if len(opts.VariantPrompts) > 0 {
			updates["variant_prompts"] = models.StringArray(opts.VariantPrompts)
		}
		if len(opts.VariantStyles) > 0 {
			updates["variant_styles"] = models.StringArray(opts.VariantStyles)
		}
	}

	if s.traceSvc != nil {
		_, _ = s.traceSvc.FailRunningBySource(task.UUID, "restart creative task")
	}

	if err := s.taskRepo.UpdateFields(ctx, task.ID, updates); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if err := s.enqueueOrProcess(task.ID); err != nil {
		return fmt.Errorf("enqueue task failed: %w", err)
	}

	return nil
}

// cloneTask 基于旧任务创建新任务，保留配置并标记来源
func (s *CreativeService) cloneTask(old *models.CreativeTask) *models.CreativeTask {
	if old == nil {
		return nil
	}
	return &models.CreativeTask{
		UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
		UserID:    old.UserID,
		ProjectID: old.ProjectID,

		Title:           old.Title,
		ProductName:     old.ProductName,
		SellingPoints:   append(models.StringArray{}, old.SellingPoints...),
		ProductImageURL: old.ProductImageURL,
		BrandLogoURL:    old.BrandLogoURL,
		CopywritingRaw:  old.CopywritingRaw,
		PromptUsed:      old.PromptUsed,

		RequestedFormats:       append(models.StringArray{}, old.RequestedFormats...),
		RequestedStyles:        append(models.StringArray{}, old.RequestedStyles...),
		NumVariants:            old.NumVariants,
		CTAText:                old.CTAText,
		CTACandidates:          append(models.StringArray{}, old.CTACandidates...),
		SellingPointCandidates: append(models.StringArray{}, old.SellingPointCandidates...),
		SelectedCTAIndex:       old.SelectedCTAIndex,
		SelectedSPIndexes:      append(models.StringArray{}, old.SelectedSPIndexes...),
		CopywritingGenerated:   old.CopywritingGenerated,
		VariantPrompts:         append(models.StringArray{}, old.VariantPrompts...),
		VariantStyles:          append(models.StringArray{}, old.VariantStyles...),

		Status:    models.TaskPending,
		Progress:  0,
		RetryFrom: old.UUID,
	}
}

// GetTask 查询任务详情（含资产）
func (s *CreativeService) GetTask(taskUUID string) (*models.CreativeTask, error) {
	if taskUUID == "" {
		return nil, errors.New("task_id is required")
	}
	task, err := s.taskRepo.GetByUUIDWithAssets(context.Background(), taskUUID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return task, nil
}

// DeleteTask 删除任务及其资产（软删除）
func (s *CreativeService) DeleteTask(taskUUID string) error {
	ctx := context.Background()

	task, err := s.taskRepo.GetByUUID(ctx, taskUUID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if err := s.assetRepo.DeleteByTaskID(ctx, task.ID); err != nil {
		return fmt.Errorf("delete assets failed: %w", err)
	}

	if err := s.taskRepo.Delete(ctx, task); err != nil {
		return fmt.Errorf("delete task failed: %w", err)
	}

	return nil
}

// ListTasksQuery 任务查询参数
type ListTasksQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	UserID   uint   `json:"user_id"`
}

// TaskDTO 任务数据传输对象
type TaskDTO struct {
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	ProductName   string             `json:"product_name,omitempty"`
	CTAText       string             `json:"cta_text,omitempty"`
	SellingPoints models.StringArray `json:"selling_points,omitempty"`
	Status        string             `json:"status"`
	Progress      int                `json:"progress"`
	CreatedAt     string             `json:"created_at"`
	CompletedAt   string             `json:"completed_at,omitempty"`
	ErrorMessage  string             `json:"error_message,omitempty"`
	FirstImage    string             `json:"first_image,omitempty"`
	RetryFrom     string             `json:"retry_from,omitempty"`
	RetryTo       string             `json:"retry_to,omitempty"`
}

// ListAssetsQuery 素材查询参数
type ListAssetsQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Format   string `json:"format"`
	TaskID   string `json:"task_id"`
}

// CreativeAssetDTO 素材数据传输对象
type CreativeAssetDTO struct {
	ID               string             `json:"id"`
	NumericID        uint               `json:"numeric_id"`
	Format           string             `json:"format"`
	ImageURL         string             `json:"image_url"`
	Width            int                `json:"width"`
	Height           int                `json:"height"`
	Title            string             `json:"title,omitempty"`
	ProductName      string             `json:"product_name,omitempty"`
	CTAText          string             `json:"cta_text,omitempty"`
	SellingPoints    models.StringArray `json:"selling_points,omitempty"`
	Style            string             `json:"style,omitempty"`
	GenerationPrompt string             `json:"generation_prompt,omitempty"`
}

// ListAllAssets 获取素材列表
func (s *CreativeService) ListAllAssets(query ListAssetsQuery) ([]CreativeAssetDTO, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	domainQuery := shared.ListAssetsQuery{
		Page:     query.Page,
		PageSize: query.PageSize,
		Format:   query.Format,
		TaskID:   query.TaskID,
	}

	assets, total, err := s.assetRepo.List(context.Background(), domainQuery)
	if err != nil {
		return nil, 0, err
	}

	result := make([]CreativeAssetDTO, 0, len(assets))
	for _, asset := range assets {
		result = append(result, CreativeAssetDTO{
			ID:               asset.UUID,
			NumericID:        asset.ID,
			Format:           asset.Format,
			ImageURL:         asset.PublicURL,
			Width:            asset.Width,
			Height:           asset.Height,
			Title:            asset.Title,
			ProductName:      asset.ProductName,
			CTAText:          asset.CTAText,
			SellingPoints:    asset.SellingPoints,
			Style:            asset.Style,
			GenerationPrompt: asset.GenerationPrompt,
		})
	}

	return result, total, nil
}

// ListAllTasks 获取所有任务
func (s *CreativeService) ListAllTasks(query ListTasksQuery) ([]TaskDTO, int64, error) {
	domainQuery := shared.ListTasksQuery{
		Page:     query.Page,
		PageSize: query.PageSize,
		Status:   query.Status,
		UserID:   query.UserID,
	}

	tasks, total, err := s.taskRepo.List(context.Background(), domainQuery)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	taskDTOs := make([]TaskDTO, 0, len(tasks))
	for _, task := range tasks {
		completedAt := ""
		if task.CompletedAt != nil {
			completedAt = task.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		firstImage := ""
		if task.FirstAssetURL != "" {
			firstImage = task.FirstAssetURL
		} else if len(task.Assets) > 0 {
			firstImage = getPublicURL(&task.Assets[0])
		}

		taskDTO := TaskDTO{
			ID:            task.UUID,
			Title:         task.Title,
			ProductName:   task.ProductName,
			CTAText:       task.CTAText,
			SellingPoints: task.SellingPoints,
			Status:        string(task.Status),
			Progress:      task.Progress,
			CreatedAt:     task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CompletedAt:   completedAt,
			ErrorMessage:  task.ErrorMessage,
			FirstImage:    firstImage,
			RetryFrom:     task.RetryFrom,
			RetryTo:       task.RetryTo,
		}
		taskDTOs = append(taskDTOs, taskDTO)
	}

	return taskDTOs, total, nil
}

func generatePrompt(title string, sellingPoints models.StringArray, style string) string {
	prompt := fmt.Sprintf("Product advertisement image for: %s", title)

	if len(sellingPoints) > 0 {
		prompt += ". Key selling points: " + strings.Join(sellingPoints, "; ")
	}

	if style != "" {
		prompt += ". Style: " + style
	}

	prompt += ". The image should be attractive, high quality, and suitable for digital advertising."

	return prompt
}

func styleAt(arr models.StringArray, idx int) string {
	if len(arr) == 0 {
		return ""
	}
	if idx < 0 || idx >= len(arr) {
		return arr[0]
	}
	return arr[idx]
}

func getPublicURL(asset *models.CreativeAsset) string {
	if asset == nil {
		return ""
	}
	return asset.PublicURL
}

func (s *CreativeService) enqueueOrProcess(taskID uint) error {
	if s.enqueueFunc != nil {
		return s.enqueueFunc(taskID)
	}
	if s.processor != nil {
		return s.processor.Process(context.Background(), taskID)
	}
	return nil
}

// ProcessTaskWithContext 提供给任务执行器/Runner 的入口。
func (s *CreativeService) ProcessTaskWithContext(ctx context.Context, taskID uint) error {
	if s.processor == nil {
		return errors.New("processor not set")
	}
	return s.processor.Process(ctx, taskID)
}

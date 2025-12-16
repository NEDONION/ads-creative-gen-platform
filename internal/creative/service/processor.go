package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ads-creative-gen-platform/internal/infra/llm"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/ports"

	"github.com/google/uuid"
)

// LLMClient 定义生成/查询接口，便于测试替换。
type LLMClient interface {
	GenerateImage(ctx context.Context, prompt string, size string, num int, productName string, a string, b string) (*llm.ImageGenResponse, string, error)
	GenerateImageWithProduct(ctx context.Context, prompt string, productImageURL string, size string, num int, productName string, a string, b string) (*llm.ImageGenResponse, string, error)
	QueryTask(ctx context.Context, traceID string, taskID string, requestID string) (*llm.ImageGenResponse, error)
	FinishTrace(traceID string, status string, firstURL string, errMsg string)
}

// StorageClient 定义上传接口，便于测试替换。
type StorageClient interface {
	UploadFromURL(ctx context.Context, url string, key string) (string, error)
	GenerateKey(name string) string
}

// Poller 控制轮询策略。
type Poller struct {
	Interval    time.Duration
	MaxAttempts int
	Sleep       func(time.Duration)
}

func (p *Poller) interval() time.Duration {
	if p.Interval <= 0 {
		return 2 * time.Second
	}
	return p.Interval
}

func (p *Poller) attempts() int {
	if p.MaxAttempts <= 0 {
		return 60
	}
	return p.MaxAttempts
}

func (p *Poller) sleep(d time.Duration) {
	if p.Sleep != nil {
		p.Sleep(d)
		return
	}
	time.Sleep(d)
}

// TaskProcessor 负责执行创意任务的完整工作流。
type TaskProcessor struct {
	llmClient     LLMClient
	storageClient StorageClient
	taskRepo      ports.TaskRepository
	assetRepo     ports.AssetRepository
	poller        Poller
}

// NewTaskProcessor 创建处理器，注入依赖与轮询策略。
func NewTaskProcessor(
	llmClient LLMClient,
	storageClient StorageClient,
	taskRepo ports.TaskRepository,
	assetRepo ports.AssetRepository,
	poller Poller,
) *TaskProcessor {
	return &TaskProcessor{
		llmClient:     llmClient,
		storageClient: storageClient,
		taskRepo:      taskRepo,
		assetRepo:     assetRepo,
		poller:        poller,
	}
}

// Process 执行任务，负责生成、轮询与落地。
func (p *TaskProcessor) Process(ctx context.Context, taskID uint) error {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now()
	if err := p.taskRepo.UpdateFields(ctx, taskID, map[string]interface{}{
		"status":     models.TaskProcessing,
		"started_at": now,
		"progress":   10,
	}); err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	task, err := p.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}

	// 针对变体配置单独处理
	if len(task.VariantPrompts) > 0 || len(task.VariantStyles) > 1 {
		return p.processPerVariant(ctx, task, now)
	}

	// 生成提示词
	prompt := generatePrompt(task.Title, task.SellingPoints, styleAt(task.RequestedStyles, 0))
	_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{"prompt_used": prompt})

	_ = p.taskRepo.UpdateProgress(ctx, task.ID, 30)

	var resp *llm.ImageGenResponse
	var traceID string

	if task.ProductImageURL != "" {
		resp, traceID, err = p.llmClient.GenerateImageWithProduct(ctx, prompt, task.ProductImageURL, "1024*1024", task.NumVariants, task.ProductName, "", task.ProductName)
	} else {
		resp, traceID, err = p.llmClient.GenerateImage(ctx, prompt, "1024*1024", task.NumVariants, task.ProductName, "", task.ProductName)
	}
	if err != nil {
		_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{
			"status":        models.TaskFailed,
			"error_message": fmt.Sprintf("API调用失败: %v", err),
			"progress":      100,
		})
		if traceID != "" {
			p.llmClient.FinishTrace(traceID, "failed", "", err.Error())
		}
		return err
	}

	tongyiTaskID := resp.Output.TaskID
	var firstPublicURL string

	for i := 0; i < p.poller.attempts(); i++ {
		p.poller.sleep(p.poller.interval())

		queryResp, err := p.llmClient.QueryTask(ctx, traceID, tongyiTaskID, task.UUID)
		if err != nil {
			log.Printf("查询任务 %s 失败: %v", tongyiTaskID, err)
			continue
		}

		if queryResp.Output.TaskStatus == "SUCCEEDED" {
			for idx, result := range queryResp.Output.Results {
				publicURL, storageType, originalPath := p.handleUpload(ctx, task.UUID, idx, result.URL)

				asset := models.CreativeAsset{
					UUIDModel:        models.UUIDModel{UUID: uuid.New().String()},
					TaskID:           task.ID,
					Title:            task.Title,
					ProductName:      task.ProductName,
					CTAText:          task.CTAText,
					SellingPoints:    task.SellingPoints,
					Format:           "1:1",
					Width:            1024,
					Height:           1024,
					StorageType:      storageType,
					PublicURL:        publicURL,
					OriginalPath:     originalPath,
					Style:            task.RequestedStyles[0],
					VariantIndex:     &idx,
					GenerationPrompt: prompt,
					ModelName:        "wanx-v1",
				}

				if err := p.assetRepo.Create(ctx, &asset); err != nil {
					log.Printf("保存资产失败: %v", err)
					continue
				}
				if idx == 0 {
					firstPublicURL = publicURL
				}
			}

			if firstPublicURL != "" {
				_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{"first_asset_url": firstPublicURL})
			}

			completedAt := time.Now()
			duration := int(completedAt.Sub(now).Seconds())
			_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{
				"status":              models.TaskCompleted,
				"progress":            100,
				"completed_at":        completedAt,
				"processing_duration": duration,
			})
			if traceID != "" {
				p.llmClient.FinishTrace(traceID, "success", firstPublicURL, "")
			}
			return nil
		}

		if queryResp.Output.TaskStatus == "FAILED" {
			msg := queryResp.Output.Message
			if msg == "" {
				msg = "任务失败，无具体错误信息"
			}
			_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{
				"status":        models.TaskFailed,
				"error_message": msg,
				"progress":      100,
			})
			if traceID != "" {
				p.llmClient.FinishTrace(traceID, "failed", "", msg)
			}
			return fmt.Errorf(msg)
		}

		progress := 60 + (i * 40 / p.poller.attempts())
		_ = p.taskRepo.UpdateProgress(ctx, task.ID, progress)
	}

	timeoutErr := fmt.Errorf("任务在%d秒后超时", int(p.poller.interval().Seconds())*p.poller.attempts())
	_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{
		"status":        models.TaskFailed,
		"error_message": timeoutErr.Error(),
		"progress":      100,
	})
	return timeoutErr
}

func (p *TaskProcessor) processPerVariant(ctx context.Context, task *models.CreativeTask, startedAt time.Time) error {
	numVariants := task.NumVariants
	if numVariants <= 0 {
		numVariants = 1
	}

	var firstPublicURL string

	for idx := 0; idx < numVariants; idx++ {
		style := styleAt(task.VariantStyles, idx)
		if style == "" {
			style = styleAt(task.RequestedStyles, 0)
		}
		prompt := strings.TrimSpace(styleAt(task.VariantPrompts, idx))
		if prompt == "" {
			prompt = generatePrompt(task.Title, task.SellingPoints, style)
		}

		size := "1024*1024"
		format := "1:1"
		if len(task.RequestedFormats) > idx && task.RequestedFormats[idx] != "" {
			format = task.RequestedFormats[idx]
		} else if len(task.RequestedFormats) > 0 {
			format = task.RequestedFormats[0]
		}

		var resp *llm.ImageGenResponse
		var err error
		var traceID string

		if task.ProductImageURL != "" {
			resp, traceID, err = p.llmClient.GenerateImageWithProduct(ctx, prompt, task.ProductImageURL, size, 1, task.ProductName, "", task.ProductName)
		} else {
			resp, traceID, err = p.llmClient.GenerateImage(ctx, prompt, size, 1, task.ProductName, "", task.ProductName)
		}
		if err != nil {
			_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{
				"status":        models.TaskFailed,
				"error_message": fmt.Sprintf("变体 %d 生成失败: %v", idx+1, err),
				"progress":      100,
			})
			if traceID != "" {
				p.llmClient.FinishTrace(traceID, "failed", "", err.Error())
			}
			return err
		}

		tongyiTaskID := resp.Output.TaskID
		success := false

		for i := 0; i < p.poller.attempts(); i++ {
			p.poller.sleep(p.poller.interval())
			queryResp, err := p.llmClient.QueryTask(ctx, traceID, tongyiTaskID, task.UUID)
			if err != nil {
				log.Printf("查询变体任务 %s 失败: %v", tongyiTaskID, err)
				continue
			}
			if queryResp.Output.TaskStatus == "SUCCEEDED" && len(queryResp.Output.Results) > 0 {
				result := queryResp.Output.Results[0]

				publicURL, storageType, originalPath := p.handleUpload(ctx, task.UUID, idx, result.URL)

				asset := models.CreativeAsset{
					UUIDModel:        models.UUIDModel{UUID: uuid.New().String()},
					TaskID:           task.ID,
					Title:            task.Title,
					ProductName:      task.ProductName,
					CTAText:          task.CTAText,
					SellingPoints:    task.SellingPoints,
					Format:           format,
					Width:            1024,
					Height:           1024,
					StorageType:      storageType,
					PublicURL:        publicURL,
					OriginalPath:     originalPath,
					Style:            style,
					VariantIndex:     &idx,
					GenerationPrompt: prompt,
					ModelName:        "wanx-v1",
				}

				if err := p.assetRepo.Create(ctx, &asset); err != nil {
					log.Printf("保存变体资产失败: %v", err)
				} else if idx == 0 {
					firstPublicURL = publicURL
				}
				success = true
				if traceID != "" {
					p.llmClient.FinishTrace(traceID, "success", publicURL, "")
				}
				break
			}

			if queryResp.Output.TaskStatus == "FAILED" {
				errMsg := queryResp.Output.Message
				if errMsg == "" {
					errMsg = "任务失败"
				}
				_ = p.taskRepo.UpdateFields(ctx, task.ID, map[string]interface{}{
					"status":        models.TaskFailed,
					"error_message": errMsg,
					"progress":      100,
				})
				if traceID != "" {
					p.llmClient.FinishTrace(traceID, "failed", "", errMsg)
				}
				return fmt.Errorf(errMsg)
			}
		}

		progress := 30 + (idx+1)*40/numVariants
		if success {
			_ = p.taskRepo.UpdateProgress(ctx, task.ID, progress)
		}
	}

	completedAt := time.Now()
	duration := int(completedAt.Sub(startedAt).Seconds())

	update := map[string]interface{}{
		"status":              models.TaskCompleted,
		"progress":            100,
		"completed_at":        completedAt,
		"processing_duration": duration,
	}
	if firstPublicURL != "" {
		update["first_asset_url"] = firstPublicURL
	}
	return p.taskRepo.UpdateFields(ctx, task.ID, update)
}

// handleUpload 处理存储上传并返回最终 URL/存储信息。
func (p *TaskProcessor) handleUpload(ctx context.Context, taskUUID string, idx int, originalURL string) (string, models.StorageType, string) {
	publicURL := originalURL
	storageType := models.StorageLocal
	originalPath := originalURL

	if p.storageClient == nil {
		return publicURL, storageType, originalPath
	}

	fileName := fmt.Sprintf("%s_%d", taskUUID, idx)
	qiniuURL, err := p.storageClient.UploadFromURL(ctx, originalURL, fileName)
	if err != nil {
		log.Printf("上传到存储失败: %v，使用原始URL", err)
		return publicURL, storageType, originalPath
	}

	publicURL = qiniuURL
	storageType = models.StorageQiniu
	originalPath = p.storageClient.GenerateKey(fileName)
	return publicURL, storageType, originalPath
}

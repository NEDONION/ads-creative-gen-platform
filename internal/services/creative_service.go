package services

import (
	"fmt"
	"log"
	"time"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"

	"github.com/google/uuid"
)

// CreativeService 创意生成服务
type CreativeService struct {
	tongyiClient *TongyiClient
	qiniuService *QiniuService
}

// NewCreativeService 创建服务
func NewCreativeService() *CreativeService {
	return &CreativeService{
		tongyiClient: NewTongyiClient(),
		qiniuService: NewQiniuService(),
	}
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
}

// CreateTask 创建创意生成任务
func (s *CreativeService) CreateTask(input CreateTaskInput) (*models.CreativeTask, error) {
	// 默认值
	if len(input.Formats) == 0 {
		input.Formats = []string{"1:1"}
	}
	if input.NumVariants == 0 {
		input.NumVariants = 1
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
		Status:           models.TaskPending,
		Progress:         0,
	}

	// 保存到数据库
	if err := database.DB.Create(&task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 异步处理任务
	go s.processTask(task.ID)

	return &task, nil
}

// processTask 处理任务
func (s *CreativeService) processTask(taskID uint) {
	// 更新状态为处理中
	now := time.Now()
	database.DB.Model(&models.CreativeTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     models.TaskProcessing,
		"started_at": now,
		"progress":   10,
	})

	// 查询任务详情
	var task models.CreativeTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		log.Printf("Failed to find task %d: %v", taskID, err)
		return
	}

	// 生成提示词
	prompt := s.generatePrompt(task.Title, task.SellingPoints, task.RequestedStyles)
	log.Printf("Generated prompt: %s", prompt)

	// 更新进度
	database.DB.Model(&task).Update("progress", 30)

	// 调用通义万相生成图片
	var resp *ImageGenResponse
	var err error

	if task.ProductImageURL != "" {
		// 带商品图生成
		resp, err = s.tongyiClient.GenerateImageWithProduct(prompt, task.ProductImageURL, "1024*1024")
	} else {
		// 纯文本生成
		resp, err = s.tongyiClient.GenerateImage(prompt, "1024*1024", task.NumVariants)
	}

	if err != nil {
		log.Printf("Failed to generate image: %v", err)
		database.DB.Model(&task).Updates(map[string]interface{}{
			"status":        models.TaskFailed,
			"error_message": err.Error(),
			"progress":      100,
		})
		return
	}

	// 更新进度
	database.DB.Model(&task).Update("progress", 60)

	// 轮询任务状态
	tongyiTaskID := resp.Output.TaskID
	for i := 0; i < 30; i++ { // 最多等待30次，每次2秒
		time.Sleep(2 * time.Second)

		queryResp, err := s.tongyiClient.QueryTask(tongyiTaskID)
		if err != nil {
			log.Printf("Failed to query task: %v", err)
			continue
		}

		log.Printf("Task status: %s", queryResp.Output.TaskStatus)

		if queryResp.Output.TaskStatus == "SUCCEEDED" {
			// 保存生成的图片
			for idx, result := range queryResp.Output.Results {
				tongyiURL := result.URL
				publicURL := tongyiURL
				storageType := models.StorageLocal

				originalPath := "" // 原始内部存储路径

				// 如果配置了七牛云，则上传到七牛云
				if s.qiniuService != nil {
					fileName := fmt.Sprintf("%s_%d", task.UUID, idx)
					qiniuURL, err := s.qiniuService.UploadFromURL(tongyiURL, fileName)
					if err != nil {
						log.Printf("Failed to upload to Qiniu: %v, using original URL", err)
						// 即使上传到七牛云失败，仍使用原始URL，但标记为本地存储
						publicURL = tongyiURL
						storageType = models.StorageLocal
						originalPath = tongyiURL // 使用原始URL作为原始路径
					} else {
						publicURL = qiniuURL
						storageType = models.StorageOSS
						// 生成对应的内部存储路径
						originalPath = s.qiniuService.generateKey(fmt.Sprintf("%s_%d", task.UUID, idx))
						log.Printf("Image uploaded to Qiniu: %s (original path: %s)", qiniuURL, originalPath)
					}
				} else {
					// 七牛云服务未配置，使用原始URL
					publicURL = tongyiURL
					storageType = models.StorageLocal
					originalPath = tongyiURL // 使用原始URL作为原始路径
				}

				asset := models.CreativeAsset{
					UUIDModel: models.UUIDModel{
						UUID: uuid.New().String(),
					},
					TaskID:           task.ID,
					Format:           "1:1",
					Width:            1024,
					Height:           1024,
					StorageType:      storageType,
					FilePath:         publicURL, // 当前的访问URL
					PublicURL:        publicURL,
					CDNURL:           publicURL,    // 公共图床URL
					OriginalPath:     originalPath, // 原始内部路径
					Style:            task.RequestedStyles[0],
					VariantIndex:     &idx,
					GenerationPrompt: prompt,
					ModelName:        "wanx-v1",
				}

				if err := database.DB.Create(&asset).Error; err != nil {
					log.Printf("Failed to save asset: %v", err)
				}
			}

			// 更新任务状态为完成
			completedAt := time.Now()
			duration := int(completedAt.Sub(now).Seconds())
			database.DB.Model(&task).Updates(map[string]interface{}{
				"status":              models.TaskCompleted,
				"progress":            100,
				"completed_at":        completedAt,
				"processing_duration": duration,
			})

			log.Printf("Task %d completed successfully", taskID)
			return

		} else if queryResp.Output.TaskStatus == "FAILED" {
			database.DB.Model(&task).Updates(map[string]interface{}{
				"status":        models.TaskFailed,
				"error_message": queryResp.Output.Message,
				"progress":      100,
			})
			log.Printf("Task %d failed: %s", taskID, queryResp.Output.Message)
			return
		}

		// 更新进度
		progress := 60 + (i * 40 / 30)
		database.DB.Model(&task).Update("progress", progress)
	}

	// 超时
	database.DB.Model(&task).Updates(map[string]interface{}{
		"status":        models.TaskFailed,
		"error_message": "Task timeout after 60 seconds",
		"progress":      100,
	})
}

// generatePrompt 生成提示词
func (s *CreativeService) generatePrompt(title string, sellingPoints models.StringArray, styles models.StringArray) string {
	prompt := fmt.Sprintf("Product advertisement image for: %s", title)

	if len(sellingPoints) > 0 {
		prompt += fmt.Sprintf(", features: %v", sellingPoints)
	}

	if len(styles) > 0 && styles[0] != "" {
		styleMap := map[string]string{
			"modern":  "modern and minimalist style",
			"elegant": "elegant and sophisticated style",
			"vibrant": "vibrant and energetic style",
		}
		if desc, ok := styleMap[styles[0]]; ok {
			prompt += ", " + desc
		}
	}

	prompt += ", high quality, professional photography, clean background, product focused"

	return prompt
}

// GetTask 获取任务详情
func (s *CreativeService) GetTask(taskUUID string) (*models.CreativeTask, error) {
	var task models.CreativeTask
	if err := database.DB.Preload("Assets").Where("uuid = ?", taskUUID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListAssetsQuery 查询参数
type ListAssetsQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Format   string `json:"format"`
	TaskID   string `json:"task_id"`
}

// ListAllAssets 获取所有创意素材
func (s *CreativeService) ListAllAssets(query ListAssetsQuery) ([]CreativeAssetDTO, int64, error) {
	var assets []models.CreativeAsset
	var total int64

	// 构建查询
	dbQuery := database.DB.Model(&models.CreativeAsset{}).Preload("Task")

	// 应用筛选条件
	if query.Format != "" {
		dbQuery = dbQuery.Where("format = ?", query.Format)
	}
	if query.TaskID != "" {
		dbQuery = dbQuery.Joins("JOIN creative_tasks ON creative_assets.task_id = creative_tasks.id").
			Where("creative_tasks.uuid = ?", query.TaskID)
	}

	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := dbQuery.Offset(offset).Limit(query.PageSize).Order("creative_assets.created_at DESC").Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	creativeDatas := make([]CreativeAssetDTO, 0, len(assets))
	for _, asset := range assets {
		creativeData := CreativeAssetDTO{
			ID:       asset.UUID,
			Format:   asset.Format,
			ImageURL: getPublicURL(&asset), // 优先使用公共URL
			Width:    asset.Width,
			Height:   asset.Height,
		}
		creativeDatas = append(creativeDatas, creativeData)
	}

	return creativeDatas, total, nil
}

// CreativeAssetDTO 素材数据传输对象
type CreativeAssetDTO struct {
	ID       string `json:"id"`
	Format   string `json:"format"`
	ImageURL string `json:"image_url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// getPublicURL 获取公共访问URL，优先级：CDNURL > PublicURL > FilePath
func getPublicURL(asset *models.CreativeAsset) string {
	if asset.CDNURL != "" {
		return asset.CDNURL
	}
	if asset.PublicURL != "" {
		return asset.PublicURL
	}
	return asset.FilePath
}

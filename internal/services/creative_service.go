package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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
	VariantPrompts  []string
	VariantStyles   []string
}

// CreateTask 创建创意生成任务
func (s *CreativeService) CreateTask(input CreateTaskInput) (*models.CreativeTask, error) {
	// 默认值
	if len(input.Formats) == 0 {
		input.Formats = []string{"1:1"}
	}
	if input.NumVariants <= 0 {
		input.NumVariants = 2 // 以2为基础
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
	updateResult := database.DB.Model(&models.CreativeTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     models.TaskProcessing,
		"started_at": now,
		"progress":   10,
	})

	if updateResult.Error != nil {
		log.Printf("更新任务状态失败: %v", updateResult.Error)
		return
	}

	// 查询任务详情
	var task models.CreativeTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		log.Printf("查找任务 %d 失败: %v", taskID, err)
		return
	}

	// 如果有针对每个变体的提示词/风格配置，逐个生成
	if len(task.VariantPrompts) > 0 || len(task.VariantStyles) > 1 {
		s.processTaskPerVariant(task)
		return
	}

	// 生成提示词
	prompt := s.generatePrompt(task.Title, task.SellingPoints, styleAt(task.RequestedStyles, 0))
	log.Printf("生成的提示词: %s", prompt)
	database.DB.Model(&task).Update("prompt_used", prompt)

	// 更新进度
	database.DB.Model(&task).Update("progress", 30)

	// 调用通义万相生成图片
	var resp *ImageGenResponse
	var err error
	var traceID string

	if task.ProductImageURL != "" {
		// 带商品图生成
		log.Printf("开始带商品图生成: %s", task.ProductImageURL)
		resp, traceID, err = s.tongyiClient.GenerateImageWithProduct(context.Background(), prompt, task.ProductImageURL, "1024*1024", task.NumVariants, task.UUID, "")
	} else {
		// 纯文本生成
		log.Printf("开始纯文本图像生成")
		resp, traceID, err = s.tongyiClient.GenerateImage(context.Background(), prompt, "1024*1024", task.NumVariants, task.UUID, "")
	}

	if err != nil {
		log.Printf("图像生成失败: %v", err)
		database.DB.Model(&task).Updates(map[string]interface{}{
			"status":        models.TaskFailed,
			"error_message": fmt.Sprintf("API调用失败: %v", err),
			"progress":      100,
		})
		if traceID != "" {
			s.tongyiClient.tracer.FinishTrace(traceID, "failed", "", err.Error())
		}
		return
	}

	log.Printf("成功创建通义任务, ID: %s", resp.Output.TaskID)

	// 更新进度
	database.DB.Model(&task).Update("progress", 60)

	// 轮询任务状态
	tongyiTaskID := resp.Output.TaskID
	for i := 0; i < 60; i++ { // 增加等待时间到120秒 (60次*2秒)
		time.Sleep(2 * time.Second)

		queryResp, err := s.tongyiClient.QueryTask(context.Background(), traceID, tongyiTaskID, task.UUID)
		if err != nil {
			log.Printf("查询任务 %s 失败: %v", tongyiTaskID, err)
			continue
		}

		log.Printf("任务 %s 状态: %s, 请求ID: %s", tongyiTaskID, queryResp.Output.TaskStatus, queryResp.RequestID)

		if queryResp.Output.TaskStatus == "SUCCEEDED" {
			log.Printf("任务 %s 成功, 生成 %d 个结果", tongyiTaskID, len(queryResp.Output.Results))

			firstPublicURL := ""

			// 保存生成的图片
			for idx, result := range queryResp.Output.Results {
				log.Printf("保存资产 %d 任务 %s, URL: %s", idx, task.UUID, result.URL)

				tongyiURL := result.URL
				publicURL := tongyiURL
				storageType := models.StorageLocal

				originalPath := "" // 原始内部存储路径

				// 如果配置了七牛云，则上传到七牛云
				if s.qiniuService != nil {
					fileName := fmt.Sprintf("%s_%d", task.UUID, idx)
					qiniuURL, err := s.qiniuService.UploadFromURL(tongyiURL, fileName)
					if err != nil {
						log.Printf("上传到七牛云失败: %v, 使用原始URL", err)
						// 即使上传到七牛云失败，仍使用原始URL，但标记为本地存储
						publicURL = tongyiURL
						storageType = models.StorageLocal
						originalPath = tongyiURL // 使用原始URL作为原始路径
					} else {
						publicURL = qiniuURL
						storageType = models.StorageQiniu
						// 生成对应的内部存储路径
						originalPath = s.qiniuService.generateKey(fmt.Sprintf("%s_%d", task.UUID, idx))
						log.Printf("图片已上传到七牛云: %s (原始路径: %s)", qiniuURL, originalPath)
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
					Title:            task.Title,
					ProductName:      task.ProductName,
					CTAText:          task.CTAText,
					SellingPoints:    task.SellingPoints,
					Format:           "1:1", // 根据实际尺寸确定格式
					Width:            1024,
					Height:           1024,
					StorageType:      storageType,
					PublicURL:        publicURL,    // 已拼接好的完整公共访问URL
					OriginalPath:     originalPath, // 原始内部路径
					Style:            task.RequestedStyles[0],
					VariantIndex:     &idx,
					GenerationPrompt: prompt,
					ModelName:        "wanx-v1",
				}

				// 在创建前验证数据
				log.Printf("创建资产 任务 %d: URL=%s, 格式=1:1, 存储类型=%s", task.ID, publicURL, storageType)

				if err := database.DB.Create(&asset).Error; err != nil {
					// 如果错误是因为缺少original_path字段，尝试手动构建SQL
					log.Printf("保存资产失败: %v", err)
					if strings.Contains(err.Error(), "original_path") {
						// 使用新的模型结构
						asset := models.CreativeAsset{
							UUIDModel: models.UUIDModel{
								UUID: uuid.New().String(),
							},
							TaskID:           task.ID,
							Format:           "1:1",
							Width:            1024,
							Height:           1024,
							StorageType:      storageType,
							PublicURL:        publicURL, // 已拼接好的完整公共访问URL
							Style:            task.RequestedStyles[0],
							VariantIndex:     &idx,
							GenerationPrompt: prompt,
							ModelName:        "wanx-v1",
						}

						// 现在使用更新后的模型结构保存
						if err := database.DB.Create(&asset).Error; err != nil {
							log.Printf("保存资产失败: %v", err)
						} else {
							log.Printf("成功保存资产 %s 任务 %d URL: %s", asset.UUID, taskID, asset.PublicURL)
						}
					} else {
						// 记录更详细的错误信息，这可能是导致资产未保存的原因
						log.Printf("资产详情 (尝试保存): 任务ID=%d, 格式=1:1, 宽度=1024, 高度=1024, 公共URL=%s", task.ID, publicURL)
						// 继续处理其他图片，不返回错误
					}
				} else {
					log.Printf("成功保存资产 %s 任务 %d URL: %s", asset.UUID, taskID, asset.PublicURL)
				}

				if idx == 0 {
					firstPublicURL = publicURL
				}
			}
			// 设置首图
			if firstPublicURL != "" {
				database.DB.Model(&task).Update("first_asset_url", firstPublicURL)
			}

			// 再次查询任务，确保在更新之前重新加载
			var updatedTask models.CreativeTask
			if err := database.DB.First(&updatedTask, taskID).Error; err != nil {
				log.Printf("重新加载任务 %d 失败: %v", taskID, err)
				return
			}

			// 更新任务状态为完成
			completedAt := time.Now()
			duration := int(completedAt.Sub(now).Seconds())

			result := database.DB.Model(&updatedTask).Updates(map[string]interface{}{
				"status":              models.TaskCompleted,
				"progress":            100,
				"completed_at":        completedAt,
				"processing_duration": duration,
			})

			if result.Error != nil {
				log.Printf("更新完成任务失败: %v", result.Error)
			} else {
				log.Printf("任务 %d 成功完成，包含 %d 个资产", taskID, len(queryResp.Output.Results))
			}
			if traceID != "" {
				s.tongyiClient.tracer.FinishTrace(traceID, "success", firstPublicURL, "")
			}

			return

		} else if queryResp.Output.TaskStatus == "FAILED" {
			errorMsg := queryResp.Output.Message
			if errorMsg == "" {
				errorMsg = "任务失败，无具体错误信息"
			}

			errUpdateResult := database.DB.Model(&task).Updates(map[string]interface{}{
				"status":        models.TaskFailed,
				"error_message": errorMsg,
				"progress":      100,
			})

			if errUpdateResult.Error != nil {
				log.Printf("更新失败任务失败: %v", errUpdateResult.Error)
			} else {
				log.Printf("任务 %d 失败: %s", taskID, errorMsg)
			}
			if traceID != "" {
				s.tongyiClient.tracer.FinishTrace(traceID, "failed", "", errorMsg)
			}
			return
		}

		// 更新进度
		progress := 60 + (i * 40 / 60) // 改为60次计算，避免进度超过100
		database.DB.Model(&task).Update("progress", progress)
	}

	// 超时
	timeoutErr := "任务在120秒后超时"
	errUpdateResult := database.DB.Model(&task).Updates(map[string]interface{}{
		"status":        models.TaskFailed,
		"error_message": timeoutErr,
		"progress":      100,
	})

	if errUpdateResult.Error != nil {
		log.Printf("更新超时任务失败: %v", errUpdateResult.Error)
	} else {
		log.Printf("任务 %d 超时: %s", taskID, timeoutErr)
	}
	if traceID != "" {
		s.tongyiClient.tracer.FinishTrace(traceID, "failed", "", timeoutErr)
	}
}

// processTaskPerVariant 针对每个变体使用单独提示词/风格生成
func (s *CreativeService) processTaskPerVariant(task models.CreativeTask) {
	now := time.Now()
	database.DB.Model(&models.CreativeTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     models.TaskProcessing,
		"started_at": now,
		"progress":   10,
	})

	// 重新获取最新任务
	if err := database.DB.First(&task, task.ID).Error; err != nil {
		log.Printf("查找任务 %d 失败: %v", task.ID, err)
		return
	}

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
			prompt = s.generatePrompt(task.Title, task.SellingPoints, style)
		}
		if prompt == "" {
			prompt = s.generatePrompt(task.Title, task.SellingPoints, "")
		}

		size := "1024*1024"
		format := "1:1"
		if len(task.RequestedFormats) > idx && task.RequestedFormats[idx] != "" {
			format = task.RequestedFormats[idx]
		} else if len(task.RequestedFormats) > 0 {
			format = task.RequestedFormats[0]
		}

		var resp *ImageGenResponse
		var err error
		var traceID string

		if task.ProductImageURL != "" {
			resp, traceID, err = s.tongyiClient.GenerateImageWithProduct(context.Background(), prompt, task.ProductImageURL, size, 1, task.UUID, "")
		} else {
			resp, traceID, err = s.tongyiClient.GenerateImage(context.Background(), prompt, size, 1, task.UUID, "")
		}
		if err != nil {
			log.Printf("变体 %d 生成失败: %v", idx, err)
			database.DB.Model(&task).Updates(map[string]interface{}{
				"status":        models.TaskFailed,
				"error_message": fmt.Sprintf("变体 %d 生成失败: %v", idx+1, err),
				"progress":      100,
			})
			if traceID != "" {
				s.tongyiClient.tracer.FinishTrace(traceID, "failed", "", err.Error())
			}
			return
		}

		tongyiTaskID := resp.Output.TaskID
		success := false

		for i := 0; i < 60; i++ {
			time.Sleep(2 * time.Second)
			queryResp, err := s.tongyiClient.QueryTask(context.Background(), traceID, tongyiTaskID, task.UUID)
			if err != nil {
				log.Printf("查询变体任务 %s 失败: %v", tongyiTaskID, err)
				continue
			}
			if queryResp.Output.TaskStatus == "SUCCEEDED" && len(queryResp.Output.Results) > 0 {
				result := queryResp.Output.Results[0]

				publicURL := result.URL
				storageType := models.StorageLocal
				originalPath := result.URL

				if s.qiniuService != nil {
					fileName := fmt.Sprintf("%s_%d", task.UUID, idx)
					qiniuURL, err := s.qiniuService.UploadFromURL(result.URL, fileName)
					if err != nil {
						log.Printf("上传变体 %d 到七牛失败: %v，使用原始URL", idx, err)
					} else {
						publicURL = qiniuURL
						storageType = models.StorageQiniu
						originalPath = s.qiniuService.generateKey(fileName)
					}
				}

				asset := models.CreativeAsset{
					UUIDModel: models.UUIDModel{
						UUID: uuid.New().String(),
					},
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

				if err := database.DB.Create(&asset).Error; err != nil {
					log.Printf("保存变体资产失败: %v", err)
				} else {
					if idx == 0 {
						firstPublicURL = publicURL
					}
				}
				success = true
				if traceID != "" {
					s.tongyiClient.tracer.FinishTrace(traceID, "success", publicURL, "")
				}
				break
			} else if queryResp.Output.TaskStatus == "FAILED" {
				errMsg := queryResp.Output.Message
				if errMsg == "" {
					errMsg = "任务失败"
				}
				database.DB.Model(&task).Updates(map[string]interface{}{
					"status":        models.TaskFailed,
					"error_message": errMsg,
					"progress":      100,
				})
				if traceID != "" {
					s.tongyiClient.tracer.FinishTrace(traceID, "failed", "", errMsg)
				}
				return
			}
		}

		progress := 30 + (idx+1)*40/numVariants
		if success {
			database.DB.Model(&task).Update("progress", progress)
		}
	}

	completedAt := time.Now()
	duration := int(completedAt.Sub(now).Seconds())

	update := map[string]interface{}{
		"status":              models.TaskCompleted,
		"progress":            100,
		"completed_at":        completedAt,
		"processing_duration": duration,
	}
	if firstPublicURL != "" {
		update["first_asset_url"] = firstPublicURL
	}
	if err := database.DB.Model(&task).Updates(update).Error; err != nil {
		log.Printf("更新变体任务完成状态失败: %v", err)
	}
}

// generatePrompt 生成提示词
func (s *CreativeService) generatePrompt(title string, sellingPoints models.StringArray, style string) string {
	prompt := fmt.Sprintf("Product advertisement image for: %s", title)

	if len(sellingPoints) > 0 {
		prompt += fmt.Sprintf(", features: %v", sellingPoints)
	}

	if style != "" {
		styleMap := map[string]string{
			"modern":       "modern and minimalist style",
			"elegant":      "elegant and sophisticated style",
			"vibrant":      "vibrant and energetic style",
			"bright":       "bright and clean style",
			"professional": "professional commercial style",
		}
		if desc, ok := styleMap[style]; ok {
			prompt += ", " + desc
		} else {
			prompt += ", " + style
		}
	}

	prompt += ", high quality, professional photography, clean background, product focused"

	return prompt
}

func statusFromErr(err error) string {
	if err != nil {
		return "failed"
	}
	return "success"
}

func errMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func styleAt(arr models.StringArray, idx int) string {
	if len(arr) == 0 {
		return ""
	}
	if idx >= 0 && idx < len(arr) {
		return arr[idx]
	}
	return arr[0]
}

// GetTask 获取任务详情
func (s *CreativeService) GetTask(taskUUID string) (*models.CreativeTask, error) {
	var task models.CreativeTask
	if err := database.DB.Preload("Assets").Where("uuid = ?", taskUUID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// StartCreativeGeneration 从已确认的文案启动创意生成
type StartCreativeOptions struct {
	ProductImageURL string
	Style           string
	NumVariants     int
	Formats         []string
	VariantPrompts  []string
	VariantStyles   []string
}

// StartCreativeGeneration 从已确认的文案启动创意生成
func (s *CreativeService) StartCreativeGeneration(taskUUID string, opts *StartCreativeOptions) error {
	var task models.CreativeTask
	if err := database.DB.Where("uuid = ?", taskUUID).First(&task).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.CTAText == "" || len(task.SellingPoints) == 0 {
		return errors.New("task missing copywriting data")
	}

	if task.Status == models.TaskProcessing || task.Status == models.TaskQueued {
		return nil
	}

	// 应用可选覆盖配置
	if opts != nil {
		updateMap := map[string]interface{}{}
		if opts.ProductImageURL != "" {
			updateMap["product_image_url"] = opts.ProductImageURL
		}
		if len(opts.Formats) > 0 {
			updateMap["requested_formats"] = models.StringArray(opts.Formats)
		}
		if opts.Style != "" {
			updateMap["requested_styles"] = models.StringArray{opts.Style}
		}
		if len(opts.VariantPrompts) > 0 {
			updateMap["variant_prompts"] = models.StringArray(opts.VariantPrompts)
		}
		if len(opts.VariantStyles) > 0 {
			updateMap["variant_styles"] = models.StringArray(opts.VariantStyles)
		}
		if opts.NumVariants > 0 {
			updateMap["num_variants"] = opts.NumVariants
		}
		if len(updateMap) > 0 {
			if err := database.DB.Model(&task).Updates(updateMap).Error; err != nil {
				return fmt.Errorf("update task options failed: %w", err)
			}
			if err := database.DB.Where("uuid = ?", taskUUID).First(&task).Error; err != nil {
				return fmt.Errorf("reload task failed: %w", err)
			}
		}
	}

	if err := database.DB.Model(&task).Updates(map[string]interface{}{
		"status":   models.TaskQueued,
		"progress": 5,
	}).Error; err != nil {
		return fmt.Errorf("update task status failed: %w", err)
	}

	go s.processTask(task.ID)

	return nil
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
			ID:            asset.UUID,
			NumericID:     asset.ID,
			Format:        asset.Format,
			ImageURL:      getPublicURL(&asset), // 优先使用公共URL
			Width:         asset.Width,
			Height:        asset.Height,
			Title:         asset.Task.Title,
			ProductName:   asset.Task.ProductName,
			CTAText:       asset.Task.CTAText,
			SellingPoints: asset.Task.SellingPoints,
		}
		creativeDatas = append(creativeDatas, creativeData)
	}

	return creativeDatas, total, nil
}

// CreativeAssetDTO 素材数据传输对象
type CreativeAssetDTO struct {
	ID            string             `json:"id"`
	NumericID     uint               `json:"numeric_id"`
	Format        string             `json:"format"`
	ImageURL      string             `json:"image_url"`
	Width         int                `json:"width"`
	Height        int                `json:"height"`
	Title         string             `json:"title,omitempty"`
	ProductName   string             `json:"product_name,omitempty"`
	CTAText       string             `json:"cta_text,omitempty"`
	SellingPoints models.StringArray `json:"selling_points,omitempty"`
}

// getPublicURL 获取公共访问URL
func getPublicURL(asset *models.CreativeAsset) string {
	return asset.PublicURL
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
}

// ListAllTasks 获取所有任务
func (s *CreativeService) ListAllTasks(query ListTasksQuery) ([]TaskDTO, int64, error) {
	var tasks []models.CreativeTask
	var total int64

	// 构建查询
	dbQuery := database.DB.Model(&models.CreativeTask{}).Preload("Assets")

	// 应用筛选条件
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	} else {
		// 默认不展示草稿
		dbQuery = dbQuery.Where("status <> ?", models.TaskDraft)
	}
	if query.UserID > 0 {
		dbQuery = dbQuery.Where("user_id = ?", query.UserID)
	}

	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	if err := dbQuery.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
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
		}
		taskDTOs = append(taskDTOs, taskDTO)
	}

	return taskDTOs, total, nil
}

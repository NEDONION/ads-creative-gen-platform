package handlers

import (
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/services"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreativeHandler 创意处理器
type CreativeHandler struct {
	service            *services.CreativeService
	copywritingService *services.CopywritingService
}

// NewCreativeHandler 创建处理器
func NewCreativeHandler() *CreativeHandler {
	return &CreativeHandler{
		service:            services.NewCreativeService(),
		copywritingService: services.NewCopywritingService(),
	}
}

// GenerateCopywriting 生成文案候选
func (h *CreativeHandler) GenerateCopywriting(c *gin.Context) {
	var req GenerateCopywritingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}

	output, err := h.copywritingService.GenerateCopywriting(services.GenerateCopywritingInput{
		UserID:      1, // TODO: 认证集成后替换
		ProductName: req.ProductName,
		Language:    req.Language,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "Failed to generate copywriting: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(output))
}

// Generate 创建创意生成任务
func (h *CreativeHandler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}

	// 创建任务
	task, err := h.service.CreateTask(services.CreateTaskInput{
		UserID:          1, // TODO: 从认证中获取
		Title:           req.Title,
		SellingPoints:   req.SellingPoints,
		ProductImageURL: req.ProductImageURL,
		Formats:         req.Formats,
		Style:           req.Style,
		CTAText:         req.CTAText,
		NumVariants:     req.NumVariants,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "Failed to create task: "+err.Error()))
		return
	}

	// 返回任务信息
	c.JSON(http.StatusOK, SuccessResponse(TaskData{
		TaskID: task.UUID,
		Status: string(task.Status),
	}))
}

// ConfirmCopywriting 确认文案选择
func (h *CreativeHandler) ConfirmCopywriting(c *gin.Context) {
	var req ConfirmCopywritingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}

	task, err := h.copywritingService.ConfirmCopywriting(services.ConfirmCopywritingInput{
		TaskID:            req.TaskID,
		SelectedCTAIndex:  req.SelectedCTAIndex,
		SelectedSPIndexes: req.SelectedSPIndexes,
		EditedCTA:         req.EditedCTA,
		EditedSPs:         req.EditedSPs,
		ProductImageURL:   req.ProductImageURL,
		Style:             req.Style,
		NumVariants:       req.NumVariants,
		Formats:           req.Formats,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to confirm copywriting: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(TaskData{
		TaskID: task.UUID,
		Status: string(task.Status),
	}))
}

// StartCreative 从确认后的任务启动生成
func (h *CreativeHandler) StartCreative(c *gin.Context) {
	var req StartCreativeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}

	opts := &services.StartCreativeOptions{
		ProductImageURL: req.ProductImageURL,
		Style:           req.Style,
		NumVariants:     req.NumVariants,
		Formats:         req.Formats,
	}
	for _, cfg := range req.VariantConfigs {
		opts.VariantPrompts = append(opts.VariantPrompts, cfg.Prompt)
		opts.VariantStyles = append(opts.VariantStyles, cfg.Style)
	}

	if err := h.service.StartCreativeGeneration(req.TaskID, opts); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to start creative generation: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(TaskData{
		TaskID: req.TaskID,
		Status: string(models.TaskQueued),
	}))
}

// DeleteTask 删除任务及资产
func (h *CreativeHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "task_id is required"))
		return
	}

	if err := h.service.DeleteTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to delete task: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(TaskData{
		TaskID: taskID,
		Status: "deleted",
	}))
}

// GetTask 查询任务状态
func (h *CreativeHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.service.GetTask(taskID)
	if err != nil {
		// 修复：确保返回标准的JSON错误响应
		c.JSON(http.StatusNotFound, ErrorResponse(404, "Task not found"))
		return
	}

	// 构建响应
	data := TaskDetailData{
		TaskID:           task.UUID,
		Status:           string(task.Status),
		Title:            task.Title,
		ProductName:      task.ProductName,
		Progress:         task.Progress,
		SellingPoints:    task.SellingPoints,
		ProductImageURL:  task.ProductImageURL,
		RequestedFormats: task.RequestedFormats,
		Style:            firstStyle(task.RequestedStyles),
		CTAText:          task.CTAText,
		NumVariants:      task.NumVariants,
	}
	data.CreatedAt = task.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	if task.CompletedAt != nil {
		data.CompletedAt = task.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	// 如果有错误信息
	if task.ErrorMessage != "" {
		data.Error = task.ErrorMessage
	}

	// 总是包含资产信息（即使为空）
	creatives := make([]CreativeData, 0, len(task.Assets))
	for _, asset := range task.Assets {
		creatives = append(creatives, CreativeData{
			ID:            asset.UUID,
			Format:        asset.Format,
			ImageURL:      getPublicURL(&asset), // 使用统一的方法获取公共URL
			Width:         asset.Width,
			Height:        asset.Height,
			Title:         asset.Title,
			ProductName:   asset.ProductName,
			CTAText:       asset.CTAText,
			SellingPoints: asset.SellingPoints,
		})
	}
	data.Creatives = creatives

	c.JSON(http.StatusOK, SuccessResponse(data))
}

// getPublicURL 获取公共访问URL
func getPublicURL(asset *models.CreativeAsset) string {
	return asset.PublicURL
}

func firstStyle(styles models.StringArray) string {
	if len(styles) > 0 {
		return styles[0]
	}
	return ""
}

// ListAllAssets 获取所有创意素材
func (h *CreativeHandler) ListAllAssets(c *gin.Context) {
	// 获取查询参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	format := c.Query("format")
	taskID := c.Query("task_id")

	// 转换分页参数
	pageNum := 1
	pageSizeNum := 20

	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageNum = p
	}
	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
		pageSizeNum = ps
	}

	// 构建查询条件
	query := services.ListAssetsQuery{
		Page:     pageNum,
		PageSize: pageSizeNum,
		Format:   format,
		TaskID:   taskID,
	}

	// 获取素材列表
	assets, total, err := h.service.ListAllAssets(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "Failed to fetch assets: "+err.Error()))
		return
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"assets":      assets,
		"total":       total,
		"page":        pageNum,
		"page_size":   pageSizeNum,
		"total_pages": int(math.Ceil(float64(total) / float64(pageSizeNum))),
	}

	c.JSON(http.StatusOK, SuccessResponse(responseData))
}

// ListAllTasks 获取所有任务
func (h *CreativeHandler) ListAllTasks(c *gin.Context) {
	// 获取查询参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	status := c.Query("status")

	// 转换分页参数
	pageNum := 1
	pageSizeNum := 20

	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageNum = p
	}
	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
		pageSizeNum = ps
	}

	// 构建查询条件
	query := services.ListTasksQuery{
		Page:     pageNum,
		PageSize: pageSizeNum,
		Status:   status,
		UserID:   0, // TODO: 从认证中获取
	}

	// 获取任务列表
	tasks, total, err := h.service.ListAllTasks(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(500, "Failed to fetch tasks: "+err.Error()))
		return
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"tasks":       tasks,
		"total":       total,
		"page":        pageNum,
		"page_size":   pageSizeNum,
		"total_pages": int(math.Ceil(float64(total) / float64(pageSizeNum))),
	}

	c.JSON(http.StatusOK, SuccessResponse(responseData))
}

package handlers

import (
	"math"
	"n
	"net/http"
	"math"

	"ads-creative-gen-platform/internal/services"

	"github.com/gin-gonic/gin"
)

// CreativeHandler 创意处理器
type CreativeHandler struct {
	service *services.CreativeService
}

// NewCreativeHandler 创建处理器
func NewCreativeHandler() *CreativeHandler {
	return &CreativeHandler{
		service: services.NewCreativeService(),
	}
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
		TaskID:   task.UUID,
		Status:   string(task.Status),
		Title:    task.Title,
		Progress: task.Progress,
	}

	// 如果有错误信息
	if task.ErrorMessage != "" {
		data.Error = task.ErrorMessage
	}

	// 如果有生成的创意
	if len(task.Assets) > 0 {
		creatives := make([]CreativeData, 0, len(task.Assets))
		for _, asset := range task.Assets {
			creatives = append(creatives, CreativeData{
				ID:       asset.UUID,
				Format:   asset.Format,
				ImageURL: asset.PublicURL,
				Width:    asset.Width,
				Height:   asset.Height,
			})
		}
		data.Creatives = creatives
	}

	c.JSON(http.StatusOK, SuccessResponse(data))
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
		"assets":      assets,
		"total":       total,
		"page":        pageNum,
		"page_size":   pageSizeNum,
		"page_size": pageSizeNum,
		"total_pages": int(math.Ceil(float64(total) / float64(pageSizeNum))),
	}

	c.JSON(http.StatusOK, SuccessResponse(responseData))
}

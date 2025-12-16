package handler

import (
	"net/http"
	"strconv"
	"time"

	"ads-creative-gen-platform/internal/experiment/service"
	"ads-creative-gen-platform/internal/shared"

	"github.com/gin-gonic/gin"
)

type ExperimentHandler struct {
	service *service.ExperimentService
}

func NewExperimentHandler() *ExperimentHandler {
	return &ExperimentHandler{
		service: service.NewExperimentService(),
	}
}

// CreateExperiment 创建实验
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	var req shared.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}

	exp, err := h.service.CreateExperiment(service.CreateExperimentInput{
		Name:        req.Name,
		ProductName: req.ProductName,
		Variants:    req.Variants,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Failed to create experiment: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, shared.SuccessResponse(map[string]interface{}{
		"experiment_id": exp.UUID,
		"status":        exp.Status,
	}))
}

// ListExperiments 获取实验列表
func (h *ExperimentHandler) ListExperiments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	result, err := h.service.ListExperiments(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Failed to list experiments: "+err.Error()))
		return
	}

	type experimentDTO struct {
		ExperimentID string     `json:"experiment_id"`
		Name         string     `json:"name"`
		ProductName  string     `json:"product_name,omitempty"`
		Status       string     `json:"status"`
		CreatedAt    time.Time  `json:"created_at"`
		StartAt      *time.Time `json:"start_at,omitempty"`
		EndAt        *time.Time `json:"end_at,omitempty"`
		Variants     []struct {
			CreativeID    uint     `json:"creative_id"`
			Weight        float64  `json:"weight"`
			BucketStart   int      `json:"bucket_start"`
			BucketEnd     int      `json:"bucket_end"`
			Title         string   `json:"title,omitempty"`
			ProductName   string   `json:"product_name,omitempty"`
			ImageURL      string   `json:"image_url,omitempty"`
			CTAText       string   `json:"cta_text,omitempty"`
			SellingPoints []string `json:"selling_points,omitempty"`
		} `json:"variants,omitempty"`
	}

	resp := struct {
		Experiments []experimentDTO `json:"experiments"`
		Total       int64           `json:"total"`
		Page        int             `json:"page"`
		PageSize    int             `json:"page_size"`
	}{
		Experiments: []experimentDTO{},
		Total:       result.Total,
		Page:        result.Page,
		PageSize:    result.PageSize,
	}

	for _, exp := range result.Experiments {
		item := experimentDTO{
			ExperimentID: exp.UUID,
			Name:         exp.Name,
			ProductName:  exp.ProductName,
			Status:       string(exp.Status),
			CreatedAt:    exp.CreatedAt,
			StartAt:      exp.StartAt,
			EndAt:        exp.EndAt,
		}
		for _, v := range exp.Variants {
			item.Variants = append(item.Variants, struct {
				CreativeID    uint     `json:"creative_id"`
				Weight        float64  `json:"weight"`
				BucketStart   int      `json:"bucket_start"`
				BucketEnd     int      `json:"bucket_end"`
				Title         string   `json:"title,omitempty"`
				ProductName   string   `json:"product_name,omitempty"`
				ImageURL      string   `json:"image_url,omitempty"`
				CTAText       string   `json:"cta_text,omitempty"`
				SellingPoints []string `json:"selling_points,omitempty"`
			}{
				CreativeID:    v.CreativeID,
				Weight:        v.Weight,
				BucketStart:   v.BucketStart,
				BucketEnd:     v.BucketEnd,
				Title:         v.Title,
				ProductName:   v.ProductName,
				ImageURL:      v.ImageURL,
				CTAText:       v.CTAText,
				SellingPoints: v.SellingPoints,
			})
		}
		resp.Experiments = append(resp.Experiments, item)
	}

	c.JSON(http.StatusOK, shared.SuccessResponse(resp))
}

// UpdateStatus 更新状态
func (h *ExperimentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req shared.UpdateExperimentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Failed to update status: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, shared.SuccessResponse(map[string]interface{}{
		"experiment_id": id,
		"status":        req.Status,
	}))
}

// Assign 分流
func (h *ExperimentHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	userKey := c.Query("user_key")
	result, err := h.service.Assign(id, userKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Assign failed: "+err.Error()))
		return
	}
	// 优先使用 variant 存储的元数据快照（用户选择的覆盖值）
	// 只有当 variant 字段为空时，才使用 asset 的默认值
	resp := map[string]interface{}{
		"creative_id": result.Variant.CreativeID,
	}

	// Title
	title := result.Variant.Title
	if title == "" && result.Asset != nil {
		title = result.Asset.Title
	}
	resp["title"] = title

	// Product Name
	productName := result.Variant.ProductName
	if productName == "" && result.Asset != nil {
		productName = result.Asset.ProductName
	}
	resp["product_name"] = productName

	// CTA Text
	ctaText := result.Variant.CTAText
	if ctaText == "" && result.Asset != nil {
		ctaText = result.Asset.CTAText
	}
	resp["cta_text"] = ctaText

	// Selling Points - 关键修复：优先使用 variant 的（用户选择的）
	sellingPoints := result.Variant.SellingPoints
	if len(sellingPoints) == 0 && result.Asset != nil && result.Asset.Task.ID > 0 {
		sellingPoints = result.Asset.Task.SellingPoints
	}
	resp["selling_points"] = sellingPoints

	// Image URL
	imageURL := result.Variant.ImageURL
	if imageURL == "" && result.Asset != nil {
		imageURL = result.Asset.PublicURL
	}
	resp["image_url"] = imageURL

	// Asset 相关信息（如果存在）
	if result.Asset != nil {
		resp["asset_uuid"] = result.Asset.UUID
		resp["task_id"] = result.Asset.TaskID
	}
	c.JSON(http.StatusOK, shared.SuccessResponse(resp))
}

// Hit 曝光
func (h *ExperimentHandler) Hit(c *gin.Context) {
	id := c.Param("id")
	var req shared.TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}
	if err := h.service.Hit(id, req.CreativeID); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Hit failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, shared.SuccessResponse(map[string]string{"status": "ok"}))
}

// Click 点击
func (h *ExperimentHandler) Click(c *gin.Context) {
	id := c.Param("id")
	var req shared.TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}
	if err := h.service.Click(id, req.CreativeID); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Click failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, shared.SuccessResponse(map[string]string{"status": "ok"}))
}

// Metrics 结果
func (h *ExperimentHandler) Metrics(c *gin.Context) {
	id := c.Param("id")
	dto, err := h.service.GetMetrics(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse(400, "Metrics failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, shared.SuccessResponse(dto))
}

package handlers

import (
	"ads-creative-gen-platform/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ExperimentHandler struct {
	service *services.ExperimentService
}

func NewExperimentHandler() *ExperimentHandler {
	return &ExperimentHandler{
		service: services.NewExperimentService(),
	}
}

// CreateExperiment 创建实验
func (h *ExperimentHandler) CreateExperiment(c *gin.Context) {
	var req CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}

	exp, err := h.service.CreateExperiment(services.CreateExperimentInput{
		Name:        req.Name,
		ProductName: req.ProductName,
		Variants:    req.Variants,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to create experiment: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
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
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to list experiments: "+err.Error()))
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

	c.JSON(http.StatusOK, SuccessResponse(resp))
}

// UpdateStatus 更新状态
func (h *ExperimentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateExperimentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}
	if err := h.service.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to update status: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
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
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Assign failed: "+err.Error()))
		return
	}
	resp := map[string]interface{}{
		"creative_id": result.Variant.CreativeID,
	}
	if result.Asset != nil {
		resp["asset_uuid"] = result.Asset.UUID
		resp["task_id"] = result.Asset.TaskID
		resp["title"] = result.Asset.Title
		resp["product_name"] = result.Asset.ProductName
		resp["cta_text"] = result.Asset.CTAText
		resp["selling_points"] = result.Asset.Task.SellingPoints
		resp["image_url"] = result.Asset.PublicURL
	} else {
		// 使用存储在变体表的元数据
		resp["title"] = result.Variant.Title
		resp["product_name"] = result.Variant.ProductName
		resp["cta_text"] = result.Variant.CTAText
		resp["selling_points"] = result.Variant.SellingPoints
		resp["image_url"] = result.Variant.ImageURL
	}
	c.JSON(http.StatusOK, SuccessResponse(resp))
}

// Hit 曝光
func (h *ExperimentHandler) Hit(c *gin.Context) {
	id := c.Param("id")
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}
	if err := h.service.Hit(id, req.CreativeID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Hit failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(map[string]string{"status": "ok"}))
}

// Click 点击
func (h *ExperimentHandler) Click(c *gin.Context) {
	id := c.Param("id")
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Invalid request: "+err.Error()))
		return
	}
	if err := h.service.Click(id, req.CreativeID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Click failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(map[string]string{"status": "ok"}))
}

// Metrics 结果
func (h *ExperimentHandler) Metrics(c *gin.Context) {
	id := c.Param("id")
	dto, err := h.service.GetMetrics(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Metrics failed: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(dto))
}

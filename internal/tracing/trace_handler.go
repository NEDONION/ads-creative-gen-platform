package tracing

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TraceHandler struct {
	service *TraceService
}

func NewTraceHandler() *TraceHandler {
	return &TraceHandler{
		service: NewTraceService(),
	}
}

// Service 暴露 service 供预热使用
func (h *TraceHandler) Service() *TraceService {
	return h.service
}

// List traces
func (h *TraceHandler) ListTraces(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	modelName := c.Query("model_name")
	traceID := c.Query("trace_id")
	productName := c.Query("product_name")

	result, err := h.service.List(page, pageSize, status, modelName, traceID, productName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Failed to list traces: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": map[string]interface{}{
			"traces":    result.Traces,
			"total":     result.Total,
			"page":      result.Page,
			"page_size": result.PageSize,
		},
	})
}

// Trace detail
func (h *TraceHandler) GetTrace(c *gin.Context) {
	traceID := c.Param("id")
	trace, err := h.service.Detail(traceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "trace not found: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": trace,
	})
}

// ForceFail 手动标记 trace 为失败
func (h *TraceHandler) ForceFail(c *gin.Context) {
	traceID := c.Param("id")
	reason := c.DefaultPostForm("reason", "manually marked as failed")
	if err := h.service.ForceFail(traceID, reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "fail trace failed: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
	})
}

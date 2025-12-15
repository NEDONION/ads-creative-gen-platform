package handlers

import (
	"ads-creative-gen-platform/internal/tracing"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TraceHandler struct {
	service *tracing.TraceService
}

func NewTraceHandler() *TraceHandler {
	return &TraceHandler{
		service: tracing.NewTraceService(),
	}
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
		c.JSON(http.StatusBadRequest, ErrorResponse(400, "Failed to list traces: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"traces":    result.Traces,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	}))
}

// Trace detail
func (h *TraceHandler) GetTrace(c *gin.Context) {
	traceID := c.Param("id")
	trace, err := h.service.Detail(traceID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(404, "trace not found: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(trace))
}

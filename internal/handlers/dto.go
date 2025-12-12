package handlers

// GenerateRequest 创意生成请求
type GenerateRequest struct {
	Title           string   `json:"title" binding:"required"`
	SellingPoints   []string `json:"selling_points"`
	ProductImageURL string   `json:"product_image_url"`
	Formats         []string `json:"formats"`
	Style           string   `json:"style"`
	CTAText         string   `json:"cta_text"`
	NumVariants     int      `json:"num_variants"`
}

// GenerateResponse 创意生成响应
type GenerateResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// TaskData 任务数据
type TaskData struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// TaskDetailData 任务详情数据
type TaskDetailData struct {
	TaskID    string         `json:"task_id"`
	Status    string         `json:"status"`
	Title     string         `json:"title"`
	Progress  int            `json:"progress"`
	Creatives []CreativeData `json:"creatives,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// CreativeData 创意数据
type CreativeData struct {
	ID       string  `json:"id"`
	Format   string  `json:"format"`
	ImageURL string  `json:"image_url"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Score    float64 `json:"score,omitempty"`
	Rank     int     `json:"rank,omitempty"`
}

// Response 工具函数
func SuccessResponse(data interface{}) GenerateResponse {
	return GenerateResponse{
		Code: 0,
		Data: data,
	}
}

func ErrorResponse(code int, message string) GenerateResponse {
	return GenerateResponse{
		Code:    code,
		Message: message,
	}
}

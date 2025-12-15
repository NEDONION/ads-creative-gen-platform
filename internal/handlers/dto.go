package handlers

import (
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/services"
)

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

// GenerateCopywritingRequest 文案生成请求
type GenerateCopywritingRequest struct {
	ProductName string `json:"product_name" binding:"required"`
	Language    string `json:"language,omitempty"`
}

// ConfirmCopywritingRequest 文案确认请求
type ConfirmCopywritingRequest struct {
	TaskID            string   `json:"task_id" binding:"required"`
	SelectedCTAIndex  int      `json:"selected_cta_index"`
	SelectedSPIndexes []int    `json:"selected_sp_indexes"`
	EditedCTA         string   `json:"edited_cta,omitempty"`
	EditedSPs         []string `json:"edited_sps,omitempty"`
	ProductImageURL   string   `json:"product_image_url,omitempty"`
	Style             string   `json:"style,omitempty"`
	NumVariants       int      `json:"num_variants"`
	Formats           []string `json:"formats"`
}

// StartCreativeRequest 启动创意生成请求
type StartCreativeRequest struct {
	TaskID          string              `json:"task_id" binding:"required"`
	ProductImageURL string              `json:"product_image_url,omitempty"`
	Style           string              `json:"style,omitempty"`
	NumVariants     int                 `json:"num_variants,omitempty"`
	Formats         []string            `json:"formats,omitempty"`
	VariantConfigs  []TaskVariantConfig `json:"variant_configs,omitempty"`
}

// TaskVariantConfig 单个变体配置
type TaskVariantConfig struct {
	Style  string `json:"style,omitempty"`
	Prompt string `json:"prompt,omitempty"`
}

// Experiment DTOs
type CreateExperimentRequest struct {
	Name        string                            `json:"name" binding:"required"`
	ProductName string                            `json:"product_name"`
	Variants    []services.ExperimentVariantInput `json:"variants" binding:"required"`
}

type UpdateExperimentStatusRequest struct {
	Status models.ExperimentStatus `json:"status" binding:"required"`
}

type TrackRequest struct {
	CreativeID uint `json:"creative_id" binding:"required"`
}

// TaskData 任务数据
type TaskData struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// TaskDetailData 任务详情数据
type TaskDetailData struct {
	TaskID           string         `json:"task_id"`
	Status           string         `json:"status"`
	Title            string         `json:"title"`
	ProductName      string         `json:"product_name,omitempty"`
	Progress         int            `json:"progress"`
	Creatives        []CreativeData `json:"creatives,omitempty"`
	Error            string         `json:"error,omitempty"`
	SellingPoints    []string       `json:"selling_points,omitempty"`
	ProductImageURL  string         `json:"product_image_url,omitempty"`
	RequestedFormats []string       `json:"requested_formats,omitempty"`
	Style            string         `json:"style,omitempty"`
	CTAText          string         `json:"cta_text,omitempty"`
	NumVariants      int            `json:"num_variants,omitempty"`
	CreatedAt        string         `json:"created_at,omitempty"`
	CompletedAt      string         `json:"completed_at,omitempty"`
	VariantPrompts   []string       `json:"variant_prompts,omitempty"`
	VariantStyles    []string       `json:"variant_styles,omitempty"`
}

// CreativeData 创意数据
type CreativeData struct {
	ID               string   `json:"id"`
	Format           string   `json:"format"`
	ImageURL         string   `json:"image_url"`
	Width            int      `json:"width"`
	Height           int      `json:"height"`
	Score            float64  `json:"score,omitempty"`
	Rank             int      `json:"rank,omitempty"`
	Title            string   `json:"title,omitempty"`
	ProductName      string   `json:"product_name,omitempty"`
	CTAText          string   `json:"cta_text,omitempty"`
	SellingPoints    []string `json:"selling_points,omitempty"`
	Style            string   `json:"style,omitempty"`
	GenerationPrompt string   `json:"generation_prompt,omitempty"`
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

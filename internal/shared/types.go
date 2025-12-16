package shared

import (
	expuc "ads-creative-gen-platform/internal/experiment/service"
	"ads-creative-gen-platform/internal/models"
)

// ===== Query DTOs =====

type ListTasksQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	UserID   uint   `json:"user_id"`
}

type ListAssetsQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Format   string `json:"format"`
	TaskID   string `json:"task_id"`
}

// ===== LLM / VLM DTOs =====

type ImageGenResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
		Message string `json:"message"`
	} `json:"output"`
	RequestID string `json:"request_id"`
}

type QueryResp struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Message    string `json:"message"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
	} `json:"output"`
	RequestID string `json:"request_id"`
}

type CopywritingResult struct {
	CTAOptions          []string `json:"cta_options"`
	SellingPointOptions []string `json:"selling_point_options"`
	RawResponse         string   `json:"-"`
}

// ===== API DTOs =====

type GenerateRequest struct {
	Title           string   `json:"title" binding:"required"`
	SellingPoints   []string `json:"selling_points"`
	ProductImageURL string   `json:"product_image_url"`
	Formats         []string `json:"formats"`
	Style           string   `json:"style"`
	CTAText         string   `json:"cta_text"`
	NumVariants     int      `json:"num_variants"`
}

type GenerateResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type GenerateCopywritingRequest struct {
	ProductName string `json:"product_name" binding:"required"`
	Language    string `json:"language,omitempty"`
}

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

type StartCreativeRequest struct {
	TaskID          string              `json:"task_id" binding:"required"`
	ProductImageURL string              `json:"product_image_url,omitempty"`
	Style           string              `json:"style,omitempty"`
	NumVariants     int                 `json:"num_variants,omitempty"`
	Formats         []string            `json:"formats,omitempty"`
	VariantConfigs  []TaskVariantConfig `json:"variant_configs,omitempty"`
}

type TaskVariantConfig struct {
	Style  string `json:"style,omitempty"`
	Prompt string `json:"prompt,omitempty"`
}

type CreateExperimentRequest struct {
	Name        string                         `json:"name" binding:"required"`
	ProductName string                         `json:"product_name"`
	Variants    []expuc.ExperimentVariantInput `json:"variants" binding:"required"`
}

type UpdateExperimentStatusRequest struct {
	Status models.ExperimentStatus `json:"status" binding:"required"`
}

type TrackRequest struct {
	CreativeID uint `json:"creative_id" binding:"required"`
}

type TaskData struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

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

// Response helpers
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

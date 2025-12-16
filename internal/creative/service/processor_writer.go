package service

import (
	"context"
	"fmt"
	"log"

	"ads-creative-gen-platform/internal/infra/llm"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/settings"

	"github.com/google/uuid"
)

type GenResult struct {
	FirstPublicURL string
	Count          int
}

func (p *TaskProcessor) persistAssets(
	ctx context.Context,
	task *models.CreativeTask,
	req GenRequest,
	queryResp *llm.ImageGenResponse,
) (GenResult, error) {
	if len(queryResp.Output.Results) == 0 {
		return GenResult{}, fmt.Errorf("任务成功但未返回结果")
	}

	var first string
	count := 0

	for i, result := range queryResp.Output.Results {
		publicURL, storageType, originalPath := p.handleUpload(ctx, task.UUID, req.VariantIndex*1000+i, result.URL)

		idx := req.VariantIndex
		asset := models.CreativeAsset{
			UUIDModel:        models.UUIDModel{UUID: uuid.New().String()},
			TaskID:           task.ID,
			Title:            task.Title,
			ProductName:      task.ProductName,
			CTAText:          task.CTAText,
			SellingPoints:    task.SellingPoints,
			Format:           req.Format,
			Width:            settings.DefaultImageWidth,
			Height:           settings.DefaultImageHeight,
			StorageType:      storageType,
			PublicURL:        publicURL,
			OriginalPath:     originalPath,
			Style:            req.Style,
			VariantIndex:     &idx,
			GenerationPrompt: req.Prompt,
			ModelName:        settings.ModelName,
		}

		if err := p.assetRepo.Create(ctx, &asset); err != nil {
			log.Printf("保存资产失败: %v", err)
			continue
		}

		if first == "" {
			first = publicURL
		}
		count++
	}

	if count == 0 {
		return GenResult{}, fmt.Errorf("资产保存失败")
	}

	return GenResult{FirstPublicURL: first, Count: count}, nil
}

// handleUpload 处理存储上传并返回最终 URL/存储信息。
func (p *TaskProcessor) handleUpload(ctx context.Context, taskUUID string, idx int, originalURL string) (string, models.StorageType, string) {
	publicURL := originalURL
	storageType := models.StorageLocal
	originalPath := originalURL

	if p.storageClient == nil {
		return publicURL, storageType, originalPath
	}

	fileName := fmt.Sprintf("%s_%d", taskUUID, idx)
	qiniuURL, err := p.storageClient.UploadFromURL(ctx, originalURL, fileName)
	if err != nil {
		log.Printf("上传到存储失败: %v，使用原始URL", err)
		return publicURL, storageType, originalPath
	}

	publicURL = qiniuURL
	storageType = models.StorageQiniu
	originalPath = p.storageClient.GenerateKey(fileName)
	return publicURL, storageType, originalPath
}

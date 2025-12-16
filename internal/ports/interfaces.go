package ports

import (
	"context"

	"ads-creative-gen-platform/internal/infra/llm"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/shared"
)

// ===== LLM / VLM Clients =====

type TongyiClient interface {
	GenerateImage(ctx context.Context, prompt, size string, numImages int, productName, traceID, taskUUID string) (*llm.ImageGenResponse, string, error)
	GenerateImageWithProduct(ctx context.Context, prompt, productImageURL, size string, numImages int, productName, traceID, taskUUID string) (*llm.ImageGenResponse, string, error)
	QueryTask(ctx context.Context, traceID, taskID, taskUUID string) (*llm.ImageGenResponse, error)
	FinishTrace(traceID, status, url, msg string)
}

type QwenClient interface {
	GenerateCopywriting(productName string, language string) (*llm.CopywritingResult, error)
}

// ===== Storage =====

type QiniuClient interface {
	UploadFromURL(ctx context.Context, url, fileName string) (string, error)
	GenerateKey(fileName string) string
}

type StorageUploader interface {
	UploadFromURL(ctx context.Context, url, fileName string) (string, error)
	GenerateKey(fileName string) string
}

// ===== Repositories =====

type TaskRepository interface {
	Create(ctx context.Context, task *models.CreativeTask) error
	GetByID(ctx context.Context, id uint) (*models.CreativeTask, error)
	GetByUUID(ctx context.Context, uuid string) (*models.CreativeTask, error)
	GetByUUIDWithAssets(ctx context.Context, uuid string) (*models.CreativeTask, error)
	UpdateStatus(ctx context.Context, id uint, status models.TaskStatus, progress int) error
	UpdateProgress(ctx context.Context, id uint, progress int) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	List(ctx context.Context, query shared.ListTasksQuery) ([]models.CreativeTask, int64, error)
	Delete(ctx context.Context, task *models.CreativeTask) error
}

type AssetRepository interface {
	Create(ctx context.Context, asset *models.CreativeAsset) error
	List(ctx context.Context, query shared.ListAssetsQuery) ([]models.CreativeAsset, int64, error)
	DeleteByTaskID(ctx context.Context, taskID uint) error
}

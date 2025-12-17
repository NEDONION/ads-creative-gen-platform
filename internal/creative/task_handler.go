package creative

import (
	"context"
	"fmt"

	"ads-creative-gen-platform/internal/creative/service"
	"ads-creative-gen-platform/internal/models"
)

// CreativeGenerateTask 将创意执行逻辑封装为 Task
type CreativeGenerateTask struct {
	service *service.CreativeService
	taskID  uint
}

// NewCreativeGenerateTask 构造
func NewCreativeGenerateTask(svc *service.CreativeService, taskID uint) *CreativeGenerateTask {
	return &CreativeGenerateTask{
		service: svc,
		taskID:  taskID,
	}
}

// Execute 执行任务
func (t *CreativeGenerateTask) Execute(ctx context.Context) error {
	if t.service == nil {
		return fmt.Errorf("service not set")
	}
	t.service.ProcessTaskWithContext(ctx, t.taskID)
	return nil
}

// ID 返回任务ID
func (t *CreativeGenerateTask) ID() uint { return t.taskID }

// ToDTO 将模型转换为 DTO（便于 handler 使用）
func ToDTO(task models.CreativeTask) service.TaskDTO {
	completedAt := ""
	if task.CompletedAt != nil {
		completedAt = task.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	firstImage := ""
	if task.FirstAssetURL != "" {
		firstImage = task.FirstAssetURL
	} else if len(task.Assets) > 0 {
		firstImage = task.Assets[0].PublicURL
	}

	return service.TaskDTO{
		ID:            task.UUID,
		Title:         task.Title,
		ProductName:   task.ProductName,
		CTAText:       task.CTAText,
		SellingPoints: task.SellingPoints,
		Status:        string(task.Status),
		Progress:      task.Progress,
		CreatedAt:     task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt:   completedAt,
		ErrorMessage:  task.ErrorMessage,
		FirstImage:    firstImage,
		RetryFrom:     task.RetryFrom,
		RetryTo:       task.RetryTo,
	}
}

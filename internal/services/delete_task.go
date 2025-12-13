package services

import (
	"fmt"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"
)

// DeleteTask 删除任务及其资产（软删除）
func (s *CreativeService) DeleteTask(taskUUID string) error {
	tx := database.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	var task models.CreativeTask
	if err := tx.Where("uuid = ?", taskUUID).First(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("task not found: %w", err)
	}

	// 删除资产
	if err := tx.Where("task_id = ?", task.ID).Delete(&models.CreativeAsset{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete assets failed: %w", err)
	}

	// 删除任务
	if err := tx.Delete(&task).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete task failed: %w", err)
	}

	return tx.Commit().Error
}

//go:build integration
// +build integration

package service

import (
	"testing"

	"ads-creative-gen-platform/internal/creative/repository"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/testutil"

	"github.com/google/uuid"
)

// TestIntegration_Creative_CreateListDelete 验证创意任务创建、列表、删除流程（不触发外部生成）。
func TestIntegration_Creative_CreateListDelete(t *testing.T) {
	testutil.EnsureIntegrationDB(t)
	testutil.ResetTables(t, []string{
		"TRUNCATE creative_assets CASCADE",
		"TRUNCATE creative_tasks CASCADE",
	})

	taskRepo := repository.NewTaskRepository(testutil.DB())
	assetRepo := repository.NewAssetRepository(testutil.DB())

	svc := NewCreativeServiceWithDeps(nil, nil, taskRepo, assetRepo, func(uint) error { return nil })

	task, err := svc.CreateTask(CreateTaskInput{
		UserID:        1,
		Title:         "集成测试任务",
		SellingPoints: []string{"亮点A", "亮点B"},
		Formats:       []string{"1:1"},
		Style:         "modern",
		CTAText:       "buy now",
	})
	if err != nil {
		t.Fatalf("CreateTask 失败: %v", err)
	}

	tasks, total, err := svc.ListAllTasks(ListTasksQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAllTasks 失败: %v", err)
	}
	if total != 1 || len(tasks) != 1 || tasks[0].ID != task.UUID {
		t.Fatalf("任务列表结果不符合预期: total=%d len=%d data=%#v", total, len(tasks), tasks)
	}

	if err := svc.DeleteTask(task.UUID); err != nil {
		t.Fatalf("DeleteTask 失败: %v", err)
	}

	tasks, total, err = svc.ListAllTasks(ListTasksQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListAllTasks 失败: %v", err)
	}
	if total != 0 || len(tasks) != 0 {
		t.Fatalf("删除后列表不为空: total=%d len=%d", total, len(tasks))
	}
}

// TestIntegration_Creative_StartOnly 验证 StartCreativeGeneration 仅更新任务字段并入队（不触发生成）。
func TestIntegration_Creative_StartOnly(t *testing.T) {
	testutil.EnsureIntegrationDB(t)
	testutil.ResetTables(t, []string{
		"TRUNCATE creative_tasks CASCADE",
		"TRUNCATE creative_assets CASCADE",
	})

	taskRepo := repository.NewTaskRepository(testutil.DB())
	assetRepo := repository.NewAssetRepository(testutil.DB())
	var enqueued uint

	svc := NewCreativeServiceWithDeps(nil, nil, taskRepo, assetRepo, func(id uint) error { enqueued = id; return nil })

	task := models.CreativeTask{
		UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
		Title:     "start-test",
		Status:    models.TaskDraft,
	}
	if err := testutil.DB().Create(&task).Error; err != nil {
		t.Fatalf("预置任务失败: %v", err)
	}

	err := svc.StartCreativeGeneration(task.UUID, &StartCreativeOptions{Style: "modern", NumVariants: 2})
	if err != nil {
		t.Fatalf("StartCreativeGeneration 失败: %v", err)
	}
	if enqueued != task.ID {
		t.Fatalf("任务未入队，期望 %d 得到 %d", task.ID, enqueued)
	}
}

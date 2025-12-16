//go:build integration

package integration

import (
	"strconv"
	"testing"

	"ads-creative-gen-platform/internal/experiment/repository"
	"ads-creative-gen-platform/internal/experiment/service"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/testutil"

	"github.com/google/uuid"
)

// ensureAsset 创建可用的素材及其所属任务。
func ensureAsset(t *testing.T) models.CreativeAsset {
	t.Helper()

	user := testutil.CreateTestUser(t)
	task := models.CreativeTask{
		UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
		UserID:    user.ID,
		Title:     "asset-task",
		Status:    models.TaskDraft,
	}
	if err := testutil.DB().Create(&task).Error; err != nil {
		t.Fatalf("预置任务失败: %v", err)
	}

	asset := models.CreativeAsset{
		UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
		TaskID:    task.ID,
		Format:    "1:1",
		Width:     1024,
		Height:    1024,
		PublicURL: "https://example.com/img.png",
	}
	if err := testutil.DB().Create(&asset).Error; err != nil {
		t.Fatalf("创建素材失败: %v", err)
	}
	return asset
}

// TestExperiment_Flow 覆盖创建实验、状态变更、分流、埋点、指标汇总的端到端流程（依赖真实数据库）。
func TestExperiment_Flow(t *testing.T) {
	testutil.ResetTables(t, []string{
		"TRUNCATE experiment_metrics CASCADE",
		"TRUNCATE experiment_variants CASCADE",
		"TRUNCATE experiments CASCADE",
		"TRUNCATE creative_assets CASCADE",
		"TRUNCATE creative_tasks CASCADE",
		"TRUNCATE users CASCADE",
	})

	asset := ensureAsset(t)

	repo := repository.NewExperimentRepository()
	svc := service.NewExperimentServiceWithRepo(repo)

	exp, err := svc.CreateExperiment(service.CreateExperimentInput{
		Name:        "integration-exp",
		ProductName: "Test Product",
		Variants: []service.ExperimentVariantInput{
			{CreativeID: uuidOrID(asset), Weight: 0.6},
			{CreativeID: uuidOrID(asset), Weight: 0.4},
		},
	})
	if err != nil {
		t.Fatalf("CreateExperiment 失败: %v", err)
	}

	if err := svc.UpdateStatus(exp.UUID, models.ExpActive); err != nil {
		t.Fatalf("UpdateStatus 失败: %v", err)
	}

	assigned, err := svc.Assign(exp.UUID, "user-key-1")
	if err != nil {
		t.Fatalf("Assign 失败: %v", err)
	}
	if assigned == nil || assigned.Asset == nil {
		t.Fatalf("Assign 返回为空或缺少素材: %#v", assigned)
	}

	if err := svc.Hit(exp.UUID, assigned.Variant.CreativeID); err != nil {
		t.Fatalf("Hit 失败: %v", err)
	}
	if err := svc.Click(exp.UUID, assigned.Variant.CreativeID); err != nil {
		t.Fatalf("Click 失败: %v", err)
	}

	dto, err := svc.GetMetrics(exp.UUID)
	if err != nil {
		t.Fatalf("GetMetrics 失败: %v", err)
	}
	metrics, ok := dto.(*service.ExperimentMetricsDTO)
	if !ok {
		t.Fatalf("Metrics 返回类型不符: %T", dto)
	}
	if len(metrics.Variants) == 0 {
		t.Fatalf("期望有指标数据，实际为空")
	}
	if metrics.Variants[0].Impressions != 1 || metrics.Variants[0].Clicks != 1 {
		t.Fatalf("曝光/点击次数不符合预期: %#v", metrics.Variants[0])
	}
	if metrics.Variants[0].CTR != 1.0 {
		t.Fatalf("期望 CTR=1.0，实际 %v", metrics.Variants[0].CTR)
	}
}

// uuidOrID 返回优先 UUID，其次数字 ID 的字符串表示，供 CreateExperimentInput 使用。
func uuidOrID(asset models.CreativeAsset) string {
	if asset.UUID != "" {
		return asset.UUID
	}
	return strconv.FormatUint(uint64(asset.ID), 10)
}

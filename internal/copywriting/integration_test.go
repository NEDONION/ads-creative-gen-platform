//go:build integration
// +build integration

package copywriting

import (
	"testing"

	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/internal/testutil"

	"github.com/google/uuid"
)

// TestIntegration_ConfirmCopywriting 验证确认文案流程更新任务字段（不调用外部 LLM）。
func TestIntegration_ConfirmCopywriting(t *testing.T) {
	testutil.EnsureIntegrationDB(t)
	testutil.ResetTables(t, []string{
		"TRUNCATE creative_assets CASCADE",
		"TRUNCATE creative_tasks CASCADE",
		"TRUNCATE users CASCADE",
	})

	user := testutil.CreateTestUser(t)

	task := models.CreativeTask{
		UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
		UserID:    user.ID,
		Title:     "cw-test",
		Status:    models.TaskDraft,
		CTACandidates: models.StringArray{
			"cta1", "cta2",
		},
		SellingPointCandidates: models.StringArray{
			"sp1", "sp2",
		},
	}
	if err := testutil.DB().Create(&task).Error; err != nil {
		t.Fatalf("预置任务失败: %v", err)
	}

	svc := &CopywritingService{} // 不调用 Generate，不需要 LLM
	updated, err := svc.ConfirmCopywriting(ConfirmCopywritingInput{
		TaskID:            task.UUID,
		SelectedCTAIndex:  1,
		SelectedSPIndexes: []int{0, 1},
		Formats:           []string{"1:1"},
		NumVariants:       2,
	})
	if err != nil {
		t.Fatalf("ConfirmCopywriting 失败: %v", err)
	}
	if updated.CTAText != "cta2" {
		t.Fatalf("CTA 未更新，得到 %s", updated.CTAText)
	}
	if len(updated.SellingPoints) != 2 {
		t.Fatalf("卖点数量不符: %v", updated.SellingPoints)
	}
}

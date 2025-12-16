//go:build integration

package integration

import (
	"os"
	"testing"

	"ads-creative-gen-platform/internal/testutil"
)

// TestMain 统一初始化集成测试数据库，缺少环境变量时直接跳过。
func TestMain(m *testing.M) {
	if os.Getenv("DB_NAME") == "" {
		os.Exit(0)
	}
	testutil.EnsureIntegrationDB(nil)
	os.Exit(m.Run())
}

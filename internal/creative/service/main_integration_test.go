//go:build integration
// +build integration

package service

import (
	"os"
	"testing"

	"ads-creative-gen-platform/internal/testutil"
)

// TestMain 仅在 integration tag 下运行，确保 DB 初始化一次。
func TestMain(m *testing.M) {
	if os.Getenv("DB_NAME") == "" {
		os.Exit(0)
	}
	testutil.EnsureIntegrationDB(nil)
	os.Exit(m.Run())
}

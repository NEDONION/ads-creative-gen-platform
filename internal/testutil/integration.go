package testutil

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/models"
	"ads-creative-gen-platform/pkg/database"

	"gorm.io/gorm"
)

var once sync.Once

// EnsureIntegrationDB 确保集成测试的数据库已初始化；若 t 为 nil 则不 skip（用于 TestMain）。
func EnsureIntegrationDB(t *testing.T) {
	if t != nil {
		t.Helper()
	}
	if t != nil && os.Getenv("DB_NAME") == "" {
		t.Skip("缺少 DB_NAME，跳过集成测试")
	}
	once.Do(func() {
		config.LoadConfig()
		database.InitDatabase()
		database.MigrateTables()
	})
}

// DB 返回初始化后的全局 *gorm.DB
func DB() *gorm.DB {
	return database.DB
}

// ResetTables 执行给定的 TRUNCATE/DELETE 语句，便于集成测试复用。
func ResetTables(t *testing.T, stmts []string) {
	t.Helper()
	for _, stmt := range stmts {
		if err := database.DB.Exec(stmt).Error; err != nil {
			t.Fatalf("重置表失败 %s: %v", stmt, err)
		}
	}
}

// CreateTestUser 创建并返回一个测试用户，避免各测试重复造用户。
func CreateTestUser(t *testing.T) *models.User {
	t.Helper()
	u := &models.User{
		UUIDModel:    models.UUIDModel{UUID: fmt.Sprintf("test-%d", time.Now().UnixNano())},
		Email:        fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Username:     fmt.Sprintf("test-%d", time.Now().UnixNano()),
		PasswordHash: "test",
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	return u
}

package database

import (
	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
	dsn := config.GetDatabaseDSN()
	dbType := config.DatabaseConfig.Db

	var dialector gorm.Dialector
	if dbType == "postgres" {
		dialector = postgres.Open(dsn)
	} else {
		dialector = mysql.Open(dsn)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		log.Fatalf("✗ Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("✗ Failed to get database instance: %v", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("✓ Database connected successfully (Type: %s)", dbType)
}

// InitMySQL 初始化数据库连接（为了向后兼容）
func InitMySQL() {
	InitDatabase()
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() {
	log.Println("Starting database migration...")

	err := DB.AutoMigrate(
		// 用户相关
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},

		// 创意相关
		&models.CreativeTask{},
		&models.CreativeAsset{},
		&models.CreativeScore{},

		// 标签
		&models.Tag{},
	)

	if err != nil {
		log.Fatalf("✗ Database migration failed: %v", err)
	}

	log.Println("✓ Database migration completed")
}

// SeedDefaultData 初始化默认数据
func SeedDefaultData() {
	log.Println("Seeding default data...")

	// 检查是否已有数据
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("⊘ Default data already exists, skipping seed")
		return
	}

	// 创建默认管理员
	adminUser := models.User{
		UUIDModel: models.UUIDModel{
			UUID: "admin-uuid-0000-0000-000000000001",
		},
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // admin123
		Role:         models.RoleAdmin,
		Status:       models.StatusActive,
	}

	if err := DB.Create(&adminUser).Error; err != nil {
		log.Printf("⚠ Failed to create admin user: %v", err)
	} else {
		log.Println("✓ Default admin user created (username: admin, password: admin123)")
	}

	// 创建默认标签
	defaultTags := []models.Tag{
		{Name: "电商", Category: "industry", Color: "#FF6B6B"},
		{Name: "游戏", Category: "industry", Color: "#4ECDC4"},
		{Name: "金融", Category: "industry", Color: "#45B7D1"},
		{Name: "教育", Category: "industry", Color: "#FFA07A"},
		{Name: "极简风", Category: "style", Color: "#95E1D3"},
		{Name: "活力风", Category: "style", Color: "#F38181"},
		{Name: "专业风", Category: "style", Color: "#AA96DA"},
	}

	for _, tag := range defaultTags {
		if err := DB.Create(&tag).Error; err != nil {
			log.Printf("⚠ Failed to create tag %s: %v", tag.Name, err)
		}
	}

	log.Println("✓ Default tags created")
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
			log.Println("✓ Database connection closed")
		}
	}
}

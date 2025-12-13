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
	log.Println("正在初始化数据库连接...")

	dsn := config.GetDatabaseDSN()
	log.Printf("数据库DSN: %s", dsn) // 添加DSN日志

	dbType := config.DatabaseConfig.Db
	log.Printf("数据库类型: %s", dbType) // 添加数据库类型日志

	var dialector gorm.Dialector
	if dbType == "postgres" {
		dialector = postgres.Open(dsn)
		log.Println("使用PostgreSQL驱动")
	} else {
		dialector = mysql.Open(dsn)
		log.Println("使用MySQL驱动")
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
		return
	}

	if DB == nil {
		log.Fatalf("✗ GORM Open 返回了 nil DB")
		return
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("✗ Failed to get database instance: %v", err)
		return
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("✓ Database connected successfully (Type: %s)", dbType)
}

// MigrateTables 自动迁移数据库表
func MigrateTables() {
	if DB == nil {
		log.Fatal("✗ Database not initialized, DB is nil")
		return
	}

	log.Println("开始数据库迁移...")

	// 先迁移基础表，再迁移关联表
	tables := []interface{}{
		// 基础表
		&models.User{},
		&models.Tag{},
		&models.Project{},

		// 创意相关表
		&models.CreativeTask{},
		&models.CreativeAsset{},  // 这个表包含我们修改的字段
		&models.CreativeScore{},
		// 实验相关表
		&models.Experiment{},
		&models.ExperimentVariant{},
		&models.ExperimentMetric{},

		// 关系表
		&models.ProjectMember{},
	}

	for _, table := range tables {
		if err := DB.AutoMigrate(table); err != nil {
			log.Printf("✗ 迁移 %T 表失败: %v", table, err)
			// 不直接退出，继续尝试迁移其他表
		} else {
			log.Printf("✓ %T 表迁移完成", table)
		}
	}

	log.Println("✓ 数据库迁移完成")
}

// InitializeDatabase 初始化数据库：迁移表结构并添加默认数据
func InitializeDatabase() {
	log.Println("开始初始化数据库...")

	// 首先初始化数据库连接
	InitDatabase()

	// 迁移表结构
	MigrateTables()

	// 添加默认数据
	SeedDefaultData()

	log.Println("✓ 数据库初始化完成")
}

// SeedDefaultData 初始化默认数据
func SeedDefaultData() {
	if DB == nil {
		log.Fatal("✗ Database not initialized, DB is nil")
		return
	}

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

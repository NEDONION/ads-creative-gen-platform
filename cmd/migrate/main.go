package main

import (
	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/pkg/database"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 定义命令行参数
	action := flag.String("action", "migrate", "Action to perform: migrate, seed, reset")
	flag.Parse()

	// 加载配置
	config.LoadConfig()

	// 初始化数据库连接
	database.InitDatabase()
	defer database.CloseDB()

	// 执行操作
	switch *action {
	case "migrate":
		fmt.Println("═══════════════════════════════════════")
		fmt.Println("  Database Migration")
		fmt.Println("═══════════════════════════════════════")
		database.AutoMigrate()
		fmt.Println("\n✓ Migration completed successfully!")

	case "seed":
		fmt.Println("═══════════════════════════════════════")
		fmt.Println("  Seed Default Data")
		fmt.Println("═══════════════════════════════════════")
		database.SeedDefaultData()
		fmt.Println("\n✓ Seeding completed successfully!")

	case "reset":
		fmt.Println("═══════════════════════════════════════")
		fmt.Println("  ⚠  RESET DATABASE (DROP ALL TABLES)")
		fmt.Println("═══════════════════════════════════════")
		fmt.Print("Are you sure? This will delete ALL data! (yes/no): ")

		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" {
			fmt.Println("✗ Reset cancelled")
			os.Exit(0)
		}

		// 执行重置
		database.AutoMigrate()
		fmt.Println("✓ Database reset completed!")

		// 重新初始化默认数据
		database.SeedDefaultData()
		fmt.Println("✓ Default data seeded!")

	default:
		fmt.Printf("✗ Unknown action: %s\n", *action)
		fmt.Println("Available actions: migrate, seed, reset")
		os.Exit(1)
	}
}

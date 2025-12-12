package main

import (
	"fmt"
	"log"

	"ads-creative-gen-platform/config"
)

func main() {
	// 加载配置
	config.LoadConfig()

	fmt.Println("=== AI 多尺寸广告创意生成平台 ===")
	fmt.Printf("Environment: %s\n", config.AppConfig.Environment)
	fmt.Printf("Server will run on port: %s\n", config.AppConfig.ServerPort)

	// 验证通义 API Key 已加载（只显示前后几位）
	apiKey := config.AppConfig.TongyiAPIKey
	if len(apiKey) > 10 {
		maskedKey := apiKey[:3] + "..." + apiKey[len(apiKey)-4:]
		fmt.Printf("Tongyi API Key loaded: %s\n", maskedKey)
	} else {
		log.Fatal("Invalid API Key format")
	}

	fmt.Println("\n✓ 配置加载成功！")
	fmt.Println("准备开始构建广告创意生成服务...")
}
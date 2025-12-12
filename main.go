package main

import (
	"ads-creative-gen-platform/config"
	"ads-creative-gen-platform/internal/handlers"
	"ads-creative-gen-platform/internal/middleware"
	"ads-creative-gen-platform/pkg/database"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	database.InitDatabase()
	defer database.CloseDB()

	// 设置 Gin 模式
	if config.AppConfig.AppMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()

	// 添加CORS中间件
	r.Use(middleware.CORSMiddleware())

	// 创建处理器
	creativeHandler := handlers.NewCreativeHandler()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "ads-creative-platform",
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// 创意生成接口
		v1.POST("/creative/generate", creativeHandler.Generate)

		// 查询任务接口
		v1.GET("/creative/task/:id", creativeHandler.GetTask)

		// 获取所有创意素材接口
		v1.GET("/creative/assets", creativeHandler.ListAllAssets)
	}

	// 启动服务
	port := config.AppConfig.HttpPort
	fmt.Printf("\n🚀 Server starting on %s\n", port)
	fmt.Printf("📖 API Docs: http://localhost%s/api/v1/ping\n", port)
	fmt.Printf("💚 Health Check: http://localhost%s/health\n\n", port)

	if err := r.Run(port); err != nil {
		panic(err)
	}
}

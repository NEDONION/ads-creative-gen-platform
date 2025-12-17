package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"ads-creative-gen-platform/config"
	creativehandler "ads-creative-gen-platform/internal/creative/handler"
	experimenthandler "ads-creative-gen-platform/internal/experiment/handler"
	"ads-creative-gen-platform/internal/middleware"
	"ads-creative-gen-platform/internal/tracing"
	"ads-creative-gen-platform/internal/warmup"
	"ads-creative-gen-platform/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库，并按开关决定是否迁移
	autoMigrate := strings.ToLower(strings.TrimSpace(os.Getenv("AUTO_MIGRATE"))) == "true"
	if autoMigrate {
		database.InitializeDatabase()
	} else {
		database.InitDatabase()
	}
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
	creativeHandler := creativehandler.NewCreativeHandler()
	experimentHandler := experimenthandler.NewExperimentHandler()
	traceHandler := tracing.NewTraceHandler()

	// 启动预热任务：保持 DB / 缓存温热
	var sqlDB *sql.DB
	if database.DB != nil {
		sqlDB, _ = database.DB.DB()
	}
	warmupManager := warmup.New(
		warmup.Config{
			Interval: 10 * time.Minute,
			Timeout:  2 * time.Second,
		},
		warmup.Targets{
			DB:         sqlDB,
			GormDB:     database.DB,
			Creative:   creativeHandler.Service(),
			Experiment: experimentHandler.Service(),
			Trace:      traceHandler.Service(),
		},
	)
	warmupManager.Start()

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

		// 文案生成/确认
		v1.POST("/copywriting/generate", creativeHandler.GenerateCopywriting)
		v1.POST("/copywriting/confirm", creativeHandler.ConfirmCopywriting)

		// 创意生成接口
		v1.POST("/creative/generate", creativeHandler.Generate)
		v1.POST("/creative/start", creativeHandler.StartCreative)

		// 查询任务接口
		v1.GET("/creative/task/:id", creativeHandler.GetTask)
		v1.DELETE("/creative/task/:id", creativeHandler.DeleteTask)

		// 获取所有创意素材接口
		v1.GET("/creative/assets", creativeHandler.ListAllAssets)

		// 获取所有任务接口
		v1.GET("/creative/tasks", creativeHandler.ListAllTasks)

		// 实验接口
		v1.POST("/experiments", experimentHandler.CreateExperiment)
		v1.GET("/experiments", experimentHandler.ListExperiments)
		v1.POST("/experiments/:id/status", experimentHandler.UpdateStatus)
		v1.GET("/experiments/:id/assign", experimentHandler.Assign)
		v1.POST("/experiments/:id/hit", experimentHandler.Hit)
		v1.POST("/experiments/:id/click", experimentHandler.Click)
		v1.GET("/experiments/:id/metrics", experimentHandler.Metrics)

		// Trace 调用链接口（目前为示例数据）
		v1.GET("/model_traces", traceHandler.ListTraces)
		v1.GET("/model_traces/:id", traceHandler.GetTrace)

		// 预热状态
		v1.GET("/warmup/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code": 0,
				"data": warmupManager.Stats(),
			})
		})
		// 手动触发预热
		v1.POST("/warmup/run", func(c *gin.Context) {
			warmupManager.RunNow()
			c.JSON(200, gin.H{
				"code": 0,
				"data": warmupManager.Stats(),
			})
		})
	}

	// 静态文件服务 - 托管前端
	r.GET("/assets/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		// 根路径或目录请求时返回前端入口，避免 404
		if path == "" || path == "/" {
			c.File("./web/dist/index.html")
			return
		}
		full := "./web/dist/assets" + path
		if _, err := os.Stat(full); err == nil {
			c.File(full)
			return
		}
		// 未找到文件时回退到 SPA
		c.File("./web/dist/index.html")
	})
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")
	r.StaticFile("/vite.svg", "./web/dist/vite.svg")
	r.StaticFile("/experiment-widget.js", "./web/dist/experiment-widget.js")

	// SPA fallback - 所有未匹配的路由返回 index.html（支持 React Router）
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	// 启动服务
	port := config.AppConfig.HttpPort
	fmt.Printf("\n🚀 Server starting on %s\n", port)
	fmt.Printf("📖 API Docs: http://localhost%s/api/v1/ping\n", port)
	fmt.Printf("💚 Health Check: http://localhost%s/health\n\n", port)

	if err := r.Run(port); err != nil {
		panic(err)
	}
}

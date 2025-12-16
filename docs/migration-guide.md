# Creative Service 重构迁移指南

## 概述

这份指南说明了如何从旧版 `CreativeService` 迁移到新版 `CreativeServiceV2`，新版本采用了 Go 最佳实践，包括：

1. **依赖注入** - 通过构造函数注入依赖，而非全局变量
2. **接口隔离** - 定义清晰的接口边界，便于测试和替换实现
3. **Repository 模式** - 数据访问层独立，解耦业务逻辑和数据库
4. **Worker Pool** - 管理后台任务，防止 goroutine 泄漏
5. **Context 传递** - 支持超时和取消，提高稳定性
6. **配置抽离** - 魔法数字转为常量，便于维护

## 主要改进

### 1. 依赖注入 + 接口隔离

**旧版 (❌):**
```go
type CreativeService struct {
    tongyiClient *TongyiClient    // 硬编码依赖
    qiniuService *QiniuService    // 硬编码依赖
}

func NewCreativeService() *CreativeService {
    return &CreativeService{
        tongyiClient: NewTongyiClient(),    // 构造函数内创建
        qiniuService: NewQiniuService(),     // 构造函数内创建
    }
}

// 业务代码直接使用全局 DB
database.DB.Create(&task)
```

**新版 (✅):**
```go
// 定义接口
type TongyiClient interface {
    GenerateImage(ctx context.Context, ...) (*ImageGenResponse, string, error)
    QueryTask(ctx context.Context, ...) (*QueryResp, error)
}

type TaskRepository interface {
    Create(ctx context.Context, task *models.CreativeTask) error
    GetByID(ctx context.Context, id uint) (*models.CreativeTask, error)
}

// 依赖注入
type CreativeServiceV2 struct {
    tongyiClient TongyiClient      // 接口类型
    qiniuClient  QiniuClient       // 接口类型
    taskRepo     TaskRepository    // 接口类型
    assetRepo    AssetRepository   // 接口类型
    workerPool   *WorkerPool
}

func NewCreativeServiceV2(
    tongyiClient TongyiClient,
    qiniuClient QiniuClient,
    taskRepo TaskRepository,
    assetRepo AssetRepository,
    workerPool *WorkerPool,
) *CreativeServiceV2 {
    return &CreativeServiceV2{
        tongyiClient: tongyiClient,
        qiniuClient:  qiniuClient,
        taskRepo:     taskRepo,
        assetRepo:    assetRepo,
        workerPool:   workerPool,
    }
}
```

**优势:**
- ✅ 可测试 - 可以注入 Mock 实现
- ✅ 可替换 - 轻松切换不同的实现
- ✅ 解耦 - 服务不依赖具体实现

### 2. Worker Pool 管理后台任务

**旧版 (❌):**
```go
func (s *CreativeService) CreateTask(...) (*models.CreativeTask, error) {
    // ... 创建任务

    // 无限制地启动 goroutine
    go s.processTask(task.ID)  // ❌ 没有并发控制
                                // ❌ 没有 panic recover
                                // ❌ 没有超时控制
                                // ❌ 进程重启任务丢失

    return &task, nil
}
```

**新版 (✅):**
```go
func (s *CreativeServiceV2) CreateTask(ctx context.Context, ...) (*models.CreativeTask, error) {
    // ... 创建任务

    // 提交到 Worker Pool
    creativeTask := NewCreativeProcessTask(task.ID, s)
    if err := s.workerPool.Submit(creativeTask); err != nil {
        // ✅ 提交失败有错误处理
        _ = s.taskRepo.UpdateStatus(ctx, task.ID, models.TaskFailed, ProgressCompleted)
        return nil, fmt.Errorf("failed to submit task: %w", err)
    }

    return task, nil
}

// Worker Pool 特性：
// ✅ 限制并发数（默认 10 个 worker）
// ✅ 任务队列（默认 100 容量）
// ✅ Panic recover（worker 不会崩溃）
// ✅ 优雅关闭（等待任务完成）
// ✅ Context 超时控制（每个任务 3 分钟）
```

### 3. Context 传递和超时控制

**旧版 (❌):**
```go
func (s *CreativeService) processTask(taskID uint) {
    // ❌ 使用 context.Background()
    resp, _, err := s.tongyiClient.GenerateImage(context.Background(), ...)

    // ❌ 无限轮询
    for i := 0; i < 60; i++ {
        time.Sleep(2 * time.Second)
        queryResp, err := s.tongyiClient.QueryTask(context.Background(), ...)
        // ...
    }
}
```

**新版 (✅):**
```go
func (s *CreativeServiceV2) processTask(ctx context.Context, taskID uint) error {
    // ✅ 接收 context 参数
    resp, _, err := s.tongyiClient.GenerateImage(ctx, ...)

    // ✅ 使用 ticker + select 模式，支持 context 取消
    ticker := time.NewTicker(PollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // ✅ Context 取消时立即返回
            return s.handleTaskFailure(ctx, task.ID, traceID, ctx.Err())

        case <-ticker.C:
            // ✅ 有超时限制
            attempts++
            if attempts > MaxPollAttempts {
                return s.handleTaskFailure(ctx, task.ID, traceID, errors.New("timeout"))
            }
            // ...
        }
    }
}
```

### 4. Repository 模式

**旧版 (❌):**
```go
// 业务代码直接操作数据库
func (s *CreativeService) CreateTask(...) {
    database.DB.Create(&task)  // ❌ 全局变量
    database.DB.Model(&task).Update("status", ...)
    database.DB.Where("uuid = ?", uuid).First(&task)
}
```

**新版 (✅):**
```go
// Repository 接口
type TaskRepository interface {
    Create(ctx context.Context, task *models.CreativeTask) error
    UpdateStatus(ctx context.Context, id uint, status models.TaskStatus, progress int) error
    GetByUUID(ctx context.Context, uuid string) (*models.CreativeTask, error)
}

// 业务代码通过接口操作
func (s *CreativeServiceV2) CreateTask(ctx context.Context, ...) {
    err := s.taskRepo.Create(ctx, task)  // ✅ 通过接口
    err = s.taskRepo.UpdateStatus(ctx, task.ID, models.TaskCompleted, 100)
    task, err := s.taskRepo.GetByUUID(ctx, uuid)
}
```

### 5. 配置抽离

**旧版 (❌):**
```go
// 魔法数字散落各处
resp, _, err := s.tongyiClient.GenerateImage(ctx, prompt, "1024*1024", ...)  // ❌
asset.Width = 1024   // ❌
asset.Height = 1024  // ❌
for i := 0; i < 60; i++ {  // ❌
    time.Sleep(2 * time.Second)  // ❌
}
```

**新版 (✅):**
```go
// 配置常量集中定义 (config.go)
const (
    DefaultImageSize   = "1024*1024"
    DefaultImageWidth  = 1024
    DefaultImageHeight = 1024
    MaxPollAttempts    = 60
    PollInterval       = 2 * time.Second
    TaskTimeout        = 3 * time.Minute
)

// 使用常量
resp, _, err := s.tongyiClient.GenerateImage(ctx, prompt, DefaultImageSize, ...)  // ✅
asset.Width = DefaultImageWidth   // ✅
asset.Height = DefaultImageHeight // ✅
```

## 迁移步骤

### 步骤 1: 初始化新服务

在 `main.go` 或应用入口处：

```go
package main

import (
    "ads-creative-gen-platform/internal/services"
    "ads-creative-gen-platform/pkg/database"
)

func main() {
    // 初始化数据库
    database.InitializeDatabase()

    // 初始化服务 (使用辅助函数)
    creativeService, _, cleanup := services.InitServices()
    defer cleanup() // 确保优雅关闭 Worker Pool

    // 使用新服务
    router := setupRouter(creativeService)
    router.Run(":4000")
}
```

### 步骤 2: 更新 Handler

**旧版 Handler (❌):**
```go
type CreativeHandler struct {
    service *services.CreativeService
}

func NewCreativeHandler() *CreativeHandler {
    return &CreativeHandler{
        service: services.NewCreativeService(),  // ❌ 内部创建
    }
}

func (h *CreativeHandler) Create(c *gin.Context) {
    task, err := h.service.CreateTask(input)  // ❌ 无 context
    // ...
}
```

**新版 Handler (✅):**
```go
type CreativeHandler struct {
    service *services.CreativeServiceV2  // ✅ 接口类型更好
}

func NewCreativeHandler(service *services.CreativeServiceV2) *CreativeHandler {
    return &CreativeHandler{
        service: service,  // ✅ 依赖注入
    }
}

func (h *CreativeHandler) Create(c *gin.Context) {
    ctx := c.Request.Context()  // ✅ 从 gin.Context 获取
    task, err := h.service.CreateTask(ctx, input)  // ✅ 传递 context
    // ...
}
```

### 步骤 3: 测试

新版本的设计使得测试变得简单：

```go
func TestCreativeService(t *testing.T) {
    // Mock 依赖
    mockTongyi := &MockTongyiClient{}
    mockQiniu := &MockQiniuClient{}
    mockTaskRepo := &MockTaskRepository{}
    mockAssetRepo := &MockAssetRepository{}
    workerPool := services.NewWorkerPool(1, 10)
    workerPool.Start()
    defer workerPool.Stop()

    // 创建服务
    service := services.NewCreativeServiceV2(
        mockTongyi,
        mockQiniu,
        mockTaskRepo,
        mockAssetRepo,
        workerPool,
    )

    // 测试
    ctx := context.Background()
    task, err := service.CreateTask(ctx, services.CreateTaskInput{
        Title: "Test Product",
    })

    assert.NoError(t, err)
    assert.NotNil(t, task)
}
```

## 性能对比

| 指标 | 旧版 | 新版 | 改进 |
|------|------|------|------|
| 并发控制 | ❌ 无限制 | ✅ 10 workers + 100 queue | 防止资源耗尽 |
| 任务超时 | ❌ 可能永久卡住 | ✅ 3 分钟超时 | 提高稳定性 |
| Panic 恢复 | ❌ 进程崩溃 | ✅ Worker 自动恢复 | 高可用 |
| DB 轮询 | ❌ 60 次无节流 | ✅ 每 5 次更新一次 | 减少 DB 压力 |
| Context 取消 | ❌ 不支持 | ✅ 立即停止 | 资源节约 |

## 常见问题

### Q1: 是否需要立即迁移所有代码？

**A:** 不需要。两个版本可以共存：

```go
// 旧代码继续使用 CreativeService
oldService := services.NewCreativeService()

// 新代码使用 CreativeServiceV2
newService, _, cleanup := services.InitServices()
defer cleanup()
```

### Q2: 如何处理现有的未完成任务？

**A:** 进程重启后，可以查询数据库中状态为 `processing` 或 `queued` 的任务，重新提交到 Worker Pool：

```go
func recoverPendingTasks(ctx context.Context, service *services.CreativeServiceV2, taskRepo services.TaskRepository) {
    // 查询未完成任务
    tasks, _, _ := taskRepo.List(ctx, services.ListTasksQuery{
        Status: string(models.TaskProcessing),
        PageSize: 100,
    })

    // 重新提交
    for _, task := range tasks {
        _ = service.StartCreativeGeneration(ctx, task.UUID, nil)
    }
}
```

### Q3: Worker Pool 的最佳配置是什么？

**A:** 取决于你的服务器资源和外部 API 限制：

```go
// 高性能服务器 + 高 API 配额
workerPool := services.NewWorkerPool(20, 200)

// 低配服务器或 API 限流严格
workerPool := services.NewWorkerPool(5, 50)

// 默认配置（推荐）
workerPool := services.NewWorkerPool(
    services.DefaultWorkerCount,  // 10
    services.DefaultQueueSize,     // 100
)
```

## 清理旧代码

迁移完成并稳定运行后，可以删除旧文件：

```bash
# 删除旧服务
rm internal/services/creative_service.go

# 重命名新服务
mv internal/services/creative_service_v2.go internal/services/creative_service.go

# 更新所有引用
sed -i 's/CreativeServiceV2/CreativeService/g' **/*.go
```

## 总结

新版本的 `CreativeServiceV2` 遵循了 Go 社区的最佳实践，解决了旧版本的以下问题：

1. ✅ **可测试性** - 依赖注入 + 接口隔离
2. ✅ **稳定性** - Worker Pool + Context 超时
3. ✅ **可维护性** - Repository 模式 + 配置抽离
4. ✅ **可观测性** - 结构化日志 + 错误处理
5. ✅ **可扩展性** - 接口设计便于替换实现

建议采用渐进式迁移：先在新功能中使用新版本，验证稳定后再逐步迁移旧代码。

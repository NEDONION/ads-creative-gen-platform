# AI 多尺寸广告创意生成平台 - 可执行实施计划

> 从 MVP 到生产级系统的分阶段实施路线图

---

## 📋 项目概述

### 核心价值主张
输入商品信息 → 自动生成多尺寸、多风格广告图 → CTR 排序 → 最优创意推荐

### 对标产品
- Meta Advantage+ Creative
- TikTok Smart Creative
- Google Performance Max Creative Studio

---

## 🎯 总体架构演进路线

```
Phase 1: MVP (核心流程打通)
   ↓
Phase 2: 多尺寸支持 (核心竞争力)
   ↓
Phase 3: 智能排序 (质量提升)
   ↓
Phase 4: 生产化 (可观测性 + 性能)
   ↓
Phase 5: 高级特性 (A/B 测试 + 自动化)
```

---

# Phase 1: MVP - 核心流程验证 🚀

**目标**: 验证端到端生成流程，产出第一张可用的广告图

## 1.1 功能清单

- [x] 项目初始化 & 配置管理
- [ ] RESTful API 框架（Gin）
- [ ] 通义万相图像生成接入
- [ ] 基础图像处理（添加文本、Logo）
- [ ] 本地文件存储
- [ ] 简单的任务状态管理

## 1.2 技术架构

```
┌─────────────┐
│   用户请求   │
└──────┬──────┘
       │
┌──────▼──────┐
│  Gin API    │ (main.go + handlers/)
└──────┬──────┘
       │
┌──────▼──────────┐
│ Creative Service│ (services/creative.go)
└──────┬──────────┘
       │
       ├─────────┬──────────┐
       │         │          │
┌──────▼───┐ ┌──▼────┐  ┌──▼────────┐
│通义万相  │ │Image  │  │Storage    │
│ Client   │ │Process│  │(local FS) │
└──────────┘ └───────┘  └───────────┘
```

## 1.3 数据模型

### 数据库表设计（SQLite/PostgreSQL）

```sql
-- 创意生成任务表
CREATE TABLE creative_tasks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(64),
    title VARCHAR(255) NOT NULL,
    selling_points JSON,
    product_image_url VARCHAR(512),
    status VARCHAR(20), -- pending, processing, completed, failed
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- 生成的创意素材表
CREATE TABLE creative_assets (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36),
    format VARCHAR(20), -- 1:1, 4:5, 9:16, etc.
    image_url VARCHAR(512),
    prompt TEXT,
    metadata JSON,
    created_at TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES creative_tasks(id)
);
```

### 配置文件扩展（.env）

```env
# 通义万相 API
TONGYI_API_KEY=sk-xxx
TONGYI_IMAGE_MODEL=wanx-v1  # 或 flux-schnell

# 服务配置
SERVER_PORT=8080
ENVIRONMENT=development

# 存储配置
STORAGE_TYPE=local  # local, oss, s3
STORAGE_PATH=./uploads
```

## 1.4 API 设计

### 1. 创建生成任务

**POST** `/api/v1/creative/generate`

**请求体**:
```json
{
  "title": "户外露营帐篷",
  "selling_points": ["防水", "三季通用", "轻量化"],
  "product_image_url": "https://example.com/tent.jpg",
  "style": "modern"  // modern, elegant, vibrant
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_abc123",
    "status": "processing",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. 查询任务状态

**GET** `/api/v1/creative/task/:task_id`

**响应**:
```json
{
  "code": 0,
  "data": {
    "task_id": "task_abc123",
    "status": "completed",
    "creative": {
      "image_url": "http://localhost:8080/uploads/abc123.png",
      "format": "1:1",
      "size": "1024x1024"
    }
  }
}
```

## 1.5 核心代码结构

```
ads-creative-gen-platform/
├── main.go                    # 入口
├── config/
│   └── config.go              # 配置管理
├── internal/
│   ├── handlers/              # HTTP 处理器
│   │   └── creative.go
│   ├── services/              # 业务逻辑
│   │   ├── creative.go        # 创意生成服务
│   │   ├── tongyi.go          # 通义 API 客户端
│   │   └── image_processor.go # 图像处理
│   ├── models/                # 数据模型
│   │   └── creative.go
│   └── storage/               # 存储层
│       └── local.go
├── uploads/                   # 本地存储目录
└── go.mod
```

## 1.6 通义万相 API 集成

### SDK 示例

```go
package services

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type TongyiClient struct {
    apiKey string
    baseURL string
}

type ImageGenRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    N      int    `json:"n"`
    Size   string `json:"size"`
}

type ImageGenResponse struct {
    Output struct {
        Results []struct {
            URL string `json:"url"`
        } `json:"results"`
    } `json:"output"`
}

func (c *TongyiClient) GenerateImage(prompt string) (string, error) {
    req := ImageGenRequest{
        Model:  "wanx-v1",
        Prompt: prompt,
        N:      1,
        Size:   "1024*1024",
    }

    // 实现 HTTP 请求...
    // 返回图片 URL
}
```

## 1.7 图像处理 - 添加文本

使用 `github.com/fogleman/gg` 进行图像处理：

```go
package services

import (
    "github.com/fogleman/gg"
    "image"
)

type ImageProcessor struct{}

func (p *ImageProcessor) AddText(img image.Image, text string) (image.Image, error) {
    dc := gg.NewContextForImage(img)

    // 加载字体
    if err := dc.LoadFontFace("/path/to/font.ttf", 48); err != nil {
        return nil, err
    }

    // 绘制文本
    dc.SetRGB(1, 1, 1) // 白色
    dc.DrawStringAnchored(text, 512, 900, 0.5, 0.5)

    return dc.Image(), nil
}
```

## 1.8 验收标准

- [ ] API 可以接收请求并返回 task_id
- [ ] 成功调用通义万相 API 生成图片
- [ ] 在生成的图片上添加商品标题文本
- [ ] 图片保存到本地并可通过 URL 访问
- [ ] 任务状态可以正确查询（pending → processing → completed）
- [ ] 基本的错误处理（API 失败、超时等）

## 1.9 依赖安装

```bash
go get -u github.com/gin-gonic/gin
go get -u github.com/joho/godotenv
go get -u github.com/google/uuid
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
go get -u github.com/fogleman/gg
```

## 1.10 测试用例

```bash
# 测试图像生成
curl -X POST http://localhost:8080/api/v1/creative/generate \
  -H "Content-Type: application/json" \
  -d '{
    "title": "户外露营帐篷",
    "selling_points": ["防水", "三季通用"],
    "style": "modern"
  }'

# 查询任务状态
curl http://localhost:8080/api/v1/creative/task/task_abc123
```

---

# Phase 2: 多尺寸支持 - 核心竞争力 📐

**目标**: 支持多种广告尺寸自动生成，实现智能布局

## 2.1 功能清单

- [ ] 支持多种尺寸规格（1:1, 4:5, 9:16, 1200x628, 728x90）
- [ ] 智能裁剪 & 自适应布局
- [ ] 主体检测（可选：SAM 模型）
- [ ] 文本区域自动避让
- [ ] CTA 按钮生成
- [ ] Logo 自动放置

## 2.2 支持的广告尺寸

| 平台 | 尺寸 | 用途 |
|------|------|------|
| Instagram Feed | 1:1 (1080x1080) | 动态广告 |
| Instagram Story | 9:16 (1080x1920) | 故事广告 |
| Facebook Feed | 4:5 (1080x1350) | 信息流 |
| Google Display | 1200x628 | 展示广告 |
| Banner | 728x90 | 横幅广告 |
| TikTok | 9:16 (1080x1920) | 视频封面 |

## 2.3 布局引擎设计

```go
package services

type LayoutEngine struct {
    config LayoutConfig
}

type LayoutConfig struct {
    Format         string  // "1:1", "4:5", "9:16"
    Width          int
    Height         int
    SafeArea       Margin  // 安全区域
    TextPosition   string  // "top", "bottom", "center"
    CTAPosition    string  // "bottom-right", "bottom-center"
}

type Margin struct {
    Top, Bottom, Left, Right int
}

func (e *LayoutEngine) GenerateLayout(
    baseImage image.Image,
    title string,
    cta string,
    logoPath string,
) (image.Image, error) {
    // 1. 调整画布尺寸
    canvas := e.resizeCanvas(baseImage)

    // 2. 检测主体位置（可选）
    subjectBounds := e.detectSubject(baseImage)

    // 3. 计算文本安全区域
    textArea := e.calculateTextArea(subjectBounds)

    // 4. 添加文本
    canvas = e.addText(canvas, title, textArea)

    // 5. 添加 CTA 按钮
    canvas = e.addCTAButton(canvas, cta)

    // 6. 添加 Logo
    canvas = e.addLogo(canvas, logoPath)

    return canvas, nil
}
```

## 2.4 API 更新

### 请求支持多尺寸

**POST** `/api/v1/creative/generate`

```json
{
  "title": "户外露营帐篷",
  "selling_points": ["防水", "三季通用"],
  "product_image_url": "https://example.com/tent.jpg",
  "formats": ["1:1", "4:5", "9:16"],  // 新增
  "cta_text": "立即购买",              // 新增
  "logo_url": "https://example.com/logo.png"  // 新增
}
```

### 响应返回多尺寸

```json
{
  "code": 0,
  "data": {
    "task_id": "task_abc123",
    "status": "completed",
    "creatives": [
      {
        "format": "1:1",
        "size": "1080x1080",
        "image_url": "http://localhost:8080/uploads/abc123_1x1.png"
      },
      {
        "format": "4:5",
        "size": "1080x1350",
        "image_url": "http://localhost:8080/uploads/abc123_4x5.png"
      },
      {
        "format": "9:16",
        "size": "1080x1920",
        "image_url": "http://localhost:8080/uploads/abc123_9x16.png"
      }
    ]
  }
}
```

## 2.5 智能裁剪策略

```go
func (e *LayoutEngine) SmartCrop(img image.Image, targetRatio float64) image.Image {
    srcBounds := img.Bounds()
    srcRatio := float64(srcBounds.Dx()) / float64(srcBounds.Dy())

    if srcRatio > targetRatio {
        // 宽图 → 窄尺寸，裁剪左右
        newWidth := int(float64(srcBounds.Dy()) * targetRatio)
        x0 := (srcBounds.Dx() - newWidth) / 2
        return crop(img, x0, 0, newWidth, srcBounds.Dy())
    } else {
        // 窄图 → 宽尺寸，裁剪上下
        newHeight := int(float64(srcBounds.Dx()) / targetRatio)
        y0 := (srcBounds.Dy() - newHeight) / 2
        return crop(img, 0, y0, srcBounds.Dx(), newHeight)
    }
}
```

## 2.6 CTA 按钮生成

```go
func (e *LayoutEngine) DrawCTAButton(dc *gg.Context, text string, x, y, width, height float64) {
    // 绘制圆角矩形按钮
    dc.DrawRoundedRectangle(x, y, width, height, 10)
    dc.SetRGB(0.2, 0.6, 1.0) // 蓝色
    dc.Fill()

    // 绘制按钮文字
    dc.SetRGB(1, 1, 1) // 白色文字
    dc.LoadFontFace("/path/to/font.ttf", 32)
    dc.DrawStringAnchored(text, x+width/2, y+height/2, 0.5, 0.5)
}
```

## 2.7 验收标准

- [ ] 一次请求可生成 3+ 种尺寸的广告图
- [ ] 不同尺寸的文本自动适配位置
- [ ] CTA 按钮正确渲染
- [ ] Logo 不遮挡主体内容
- [ ] 智能裁剪保留图像主体

---

# Phase 3: 智能排序 - 质量提升 🎯

**目标**: 引入质量评分和 CTR 预测，自动筛选最优创意

## 3.1 功能清单

- [ ] 通义千问生成多组创意文案
- [ ] 基于规则的质量评分（亮度、对比度、清晰度）
- [ ] CLIP 图文匹配评分（可选）
- [ ] 简单 CTR 预测模型（基于历史数据）
- [ ] 返回 Top-K 最优创意

## 3.2 创意变体生成

### 通义千问生成文案

```go
package services

type QwenClient struct {
    apiKey string
}

type CreativeBrief struct {
    Theme           string   // "节日版", "极简版", "活力版"
    Headline        string   // 主标题
    Subheadline     string   // 副标题
    CTA             string   // 行动号召
    BackgroundStyle string   // 背景风格描述
    ColorScheme     []string // 配色建议
}

func (c *QwenClient) GenerateBriefs(
    title string,
    sellingPoints []string,
    numVariants int,
) ([]CreativeBrief, error) {
    prompt := fmt.Sprintf(`
你是一个广告创意专家。请为以下产品生成 %d 组不同风格的广告创意方案：

产品标题：%s
卖点：%s

每组方案包括：
1. 创意主题（如：节日版、极简版、活力版）
2. 主标题（不超过15字）
3. 副标题（不超过20字）
4. CTA 文案（不超过5字）
5. 背景风格描述
6. 配色建议（3-5个颜色）

以 JSON 格式输出。
`, numVariants, title, strings.Join(sellingPoints, "、"))

    // 调用通义千问 API
    // 解析返回的 JSON
    return briefs, nil
}
```

## 3.3 质量评分系统

### 基于规则的评分

```go
package services

type QualityScorer struct{}

type QualityScore struct {
    Brightness  float64 // 0-1，亮度评分
    Contrast    float64 // 0-1，对比度评分
    Sharpness   float64 // 0-1，清晰度评分
    Composition float64 // 0-1，构图评分
    Overall     float64 // 综合评分
}

func (s *QualityScorer) Score(img image.Image) QualityScore {
    score := QualityScore{}

    // 1. 检测亮度
    score.Brightness = s.calculateBrightness(img)

    // 2. 检测对比度
    score.Contrast = s.calculateContrast(img)

    // 3. 检测清晰度（拉普拉斯方差）
    score.Sharpness = s.calculateSharpness(img)

    // 4. 构图评分（三分法）
    score.Composition = s.evaluateComposition(img)

    // 综合评分
    score.Overall = (score.Brightness*0.2 +
                     score.Contrast*0.3 +
                     score.Sharpness*0.3 +
                     score.Composition*0.2)

    return score
}

func (s *QualityScorer) calculateBrightness(img image.Image) float64 {
    // 计算平均亮度，理想范围 0.4-0.7
    // 实现略...
    return 0.6
}
```

### CTR 预测（简化版）

```go
type CTRPredictor struct {
    db *gorm.DB
}

type CTRFeatures struct {
    Format          string
    HasCTA          bool
    TextLength      int
    BrightnessScore float64
    ContrastScore   float64
}

func (p *CTRPredictor) Predict(features CTRFeatures) float64 {
    // Phase 3: 使用简单的加权规则
    // Phase 5: 可升级为真实的机器学习模型

    score := 0.5 // 基础分

    // 格式权重
    if features.Format == "9:16" {
        score += 0.1 // Story 格式 CTR 更高
    }

    // CTA 加分
    if features.HasCTA {
        score += 0.15
    }

    // 质量评分加权
    score += features.BrightnessScore * 0.1
    score += features.ContrastScore * 0.15

    return math.Min(score, 1.0)
}
```

## 3.4 数据库扩展

```sql
-- 创意评分表
CREATE TABLE creative_scores (
    creative_id VARCHAR(36) PRIMARY KEY,
    quality_score DECIMAL(3,2),
    ctr_prediction DECIMAL(3,2),
    brightness DECIMAL(3,2),
    contrast DECIMAL(3,2),
    sharpness DECIMAL(3,2),
    created_at TIMESTAMP,
    FOREIGN KEY (creative_id) REFERENCES creative_assets(id)
);

-- 实际 CTR 数据（用于后续模型训练）
CREATE TABLE creative_performance (
    id VARCHAR(36) PRIMARY KEY,
    creative_id VARCHAR(36),
    impressions INT,
    clicks INT,
    ctr DECIMAL(5,4),
    date DATE,
    FOREIGN KEY (creative_id) REFERENCES creative_assets(id)
);
```

## 3.5 API 响应更新

```json
{
  "code": 0,
  "data": {
    "task_id": "task_abc123",
    "status": "completed",
    "creatives": [
      {
        "format": "1:1",
        "image_url": "...",
        "scores": {
          "quality": 0.87,
          "ctr_prediction": 0.74,
          "brightness": 0.65,
          "contrast": 0.82
        },
        "rank": 1  // 根据 CTR 预测排序
      },
      {
        "format": "4:5",
        "image_url": "...",
        "scores": {
          "quality": 0.79,
          "ctr_prediction": 0.68
        },
        "rank": 2
      }
    ]
  }
}
```

## 3.6 验收标准

- [ ] 每个请求生成 3-5 组不同风格的创意
- [ ] 质量评分系统可以识别模糊/过暗的图片
- [ ] CTR 预测分数合理（0-1 之间）
- [ ] 创意按 CTR 预测分数降序排列
- [ ] 返回 Top-3 最优创意

---

# Phase 4: 生产化 - 可观测性与性能 🚀

**目标**: 系统生产就绪，支持高并发，完善监控

## 4.1 功能清单

- [ ] 任务队列（Redis + Goroutine Pool）
- [ ] 对象存储（MinIO/阿里云 OSS）
- [ ] 日志系统（Zap）
- [ ] 指标监控（Prometheus + Grafana）
- [ ] 限流与熔断
- [ ] Docker 容器化
- [ ] API 文档（Swagger）

## 4.2 异步任务队列

```go
package queue

import (
    "github.com/go-redis/redis/v8"
)

type TaskQueue struct {
    redis *redis.Client
    workers int
}

func (q *TaskQueue) Enqueue(taskID string, payload map[string]interface{}) error {
    data, _ := json.Marshal(payload)
    return q.redis.LPush(ctx, "creative:tasks", data).Err()
}

func (q *TaskQueue) StartWorkers() {
    for i := 0; i < q.workers; i++ {
        go q.worker(i)
    }
}

func (q *TaskQueue) worker(id int) {
    for {
        result := q.redis.BRPop(ctx, 0, "creative:tasks").Val()
        if len(result) > 1 {
            var task Task
            json.Unmarshal([]byte(result[1]), &task)

            // 处理任务
            q.processTask(task)
        }
    }
}
```

## 4.3 对象存储

```go
package storage

import (
    "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSSStorage struct {
    client *oss.Client
    bucket string
}

func (s *OSSStorage) Upload(key string, data []byte) (string, error) {
    bucket, _ := s.client.Bucket(s.bucket)
    err := bucket.PutObject(key, bytes.NewReader(data))
    if err != nil {
        return "", err
    }

    // 返回公开访问 URL
    return fmt.Sprintf("https://%s.oss-cn-hangzhou.aliyuncs.com/%s",
        s.bucket, key), nil
}
```

## 4.4 监控指标

### Prometheus 指标定义

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    // 任务处理时长
    taskDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "creative_task_duration_seconds",
            Help: "Duration of creative generation tasks",
        },
        []string{"status"},
    )

    // 生成的创意数量
    creativesGenerated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "creatives_generated_total",
            Help: "Total number of creatives generated",
        },
        []string{"format"},
    )

    // API 调用次数
    apiCalls = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_calls_total",
            Help: "Total API calls",
        },
        []string{"endpoint", "status"},
    )
)
```

### Grafana Dashboard 配置

监控面板应包括：
- 每秒请求数（QPS）
- 平均响应时间
- 成功率 / 失败率
- 各尺寸生成量分布
- 通义 API 调用延迟
- 存储使用量

## 4.5 Docker 化

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

EXPOSE 8080
CMD ["./main"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - TONGYI_API_KEY=${TONGYI_API_KEY}
      - REDIS_ADDR=redis:6379
      - DB_HOST=postgres
    depends_on:
      - redis
      - postgres

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=creative_platform
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=password
    volumes:
      - pg_data:/var/lib/postgresql/data

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

volumes:
  pg_data:
```

## 4.6 限流中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

func RateLimiter(r rate.Limit, b int) gin.HandlerFunc {
    limiter := rate.NewLimiter(r, b)

    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

// 使用
// router.Use(RateLimiter(10, 20)) // 每秒10个请求，桶容量20
```

## 4.7 验收标准

- [ ] 单机 QPS 达到 100+（使用任务队列）
- [ ] Prometheus 指标正常采集
- [ ] Grafana Dashboard 可视化正常
- [ ] Docker 镜像构建成功
- [ ] API 响应时间 P95 < 500ms（不含模型推理）
- [ ] 日志结构化，可按 task_id 追踪

---

# Phase 5: 高级特性 - A/B 测试与自动化 🧪

**目标**: 完整的创意优化闭环，自动化投放与实验

## 5.1 功能清单

- [ ] A/B 测试管理
- [ ] 实际 CTR 数据回传
- [ ] CTR 预测模型训练（ML Pipeline）
- [ ] 自动化创意优化建议
- [ ] Webhook 通知
- [ ] 批量生成 API

## 5.2 A/B 测试系统

### 实验配置表

```sql
CREATE TABLE ab_experiments (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255),
    start_date DATE,
    end_date DATE,
    status VARCHAR(20), -- running, paused, completed
    config JSON, -- 实验参数
    created_at TIMESTAMP
);

CREATE TABLE ab_variants (
    id VARCHAR(36) PRIMARY KEY,
    experiment_id VARCHAR(36),
    creative_id VARCHAR(36),
    traffic_allocation DECIMAL(3,2), -- 流量分配 0-1
    FOREIGN KEY (experiment_id) REFERENCES ab_experiments(id),
    FOREIGN KEY (creative_id) REFERENCES creative_assets(id)
);
```

### API: 创建 A/B 实验

**POST** `/api/v1/experiment/create`

```json
{
  "name": "618 促销广告测试",
  "variants": [
    {"creative_id": "creative_001", "traffic": 0.5},
    {"creative_id": "creative_002", "traffic": 0.5}
  ],
  "start_date": "2024-06-01",
  "end_date": "2024-06-18"
}
```

## 5.3 CTR 数据回传

### Webhook API

**POST** `/api/v1/performance/report`

```json
{
  "creative_id": "creative_001",
  "date": "2024-06-01",
  "impressions": 10000,
  "clicks": 320,
  "conversions": 15
}
```

## 5.4 机器学习 Pipeline

### 模型训练脚本（Python）

```python
# scripts/train_ctr_model.py

import pandas as pd
from sklearn.ensemble import RandomForestRegressor
import joblib

# 1. 从数据库加载历史数据
df = pd.read_sql("""
    SELECT
        ca.format,
        cs.quality_score,
        cs.brightness,
        cs.contrast,
        cp.ctr
    FROM creative_assets ca
    JOIN creative_scores cs ON ca.id = cs.creative_id
    JOIN creative_performance cp ON ca.id = cp.creative_id
""", connection)

# 2. 特征工程
X = df[['quality_score', 'brightness', 'contrast']]
y = df['ctr']

# 3. 训练模型
model = RandomForestRegressor(n_estimators=100)
model.fit(X, y)

# 4. 保存模型
joblib.dump(model, 'models/ctr_predictor.pkl')
```

### Go 调用 Python 模型

```go
package services

import (
    "os/exec"
    "encoding/json"
)

type MLPredictor struct {
    scriptPath string
}

func (p *MLPredictor) PredictCTR(features CTRFeatures) (float64, error) {
    input, _ := json.Marshal(features)

    cmd := exec.Command("python3", p.scriptPath, string(input))
    output, err := cmd.Output()
    if err != nil {
        return 0, err
    }

    var result struct {
        CTR float64 `json:"ctr"`
    }
    json.Unmarshal(output, &result)

    return result.CTR, nil
}
```

## 5.5 自动化优化建议

```go
package services

type CreativeOptimizer struct {
    db *gorm.DB
}

type OptimizationSuggestion struct {
    CreativeID   string
    CurrentCTR   float64
    Suggestions  []string
    ExpectedLift float64
}

func (o *CreativeOptimizer) Analyze(creativeID string) OptimizationSuggestion {
    // 获取创意数据
    creative := o.getCreative(creativeID)
    performance := o.getPerformance(creativeID)

    suggestions := []string{}

    // 规则引擎
    if creative.BrightnessScore < 0.4 {
        suggestions = append(suggestions, "建议提高图片亮度")
    }

    if !creative.HasCTA {
        suggestions = append(suggestions, "建议添加明确的 CTA 按钮")
    }

    if performance.CTR < 0.02 {
        suggestions = append(suggestions, "CTR 低于平均水平，建议更换创意风格")
    }

    return OptimizationSuggestion{
        CreativeID:   creativeID,
        CurrentCTR:   performance.CTR,
        Suggestions:  suggestions,
        ExpectedLift: 0.15, // 预期提升
    }
}
```

## 5.6 批量生成 API

**POST** `/api/v1/creative/batch`

```json
{
  "products": [
    {
      "title": "产品 A",
      "selling_points": ["卖点1", "卖点2"],
      "image_url": "..."
    },
    {
      "title": "产品 B",
      "selling_points": ["卖点3", "卖点4"],
      "image_url": "..."
    }
  ],
  "formats": ["1:1", "9:16"],
  "num_variants_per_product": 3
}
```

响应：
```json
{
  "batch_id": "batch_xyz",
  "total_tasks": 6,
  "estimated_time": "120s"
}
```

## 5.7 验收标准

- [ ] A/B 实验可以正确分配流量
- [ ] CTR 数据回传并存储
- [ ] 机器学习模型定期重训练（每周/每月）
- [ ] 优化建议准确率 > 60%
- [ ] 批量 API 支持 100+ 产品同时生成
- [ ] Webhook 通知任务完成事件

---

# 📊 整体项目里程碑

| 阶段 | 预估工作量 | 核心交付物 | 关键指标 |
|------|-----------|-----------|----------|
| **Phase 1** | 1-2 周 | 可工作的 MVP | 生成第一张广告图 |
| **Phase 2** | 2-3 周 | 多尺寸生成引擎 | 支持 5+ 种尺寸 |
| **Phase 3** | 2 周 | 智能排序系统 | CTR 预测准确率 > 50% |
| **Phase 4** | 2-3 周 | 生产级系统 | QPS > 100, P95 < 500ms |
| **Phase 5** | 3-4 周 | 完整闭环 | A/B 测试自动化 |

---

# 🛠️ 技术栈总结

## 后端

- **Go 1.21+**: 核心服务
- **Gin**: Web 框架
- **GORM**: ORM
- **Redis**: 任务队列 & 缓存
- **PostgreSQL**: 关系数据库

## AI & 图像

- **通义万相**: 图像生成
- **通义千问**: 文案生成
- **gg**: Go 图像处理库
- **Python scikit-learn**: CTR 模型训练（Phase 5）

## 基础设施

- **Docker & Docker Compose**: 容器化
- **MinIO / 阿里云 OSS**: 对象存储
- **Prometheus**: 指标采集
- **Grafana**: 监控面板
- **Zap**: 结构化日志

---

# 📝 下一步行动

建议从 **Phase 1** 开始，逐步迭代：

1. **立即可做**:
   - ✅ 配置管理（已完成）
   - → 搭建 Gin API 框架
   - → 接入通义万相 API
   - → 实现基础图像处理

2. **本周目标**:
   - 完成第一张 1:1 广告图生成
   - 部署到本地测试

3. **下周目标**:
   - 支持 3 种尺寸
   - 添加 CTA 按钮

---

**需要我立即开始实现 Phase 1 吗？**

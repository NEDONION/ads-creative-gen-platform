# 🎨 AI 广告创意生成平台

> 基于阿里云通义万相和千问的智能广告创意生成平台，支持文案生成、创意生成、AB 实验和模型追踪

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev/)
[![React Version](https://img.shields.io/badge/React-18+-61dafb.svg)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7+-3178c6.svg)](https://www.typescriptlang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791.svg)](https://www.postgresql.org/)

## ✨ 功能特性

### 🎯 智能创意生成
- **完整的创意工作流** - 文案生成 → 确认 → 图片生成
- **AI 文案生成** - 基于通义千问自动生成 CTA 和卖点
- **AI 图像生成** - 集成通义万相生成专业广告图
- **多尺寸支持** - 支持 1:1、9:16、16:9 等多种广告尺寸
- **多变体生成** - 一次生成多个创意变体供选择

### 🧪 AB 实验平台
- **实验管理** - 创建和管理创意 AB 实验
- **智能分流** - 基于分桶的流量分配机制
- **实时指标** - 曝光、点击、CTR 实时统计
- **权重配置** - 支持自定义变体权重分配

### 📊 数据分析
- **任务管理** - 完整的任务生命周期管理
- **素材评分** - 质量评分、CTR 预测、NSFW 检测
- **模型追踪** - 记录所有 AI 模型调用链路
- **性能监控** - 追踪每个步骤的耗时和状态

### ☁️ 云服务集成
- **七牛云存储** - 自动上传并管理生成的素材
- **阿里云 RDS** - 生产环境 PostgreSQL 数据库
- **通义万相** - AI 图像生成
- **通义千问** - AI 文案生成

## 🏗️ 技术架构

### 后端技术栈
```
Go 1.21+ (Gin + GORM)
├── Web 框架: Gin
├── ORM: GORM
├── 数据库:
│   ├── PostgreSQL 14+ (生产环境 - 阿里云 RDS)
│   └── MySQL 8.0 (本地开发)
├── AI 服务:
│   ├── 阿里云通义万相 (图像生成)
│   └── 阿里云通义千问 (文案生成)
└── 存储: 七牛云对象存储
```

### 前端技术栈
```
React 18 + TypeScript 5.7 + Vite 6.0
├── UI 框架: React 18
├── 类型安全: TypeScript 5.7
├── 构建工具: Vite 6.0
├── 路由: React Router v6
└── HTTP 客户端: Axios
```

### 部署架构
- **前后端一体化** - Go 服务器托管前端静态文件
- **云原生部署** - 支持 Fly.io、Render 等云平台
- **容器化支持** - 提供 Docker 和 Docker Compose 配置

## 🚀 快速开始

### 环境要求

- **Go**: 1.21+
- **Node.js**: 18+
- **数据库**: PostgreSQL 14+ 或 MySQL 8.0+

### 1. 克隆项目

```bash
git clone https://github.com/your-org/ads-creative-gen-platform.git
cd ads-creative-gen-platform
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
vim .env
```

配置示例：
```bash
# 应用配置
APP_MODE=debug              # debug/release
HTTP_PORT=:4000

# 数据库配置
DB_TYPE=postgres            # postgres/mysql
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ads_creative_gen_platform
DB_CHARSET=utf8

# 通义 API 配置
TONGYI_API_KEY=your_tongyi_api_key
TONGYI_IMAGE_MODEL=wanx-v1
TONGYI_LLM_MODEL=qwen-turbo

# 七牛云配置
QINIU_ACCESS_KEY=your_qiniu_access_key
QINIU_SECRET_KEY=your_qiniu_secret_key
QINIU_BUCKET=your_bucket_name
QINIU_DOMAIN=your_cdn_domain
QINIU_REGION=cn-south-1
QINIU_BASE_PATH=s3/
```

### 3. 初始化数据库

```bash
# PostgreSQL
psql -U postgres -c "CREATE DATABASE ads_creative_gen_platform;"

# 或 MySQL
mysql -u root -p -e "CREATE DATABASE ads_creative_gen_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行数据库迁移（首次安装或更新表结构时）
go run cmd/migrate/main.go -action migrate

# 添加默认数据（可选）
go run cmd/migrate/main.go -action seed
```

> **迁移命令说明**：
> - `migrate`: 创建/更新表结构（不会删除现有数据）
> - `seed`: 添加默认数据（管理员账号、默认标签等）
> - `reset`: ⚠️ 重置数据库（删除所有数据并重新初始化）

### 4. 启动后端服务

```bash
# 安装依赖
go mod download

# 启动服务
go run main.go
```

### 5. 构建并启动前端

```bash
cd web

# 安装依赖
npm install

# 构建前端（生产环境）
npm run build

# 或启动开发服务器
npm run dev
```

### 6. 访问应用

**生产模式（前后端一体化）：**
- 应用地址: http://localhost:4000
- API 接口: http://localhost:4000/api/v1
- 健康检查: http://localhost:4000/health

**开发模式（前后端分离）：**
- 前端: http://localhost:5173 (Vite dev server)
- 后端: http://localhost:4000

## 📁 项目结构

```
ads-creative-gen-platform/
├── cmd/                        # 命令行工具
│   └── migrate/               # 数据库迁移工具
├── config/                    # 配置管理
│   ├── config.go             # 配置加载器
│   └── sql/                  # 迁移 SQL 文件
├── internal/
│   ├── handlers/             # HTTP 处理器
│   │   ├── creative_handler.go    # 创意生成接口
│   │   ├── experiment_handler.go  # AB 实验接口
│   │   └── trace_handler.go       # 模型追踪接口
│   ├── services/             # 业务逻辑层
│   │   ├── creative_service.go    # 创意生成服务
│   │   ├── copywriting_service.go # 文案生成服务
│   │   ├── experiment_service.go  # 实验管理服务
│   │   ├── tongyi_client.go       # 通义万相客户端
│   │   ├── qwen_client.go         # 通义千问客户端
│   │   └── qiniu_service.go       # 七牛云客户端
│   ├── models/               # 数据模型
│   │   ├── creative.go       # 创意任务和素材
│   │   ├── experiment.go     # AB 实验
│   │   └── trace.go          # 模型追踪
│   └── middleware/           # 中间件
│       └── cors.go           # CORS 配置
├── pkg/
│   └── database/             # 数据库连接层
├── web/                      # React 前端
│   ├── src/
│   │   ├── pages/           # 页面组件
│   │   │   ├── DashboardPage.tsx
│   │   │   ├── CreativeGeneratorPage.tsx
│   │   │   ├── TasksPage.tsx
│   │   │   ├── AssetsPage.tsx
│   │   │   ├── ExperimentsPage.tsx
│   │   │   └── TracePage.tsx
│   │   ├── components/      # 可复用组件
│   │   ├── services/        # API 客户端
│   │   └── types/           # TypeScript 类型
│   └── dist/                # 构建产物（前端静态文件）
├── docs/                    # 📚 文档中心
│   ├── README.md           # 文档索引
│   ├── api-reference.md    # API 接口文档
│   ├── database.md         # 数据库设计
│   ├── guides/             # 开发和部署指南
│   │   ├── development.md  # 开发指南
│   │   └── deployment.md   # 部署指南
│   └── design/             # 功能设计文档
│       ├── copywriting-feature.md
│       ├── experiment-feature.md
│       ├── model-trace-page.md
│       └── plugin-widget.md
├── scripts/                # 🔧 管理脚本
│   ├── start.sh           # 启动服务
│   ├── stop.sh            # 停止服务
│   └── status.sh          # 查看状态
├── Dockerfile             # Docker 镜像配置
├── fly.toml              # Fly.io 部署配置
├── render.yaml           # Render 部署配置
├── docker-compose.yml    # Docker Compose 配置
├── .env.example          # 环境变量模板
└── main.go              # 程序入口
```

## 🎯 核心 API

### 文案生成工作流

```bash
# 1. 生成文案候选
POST /api/v1/copywriting/generate
{
  "product_name": "夏季清凉T恤"
}

# 2. 确认文案并启动创意生成
POST /api/v1/copywriting/confirm
{
  "product_name": "夏季清凉T恤",
  "cta": "立即抢购",
  "selling_point": "纯棉透气，清凉一夏",
  "requested_formats": ["1:1", "9:16"],
  "num_variants": 3
}
```

### 任务管理

```bash
# 查询任务状态
GET /api/v1/creative/task/{task_id}

# 获取任务列表
GET /api/v1/creative/tasks?page=1&page_size=10

# 删除任务
DELETE /api/v1/creative/task/{task_id}
```

### 素材管理

```bash
# 获取素材列表
GET /api/v1/creative/assets?page=1&page_size=20&format=1:1
```

### AB 实验

```bash
# 创建实验
POST /api/v1/experiments
{
  "name": "夏季T恤广告测试",
  "description": "测试不同风格的广告效果",
  "variants": [
    {"creative_id": "uuid1", "weight": 0.5},
    {"creative_id": "uuid2", "weight": 0.5}
  ]
}

# 分流分配
GET /api/v1/experiments/{id}/assign?user_id=user123

# 记录曝光
POST /api/v1/experiments/{id}/hit
{
  "user_id": "user123",
  "variant_id": "variant_uuid"
}

# 记录点击
POST /api/v1/experiments/{id}/click
{
  "user_id": "user123",
  "variant_id": "variant_uuid"
}

# 查看实验指标
GET /api/v1/experiments/{id}/metrics
```

### 模型追踪

```bash
# 查询追踪列表
GET /api/v1/model_traces

# 获取追踪详情
GET /api/v1/model_traces/{id}
```

## 📚 完整文档

- [📖 文档中心](docs/README.md) - 完整文档导航
- [🔌 API 参考](docs/api-reference.md) - 详细的 REST API 接口说明
- [💾 数据库设计](docs/database.md) - 数据库表结构和 ER 图
- [👨‍💻 开发指南](docs/guides/development.md) - 本地开发环境配置
- [🚀 部署指南](docs/guides/deployment.md) - 云端部署和自建服务器部署

## ☁️ 部署方式

### 方式一：Fly.io 部署（推荐）

```bash
# 安装 Fly CLI
brew install flyctl

# 登录
flyctl auth login

# 部署应用
flyctl deploy
```

**特点**：
- ✅ 全球部署，香港节点
- ✅ 自动 HTTPS
- ✅ 自动休眠节省费用

详见：[Fly.io 部署指南](docs/guides/deployment.md#方式一-flyio-部署推荐)

### 方式二：Render 部署

```bash
# 1. 连接 GitHub 仓库到 Render
# 2. 使用 render.yaml 自动配置
# 3. 在 Dashboard 设置环境变量
# 4. 自动部署
```

**特点**：
- ✅ GitHub 集成，推送自动部署
- ✅ 零配置
- ✅ 免费层可用

详见：[Render 部署指南](docs/guides/deployment.md#方式二-render-部署)

### 方式三：Docker 部署

```bash
# 启动所有服务（MySQL、Redis、MinIO）
docker-compose up -d

# 查看状态
docker-compose ps

# 停止服务
docker-compose down
```

详见：[Docker 部署指南](docs/guides/deployment.md#方式四-docker-部署)

### 方式四：自建服务器

使用 Systemd + Nginx 部署，详见：[生产环境部署指南](docs/guides/deployment.md#方式五-生产环境部署自建服务器)

## 🔧 管理脚本

项目提供了便捷的管理脚本（位于 `scripts/` 目录）：

```bash
# 启动服务（自动检查数据库、运行迁移）
./scripts/start.sh

# 停止服务
./scripts/stop.sh

# 查看状态
./scripts/status.sh
```

## 📊 数据库

### 核心数据表

- **用户与项目**
  - `users` - 用户表
  - `projects` - 项目表
  - `tags` - 标签表

- **创意生成**
  - `creative_tasks` - 创意任务表
  - `creative_assets` - 素材表
  - `creative_scores` - 评分表

- **AB 实验**
  - `experiments` - 实验主表
  - `experiment_variants` - 实验变体表
  - `experiment_metrics` - 实验指标表

- **模型追踪**
  - `model_traces` - 模型追踪主表
  - `model_trace_steps` - 追踪步骤表

详见：[数据库设计文档](docs/database.md)

## 🛠️ 开发指南

### 后端开发

```bash
# 安装依赖
go mod download

# 运行数据库迁移（如需更新表结构）
go run cmd/migrate/main.go -action migrate

# 启动开发服务器
go run main.go

# 运行测试
go test ./...
```

### 前端开发

```bash
cd web

# 安装依赖
npm install

# 启动开发服务器（热重载）
npm run dev

# 构建生产版本
npm run build

# TypeScript 类型检查
npx tsc --noEmit

# 预览生产构建
npm run preview
```

### 代码规范

- 后端遵循 Go 标准代码规范
- 前端使用 TypeScript 严格模式
- 使用 ESLint 和 Prettier 格式化代码

## 🔐 环境配置

### 必需的环境变量

```bash
# 数据库
DB_TYPE=postgres
DB_HOST=your_db_host
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=ads_creative_gen_platform

# 通义 API
TONGYI_API_KEY=your_api_key

# 七牛云
QINIU_ACCESS_KEY=your_access_key
QINIU_SECRET_KEY=your_secret_key
QINIU_BUCKET=your_bucket
```

### 可选的环境变量

```bash
# 应用配置
APP_MODE=release              # debug/release
HTTP_PORT=:8080               # 默认 :4000

# 通义模型配置
TONGYI_IMAGE_MODEL=wanx-v1    # 图像生成模型
TONGYI_LLM_MODEL=qwen-turbo   # 文案生成模型

# 七牛云配置
QINIU_DOMAIN=cdn.example.com  # 自定义域名
QINIU_REGION=cn-south-1       # 存储区域
QINIU_BASE_PATH=s3/           # 存储路径前缀
```

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发流程

```bash
# 1. Fork 项目
# 2. 创建功能分支
git checkout -b feature/your-feature

# 3. 提交更改
git commit -m "feat: add your feature"

# 4. 推送到分支
git push origin feature/your-feature

# 5. 创建 Pull Request
```

### 提交规范

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构代码
- `test`: 测试相关
- `chore`: 构建/工具相关

## 🆘 常见问题

### 1. 数据库连接失败

**问题**：服务启动时提示数据库连接失败

**解决方案**：
```bash
# 检查数据库是否运行
# PostgreSQL
pg_isready

# MySQL
mysqladmin ping

# 检查环境变量配置
cat .env | grep DB_
```

### 2. 前端访问 404

**问题**：访问前端页面返回 404

**解决方案**：
```bash
# 确保前端已构建
cd web && npm run build

# 检查 web/dist 目录是否存在
ls -la web/dist/

# 重启后端服务
go run main.go
```

### 3. API Key 配置错误

**问题**：通义 API 调用失败

**解决方案**：
```bash
# 检查 API Key 配置
echo $TONGYI_API_KEY

# 测试 API 连接
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
```

### 4. 七牛云上传失败

**问题**：图片上传到七牛云失败

**解决方案**：
```bash
# 检查七牛云配置
cat .env | grep QINIU_

# 确保 Bucket 存在且有权限
# 检查存储区域配置是否正确
```

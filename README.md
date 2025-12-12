# AI 广告创意生成平台

一键生成多尺寸广告图，自动排序推荐最优创意。

## 项目简介

这是一个基于 Go + Gin + GORM + PostgreSQL/MySQL 的广告创意生成平台，结合 AI 生成技术，帮助用户快速创建广告创意素材。系统支持自动建表、数据库迁移、云端存储和智能推荐等功能。

## 功能特性

- **创意生成**：根据产品信息和卖点自动生成创意素材
- **多尺寸支持**：支持 1:1、16:9 等多种常见尺寸
- **任务管理**：支持创意生成任务的创建和状态查询
- **资源存储**：集成七牛云对象存储服务
- **智能评分**：对生成的创意素材进行质量评分
- **自动建表**：使用 GORM 自动创建和更新数据库表结构
- **跨库兼容**：支持 PostgreSQL 和 MySQL

## 技术栈

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL 12+/MySQL 8+ (GORM ORM)
- **AI Service**: 阿里云通义千问万相 API
- **Storage**: 七牛云对象存储
- **Frontend**: HTML/CSS/JS
- **Build**: Go Modules
- **Container**: Docker

## 系统架构

```
Frontend → API Gateway → Gin Server → GORM → PostgreSQL/MySQL
                                   ↓
                            七牛云存储 ← AI API
```

## 快速开始

### 环境准备

1. **Go**: 1.20+
2. **Database**: PostgreSQL 12+ 或 MySQL 8+
3. **AI Service**: 通义千问 API 密钥
4. **Storage**: 七牛云存储账号（可选）

### 1. 克隆项目

```bash
git clone <repository-url>
cd ads-creative-gen-platform
```

### 2. 数据库配置（重要）

#### PostgreSQL 配置
```ini
# config/config.ini
[database]
Db = postgres
DbHost = 127.0.0.1
DbPort = 5432
DbUser = postgres
DbPassWord = postgres
DbName = ads_creative_gen_platform
Charset = utf8  # Only used for MySQL
```

#### MySQL 配置
```ini
# config/config.ini
[database]
Db = mysql
DbHost = 127.0.0.1
DbPort = 3306
DbUser = root
DbPassWord = yourpassword
DbName = ads_creative_gen_platform
Charset = utf8mb4
```

#### 重要数据库配置说明：

**自动建表机制**：
- 系统在启动时自动创建所有必要的表结构
- 使用 GORM 的 `AutoMigrate` 功能，无需手动建表
- 支持增量迁移，只更新缺失的表和字段

**Schema 配置**：
- PostgreSQL 自动使用 `search_path=public` 确保表创建在正确 schema
- 跨数据库兼容性：使用 `varchar` 而非 `enum` 确保 MySQL/PostgreSQL 兼容

**外键约束**：
- 系统确保所有外键关系正确
- 防止数据不一致和引用完整性问题
- 在创建相关记录前必须先创建父记录

### 3. 环境变量配置

复制 `.env.example` 到 `.env` 并填写配置：

```bash
cp .env.example .env
```

#### 必需环境变量：
```bash
# 通义千问 API 配置
TONGYI_API_KEY=sk-your-key-here
TONGYI_IMAGE_MODEL=wanx-v1
TONGYI_LLM_MODEL=qwen-turbo

# 服务配置
APP_MODE=debug
HTTP_PORT=:4000
```

#### 可选环境变量（七牛云）：
```bash
# 七牛云配置
QINIU_ACCESS_KEY=your-access-key
QINIU_SECRET_KEY=your-secret-key
QINIU_BUCKET=your-bucket-name
QINIU_DOMAIN=http://your-domain.clouddn.com  # 自定义域名
QINIU_PUBLIC_CLOUD_DOMAIN=http://t75ejhbs9.hn-bkt.clouddn.com  # 公共访问域名
QINIU_REGION=cn-south-1
QINIU_BASE_PATH=s3/
```

### 4. 启动数据库

**Option 1: Using Docker (Recommended)**
```bash
docker-compose up -d postgres
```

**Option 2: Using Local PostgreSQL**
确保 PostgreSQL 已安装并运行，然后确认配置文件指向本地实例。

### 5. 初始化数据库

系统会在启动时自动执行数据库初始化，但你也可以手动执行：

```bash
# 执行数据库迁移（创建表结构）
go run cmd/migrate/main.go -action migrate

# 添加默认数据（创建管理员账户等）
go run cmd/migrate/main.go -action seed

# 重置数据库（删除所有数据，重新开始）
go run cmd/migrate/main.go -action reset
```

**数据库初始化流程**：
1. 自动连接到数据库
2. 自动创建所有表结构（如果不存在）
3. 自动创建默认管理员账户和标签
4. 确保外键约束完整

### 6. 运行服务

```bash
go run main.go
```

服务启动后会自动：
- 加载配置文件
- 连接数据库
- 自动创建表结构
- 添加默认数据
- 启动 Web 服务器

访问 http://localhost:4000/health 检查服务状态

## 数据库设计详解

### 表结构概览

**用户表 (users)**:
- id, uuid, created_at, updated_at, deleted_at
- username, email, password_hash, phone, avatar_url
- role, status, last_login_at

**项目表 (projects)**:
- id, uuid, created_at, updated_at, deleted_at
- name, description, owner_id, status, settings
- 外键: owner_id → users.id

**标签表 (tags)**:
- id, created_at, updated_at, deleted_at
- name, category, color, usage_count

**创意任务表 (creative_tasks)**:
- id, uuid, created_at, updated_at, deleted_at
- user_id, project_id, title, selling_points, product_image_url
- requested_formats, requested_styles, num_variants, cta_text
- status, progress, error_message, queued_at, started_at, completed_at
- 外键: user_id → users.id, project_id → projects.id

**创意素材表 (creative_assets)**:
- id, uuid, created_at, updated_at, deleted_at
- task_id, format, width, height, file_size, storage_type
- **public_url** (存储完整访问 URL，如 http://t75ejhbs9.hn-bkt.clouddn.com/s3/...)
- original_path, style, variant_index, generation_prompt
- 外键: task_id → creative_tasks.id

**创意评分表 (creative_scores)**:
- id, creative_id, quality_overall, brightness_score, contrast_score
- sharpness_score, composition_score, color_harmony_score
- ctr_prediction (CTR 预测), ctr_confidence, model_version
- cl_ip_score, aesthetic_score, nsfw_score, is_safe
- 外键: creative_id → creative_assets.id

**项目成员表 (project_members)**:
- id, created_at, updated_at, deleted_at
- project_id, user_id, role
- 外键: project_id → projects.id, user_id → users.id

### 数据库连接池配置
- 最大空闲连接: 10
- 最大打开连接: 100
- 连接最大生命周期: 1小时

### URL 存储策略
- **public_url**: 存储完整的访问 URL（拼接后的）
- 优先级: QINIU_PUBLIC_CLOUD_DOMAIN > QINIU_DOMAIN > 默认域名
- 示例: `http://t75ejhbs9.hn-bkt.clouddn.com/s3/2025/12/12/filename.png`

## API 接口文档

### 创意生成接口
```
POST /api/v1/creative/generate
Content-Type: application/json

{
  "title": "夏季促销活动",
  "selling_points": ["5折优惠", "限时抢购", "独家优惠"],
  "product_image_url": "https://example.com/product.jpg",
  "formats": ["1:1", "16:9"],
  "style": "bright",
  "cta_text": "立即购买",
  "num_variants": 3
}
```

### 查询任务状态
```
GET /api/v1/creative/task/{task-id}
```

### 获取素材列表
```
GET /api/v1/creative/assets?page=1&page_size=20&format=1:1&task_id={task-id}
```

## 七牛云配置（可选）

系统支持七牛云对象存储，用于存储生成的图片素材：

```
QINIU_ACCESS_KEY=your-access-key
QINIU_SECRET_KEY=your-secret-key
QINIU_BUCKET=your-bucket-name
QINIU_DOMAIN=http://your-domain.clouddn.com  # 自定义域名
QINIU_PUBLIC_CLOUD_DOMAIN=http://t75ejhbs9.hn-bkt.clouddn.com  # 公共访问域名
QINIU_REGION=cn-south-1  # 区域标识
QINIU_BASE_PATH=s3/     # 存储路径前缀
```

**URL 生成逻辑**:
1. 优先使用 `QINIU_PUBLIC_CLOUD_DOMAIN` (公共云域名)
2. 其次使用 `QINIU_DOMAIN` (自定义域名)  
3. 最后使用默认的七牛云域名格式

## 项目结构

```
ads-creative-gen-platform/
├── cmd/migrate/        # 数据库迁移工具
├── config/             # 配置文件
├── docs/               # 文档
├── internal/           # 内部代码
│   ├── models/         # 数据模型
│   ├── handlers/       # API 处理器
│   └── services/       # 业务逻辑
├── pkg/database/       # 数据库相关工具
├── web/                # 前端界面
│   └── index.html      # Web 界面
├── main.go             # 入口文件
├── config/config.ini   # 配置文件
├── .env.example        # 环境变量示例
├── docker-compose.yml  # Docker 配置
├── go.mod
├── go.sum
└── README.md
```

## 开发指南

### 数据库开发注意事项
1. **模型定义**: 在 `internal/models/` 中定义所有模型
2. **GORM 标签**: 使用适当的 GORM 标签定义字段
3. **迁移函数**: 使用 `database.MigrateTables()` 进行自动迁移
4. **外键关系**: 使用 GORM 关联定义表关系
5. **数据类型**: 使用 `varchar` 而非 `enum` 确保跨数据库兼容

### 数据库迁移工具
- `go run cmd/migrate/main.go -action migrate`: 迁移表结构
- `go run cmd/migrate/main.go -action seed`: 添加默认数据
- `go run cmd/migrate/main.go -action reset`: 重置数据库

## Docker 部署

### 使用 Docker Compose
```bash
docker-compose up -d
```

### 手动构建镜像
```bash
docker build -t ads-creative-platform .
```

### 运行容器
```bash
docker run -d \
  -p 4000:4000 \
  -v ./config:/app/config \
  -v ./.env:/app/.env \
  --name ads-platform \
  ads-creative-platform
```

## 环境变量说明

| 变量 | 说明 | 默认值 |
|------|------|--------|
| TONGYI_API_KEY | 通义千问 API 密钥 | - |
| TONGYI_IMAGE_MODEL | 图片生成模型 | wanx-v1 |
| TONGYI_LLM_MODEL | 语言模型 | qwen-turbo |
| APP_MODE | 应用模式 | debug |
| HTTP_PORT | HTTP 端口 | :4000 |
| QINIU_ACCESS_KEY | 七牛云 Access Key | - |
| QINIU_SECRET_KEY | 七牛云 Secret Key | - |
| QINIU_BUCKET | 七牛云存储空间 | ads-creative-gen-platform |
| QINIU_DOMAIN | 七牛云自定义域名 | - |
| QINIU_PUBLIC_CLOUD_DOMAIN | 七牛云公共访问域名 | - |
| QINIU_REGION | 七牛云区域 | cn-south-1 |
| QINIU_BASE_PATH | 存储路径前缀 | s3/ |

## Web 界面

前端界面已移动到单独的目录 `web/` 中，文件为 `index.html`，您可以通过以下方式使用：

- 直接在浏览器中打开 `web/index.html` 文件
- 或通过本地服务器提供服务（如 Python 的 `python -m http.server` 或 Node 的 `npx serve`）

## 测试

```bash
# 健康检查
curl http://localhost:4000/health

# Ping
curl http://localhost:4000/api/v1/ping
```

## 故障排除

### 数据库连接问题
1. 确认 `config/config.ini` 中数据库配置正确
2. 确认数据库服务正在运行
3. 检查网络连接和防火墙设置
4. 查看应用启动日志获取详细错误信息

### 数据库迁移问题
- 执行 `go run cmd/migrate/main.go -action reset` 重置数据库
- 检查表结构和外键约束
- 确认 GORM 模型定义正确

### 文件存储问题
- 确认七牛云配置正确
- 检查存储空间权限设置
- 确认公共域名可正常访问

## 默认账号

数据库管理员：
- 用户名: `admin`
- 密码: `admin123`

## 开发计划

- [x] 数据库设计
- [x] 基础框架
- [x] 自动建表机制
- [x] 七牛云集成
- [ ] 图像生成（进行中）
- [ ] 多尺寸布局
- [ ] CTR 预测

## 文档

- [API 文档](./API.md)
- [实施计划](./docs/implementation-plan.md)
- [数据库指南](./docs/database-guide.md)
- [Docker 指南](./docs/docker-guide.md)
- [端口配置](./docs/PORTS.md)

## License

MIT
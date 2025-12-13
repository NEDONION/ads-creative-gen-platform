# 🎨 AI 广告创意生成平台

> 一键生成多尺寸广告图，AI 智能排序推荐最优创意

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev/)
[![React Version](https://img.shields.io/badge/React-18+-61dafb.svg)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178c6.svg)](https://www.typescriptlang.org/)

## ✨ 功能特性

- **🎯 智能创意生成** - 基于商品信息自动生成多风格广告创意
- **📐 多尺寸支持** - 支持 1:1、9:16 等多种广告尺寸
- **📊 任务管理** - 完整的任务创建、查询、进度跟踪
- **☁️ 云端存储** - 集成七牛云对象存储
- **🎨 现代化 UI** - React + TypeScript 工程化前端
- **🔄 实时更新** - WebSocket 实时任务状态推送
- **📈 数据分析** - 素材质量评分与性能分析

## 🏗️ 技术栈

### 后端
- **Go 1.21+** - 高性能后端服务
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **MySQL 8.0** - 关系型数据库
- **阿里云通义万相** - AI 图像生成

### 前端
- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **React Router** - 路由管理
- **Axios** - HTTP 客户端

## 🚀 快速开始

### 方式一: 使用脚本启动（推荐）

```bash
# 1. 配置环境
cp config/config.ini.example config/config.ini
vim config/config.ini  # 填入数据库和 API 密钥

# 2. 启动后端（自动初始化数据库）
./scripts/start.sh

# 3. 启动前端
cd web
npm install
npm run dev
```

### 方式二: Docker 部署

```bash
# 启动所有服务（MySQL、Redis、MinIO等）
docker-compose up -d

# 查看状态
docker-compose ps
```

### 访问应用

- **前端**: http://localhost:3000
- **后端**: http://localhost:4000
- **API 文档**: http://localhost:4000/api/v1/ping

## 📁 项目结构

```
ads-creative-gen-platform/
├── cmd/                    # 命令行工具
│   └── migrate/           # 数据库迁移
├── config/                # 配置文件
├── docs/                  # 📚 文档目录
│   ├── api.md            # API 接口文档
│   ├── database.md       # 数据库设计
│   ├── development.md    # 开发指南
│   └── deployment.md     # 部署指南
├── internal/
│   ├── handlers/         # HTTP 处理器
│   ├── middleware/       # 中间件
│   ├── models/          # 数据模型
│   └── services/        # 业务逻辑
├── pkg/
│   └── database/        # 数据库连接
├── scripts/             # 🔧 管理脚本
│   ├── start.sh        # 启动服务
│   ├── stop.sh         # 停止服务
│   └── status.sh       # 查看状态
├── web/                 # 前端项目
│   ├── src/
│   │   ├── components/ # React 组件
│   │   ├── pages/     # 页面
│   │   ├── services/  # API 服务
│   │   └── types/     # TypeScript 类型
│   └── package.json
└── main.go             # 主入口

```

## 📚 文档

- [API 接口文档](docs/api.md) - 详细的 API 接口说明
- [数据库设计](docs/database.md) - 数据库表结构和 ER 图
- [开发指南](docs/development.md) - 本地开发环境配置
- [部署指南](docs/deployment.md) - 生产环境部署方案

## 🔧 管理脚本

项目提供了便捷的管理脚本（位于 `scripts/` 目录）:

```bash
# 启动服务
./scripts/start.sh

# 停止服务
./scripts/stop.sh

# 查看状态
./scripts/status.sh
```

## 🎯 核心 API

### 创建创意生成任务

```bash
POST /api/v1/creative/generate
Content-Type: application/json

{
  "title": "夏季清凉T恤",
  "selling_points": ["纯棉面料", "透气舒适", "多色可选"],
  "requested_formats": ["1:1", "9:16"],
  "num_variants": 3
}
```

### 查询任务状态

```bash
GET /api/v1/creative/task/{task_id}
```

### 获取任务列表

```bash
GET /api/v1/creative/tasks?page=1&page_size=10
```

### 获取素材列表

```bash
GET /api/v1/creative/assets?page=1&page_size=20&format=1:1
```

## 🔐 配置说明

编辑 `config/config.ini`:

```ini
[app]
AppMode = debug
HttpPort = :4000

[mysql]
DbHost = 127.0.0.1
DbPort = 3306
DbUser = root
DbPassWord = your_password
DbName = ads_creative_platform

[tongyi]
ApiKey = your_tongyi_api_key

[qiniu]
AccessKey = your_qiniu_access_key
SecretKey = your_qiniu_secret_key
Bucket = your_bucket_name
Domain = your_cdn_domain
```

## 🛠️ 开发

### 后端开发

```bash
# 安装依赖
go mod download

# 运行迁移
go run cmd/migrate/main.go -action reset

# 启动开发服务器
go run main.go
```

### 前端开发

```bash
cd web

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# TypeScript 类型检查
npx tsc --noEmit
```

## 📊 数据库

### 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE ads_creative_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行迁移（自动创建表）
go run cmd/migrate/main.go -action migrate

# 初始化种子数据
go run cmd/migrate/main.go -action seed

# 或一次性完成
go run cmd/migrate/main.go -action reset
```

### 核心表

- `users` - 用户表
- `creative_tasks` - 创意任务表
- `creative_assets` - 素材表
- `creative_scores` - 评分表

详见 [数据库设计文档](docs/database.md)

## 🐳 Docker 部署

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

包含的服务:
- MySQL 8.0 (:3306)
- Redis 7 (:6379)
- MinIO (:9000, :9001)
- phpMyAdmin (:8081)

## 🤝 贡献

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

## 🆘 常见问题

### 启动失败

1. 检查 MySQL 是否运行
2. 检查配置文件中的数据库连接信息
3. 查看日志: `tail -f logs/app.log`

### 前端无法连接后端

1. 检查后端是否在 4000 端口运行
2. 检查 `web/vite.config.js` 中的代理配置

### 数据库迁移失败

```bash
# 重置数据库
go run cmd/migrate/main.go -action reset
```

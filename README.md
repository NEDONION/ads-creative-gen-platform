# AI 多尺寸广告创意生成平台

> 一个端到端广告创意自动化平台：输入商品信息，自动生成多尺寸广告图，并根据 CTR 预测排序

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql&logoColor=white)](https://www.mysql.com)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## 📚 文档导航

- [项目设计文档](./docs/design.md) - 原始项目愿景
- [实施计划](./docs/implementation-plan.md) - 分阶段开发路线图
- [数据库设计](./docs/database-schema.sql) - 完整的SQL建表语句
- [数据库使用指南](./docs/database-guide.md) - 表结构说明与查询示例
- [ER关系图](./docs/database-er-diagram.md) - 实体关系图

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- MySQL 8.0+
- 通义万相 API Key ([申请地址](https://help.aliyun.com/zh/dashscope/))

### 1. 克隆项目

```bash
git clone <your-repo>
cd ads-creative-gen-platform
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env`：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的通义 API Key：

```env
# 通义 API 配置
TONGYI_API_KEY=sk-your-api-key-here
TONGYI_IMAGE_MODEL=wanx-v1
TONGYI_LLM_MODEL=qwen-turbo

# 服务配置
SERVER_PORT=8080
ENVIRONMENT=development
```

### 3. 配置数据库

编辑 `config/config.ini`：

```ini
[mysql]
Db = mysql
DbHost = 127.0.0.1
DbPort = 3306
DbUser = root
DbPassWord = your_password
DbName = ads_creative_platform
Charset = utf8mb4
```

### 4. 启动数据库

**推荐：使用 Docker** 🐳

```bash
# 一键启动 MySQL + Redis + RabbitMQ + MinIO
docker-compose up -d

# 查看服务状态
docker-compose ps
```

**或者：使用本地 MySQL**

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE ads_creative_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行完整的 SQL 文件
mysql -u root -p < docs/database-schema.sql
```

详见 [Docker 部署指南](./docs/docker-guide.md)

### 5. 安装依赖

```bash
go mod download
```

### 6. 运行数据库迁移

```bash
# 创建所有表
go run cmd/migrate/main.go -action migrate

# 初始化默认数据（管理员账号、默认标签等）
go run cmd/migrate/main.go -action seed
```

### 7. 运行项目

```bash
go run main.go
```

你应该看到类似输出：

```
✓ App config loaded (Mode: debug, Port: :4000)
✓ MySQL config loaded (Database: ads_creative_platform)
✓ RabbitMQ config loaded
✓ Etcd config loaded
✓ Tongyi config loaded (Model: wanx-v1)
✓ All configurations loaded successfully
```

---

## 🗂️ 项目结构

```
ads-creative-gen-platform/
├── cmd/                          # 命令行工具
│   └── migrate/                  # 数据库迁移工具
│       └── main.go
├── config/                       # 配置文件
│   ├── config.go                 # 配置加载逻辑
│   └── config.ini                # 配置文件
├── docs/                         # 文档
│   ├── design.md                 # 原始设计文档
│   ├── implementation-plan.md    # 实施计划
│   ├── database-schema.sql       # 数据库建表SQL
│   ├── database-guide.md         # 数据库使用指南
│   └── database-er-diagram.md    # ER关系图
├── internal/                     # 内部代码
│   ├── handlers/                 # HTTP 处理器
│   ├── models/                   # 数据模型
│   │   ├── base.go              # 基础模型
│   │   ├── user.go              # 用户模型
│   │   ├── creative.go          # 创意模型
│   │   ├── project.go           # 项目模型
│   │   └── tag.go               # 标签模型
│   ├── services/                 # 业务逻辑
│   └── storage/                  # 存储层
├── pkg/                          # 公共包
│   └── database/                 # 数据库
│       └── mysql.go              # MySQL连接与迁移
├── uploads/                      # 本地上传目录
├── .env                          # 环境变量（不提交）
├── .env.example                  # 环境变量模板
├── .gitignore                    # Git忽略文件
├── go.mod                        # Go模块定义
├── go.sum                        # 依赖锁定
├── main.go                       # 入口文件
└── README.md                     # 本文件
```

---

## 📊 数据库概览

### 核心表

| 表名 | 说明 | 记录数（估计） |
|------|------|-------------|
| `users` | 用户表 | 1000+ |
| `creative_tasks` | 创意任务表 | 10万+ |
| `creative_assets` | 创意素材表 | 50万+ |
| `creative_scores` | 评分表 | 50万+ |
| `creative_performance` | 性能数据表 | 100万+ |

### 默认数据

运行 `seed` 后会自动创建：

- **管理员账号**
  - 用户名: `admin`
  - 密码: `admin123`
  - 邮箱: `admin@example.com`

- **默认标签**
  - 行业: 电商、游戏、金融、教育
  - 风格: 极简风、活力风、专业风

---

## 🔧 数据库管理命令

```bash
# 创建所有表
go run cmd/migrate/main.go -action migrate

# 初始化默认数据
go run cmd/migrate/main.go -action seed

# 重置数据库（⚠️ 会删除所有数据）
go run cmd/migrate/main.go -action reset
```

---

## 📐 技术架构

### 后端技术栈

- **Go 1.21+**: 核心服务
- **Gin**: Web 框架（Phase 1）
- **GORM**: ORM
- **MySQL 8.0+**: 关系数据库
- **Redis**: 缓存与任务队列（Phase 4）
- **RabbitMQ**: 消息队列（Phase 4）

### AI 模型

- **通义万相**: 图像生成
- **通义千问**: 文案生成
- **CLIP**: 图文匹配评分（Phase 3）

### 基础设施（规划中）

- **MinIO / 阿里云OSS**: 对象存储
- **Prometheus + Grafana**: 监控
- **Docker**: 容器化部署

---

## 📝 开发路线图

详见 [实施计划](./docs/implementation-plan.md)

### Phase 1: MVP（1-2周）✅ 进行中

- [x] 项目初始化
- [x] 配置管理（ini + env）
- [x] 数据库设计与建表
- [x] GORM 模型定义
- [ ] Gin API 框架搭建
- [ ] 通义万相 API 集成
- [ ] 基础图像处理

### Phase 2: 多尺寸支持（2-3周）

- [ ] 支持 1:1, 4:5, 9:16, 1200x628 等尺寸
- [ ] 智能裁剪与自适应布局
- [ ] CTA 按钮生成
- [ ] Logo 自动放置

### Phase 3: 智能排序（2周）

- [ ] 质量评分系统
- [ ] CTR 预测模型
- [ ] 创意排序

### Phase 4: 生产化（2-3周）

- [ ] 任务队列（Redis）
- [ ] 对象存储（OSS）
- [ ] 监控与日志（Prometheus + Grafana）
- [ ] Docker 容器化

### Phase 5: 高级特性（3-4周）

- [ ] A/B 测试管理
- [ ] 实际 CTR 数据回传
- [ ] 自动化创意优化

---

## 🔐 安全建议

1. **修改默认管理员密码**
   ```sql
   -- 首次登录后立即修改
   UPDATE users SET password_hash = '新的哈希' WHERE username = 'admin';
   ```

2. **不要提交 .env 文件**
   ```bash
   # .gitignore 已配置，但请确保：
   git status  # 不应看到 .env
   ```

3. **定期备份数据库**
   ```bash
   mysqldump -u root -p ads_creative_platform > backup_$(date +%Y%m%d).sql
   ```

---

## 🆘 常见问题

### Q: 如何重置数据库？

```bash
go run cmd/migrate/main.go -action reset
```

### Q: 数据库连接失败？

检查 `config/config.ini` 中的数据库配置是否正确：
- 端口号（3306 vs 4306）
- 用户名和密码
- 数据库名是否已创建

### Q: 通义 API 调用失败？

检查 `.env` 文件：
- `TONGYI_API_KEY` 是否正确
- 是否有足够的额度
- 网络连接是否正常

# AI 广告创意生成平台

一键生成多尺寸广告图，自动排序推荐最优创意。

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置

**编辑 `.env`**
```env
TONGYI_API_KEY=sk-your-key-here
# 七牛云配置（可选）
QINIU_ACCESS_KEY=your-access-key
QINIU_SECRET_KEY=your-secret-key
QINIU_BUCKET=your-bucket-name
QINIU_DOMAIN=http://your-domain.clouddn.com  # 你的公共访问域名
QINIU_REGION=zone-identifier  # 如 cn-south-1
```

**编辑 `config/config.ini`**
```ini
[database]
Db = postgres
DbHost = 127.0.0.1
DbPort = 5432
DbUser = postgres
DbPassWord = postgres
DbName = ads_creative_gen_platform
Charset = utf8  # Only used for MySQL
```

### 3. 启动数据库

**Option 1: Using Docker (Recommended)**
```bash
docker-compose up -d postgres
```

**Option 2: Using Local PostgreSQL**
Make sure PostgreSQL is installed and running on your system, then ensure the configuration points to your local instance.

### 4. 初始化数据库

```bash
go run cmd/migrate/main.go -action migrate
go run cmd/migrate/main.go -action seed
```

### 5. 运行服务

```bash
go run main.go
```

访问 http://localhost:4000/health

## Web 界面

前端界面已移动到单独的目录 `web/` 中，文件为 `index.html`，您可以通过以下方式使用：

- 直接在浏览器中打开 `web/index.html` 文件
- 或通过本地服务器提供服务（如 Python 的 `python -m http.server` 或 Node 的 `npx serve`）

## API 文档

见 [API.md](./API.md)

## 测试

```bash
# 健康检查
curl http://localhost:4000/health

# Ping
curl http://localhost:4000/api/v1/ping
```

## 项目结构

```
├── cmd/migrate/        # 数据库迁移工具
├── config/             # 配置
├── docs/               # 文档
├── internal/           # 内部代码
│   ├── models/         # 数据模型
│   ├── handlers/       # API 处理器
│   └── services/       # 业务逻辑
├── pkg/database/       # 数据库
├── web/                # 前端界面
│   └── index.html      # Web 界面
├── main.go             # 入口
└── docker-compose.yml  # Docker 配置
```

## 技术栈

- Go 1.20+ + Gin
- PostgreSQL 12+
- 阿里云通义万相
- Docker

## 开发计划

- [x] 数据库设计
- [x] 基础框架
- [ ] 图像生成（进行中）
- [ ] 多尺寸布局
- [ ] CTR 预测

## 文档

- [实施计划](./docs/implementation-plan.md)
- [数据库指南](./docs/database-guide.md)
- [Docker 指南](./docs/docker-guide.md)
- [端口配置](./docs/PORTS.md)

## 默认账号

数据库管理员：
- 用户名: `admin`
- 密码: `admin123`

## License

MIT

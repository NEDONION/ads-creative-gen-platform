# 开发指南

## 🛠️ 开发环境设置

### 前置要求

- **Go**: 1.21+
- **Node.js**: 18+
- **MySQL**: 8.0+
- **Git**: 2.0+

### 克隆项目

```bash
git clone https://github.com/your-org/ads-creative-gen-platform.git
cd ads-creative-gen-platform
```

---

## 📦 后端开发

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境

复制配置文件:
```bash
cp config/config.ini.example config/config.ini
```

编辑 `config/config.ini`:
```ini
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
```

### 3. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE ads_creative_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行迁移
go run cmd/migrate/main.go -action reset
```

### 4. 启动后端服务

```bash
# 开发模式
go run main.go

# 或使用脚本
./scripts/start.sh
```

服务将运行在 `http://localhost:4000`

### 5. 测试 API

```bash
# 健康检查
curl http://localhost:4000/health

# 测试 Ping
curl http://localhost:4000/api/v1/ping
```

---

## 🎨 前端开发

### 1. 进入前端目录

```bash
cd web
```

### 2. 安装依赖

```bash
npm install
```

### 3. 启动开发服务器

```bash
npm run dev
```

前端将运行在 `http://localhost:3000`

### 4. 构建生产版本

```bash
npm run build
```

---

## 📂 项目结构

```
ads-creative-gen-platform/
├── cmd/
│   └── migrate/           # 数据库迁移工具
├── config/                # 配置文件
├── docs/                  # 文档
├── internal/
│   ├── handlers/          # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   └── services/          # 业务逻辑
├── pkg/
│   └── database/          # 数据库连接
├── scripts/               # 脚本文件
├── web/                   # 前端代码
│   ├── src/
│   │   ├── components/    # React 组件
│   │   ├── pages/         # 页面
│   │   ├── services/      # API 服务
│   │   └── types/         # TypeScript 类型
│   └── package.json
├── main.go               # 主入口
└── README.md
```

---

## 🔧 常用开发任务

### 添加新的 API 端点

1. 在 `internal/handlers/` 创建处理器
2. 在 `internal/services/` 实现业务逻辑
3. 在 `main.go` 注册路由

示例:
```go
// internal/handlers/example_handler.go
func (h *ExampleHandler) GetExample(c *gin.Context) {
    c.JSON(200, gin.H{"message": "example"})
}

// main.go
v1.GET("/example", exampleHandler.GetExample)
```

### 添加新的数据模型

1. 在 `internal/models/` 定义模型
2. 运行迁移创建表

```go
// internal/models/example.go
type Example struct {
    UUIDModel
    Name string `gorm:"type:varchar(255)" json:"name"`
}

func (Example) TableName() string {
    return "examples"
}
```

### 添加前端页面

1. 在 `web/src/pages/` 创建页面组件
2. 在 `web/src/App.tsx` 添加路由

```tsx
// web/src/pages/ExamplePage.tsx
const ExamplePage: React.FC = () => {
  return <Layout title="Example">...</Layout>;
};

// web/src/App.tsx
<Route path="/example" element={<ExamplePage />} />
```

---

## 🧪 测试

### 后端测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/services/...

# 带覆盖率
go test -cover ./...
```

### 前端测试

```bash
cd web

# 运行测试
npm test

# 类型检查
npm run type-check
```

---

## 📝 代码规范

### Go 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 错误处理不能被忽略
- 导出的函数和类型必须有注释

```bash
# 格式化代码
gofmt -w .

# 检查代码
go vet ./...
```

### TypeScript 代码规范

- 使用 ESLint 和 Prettier
- 所有组件必须有类型定义
- Props 和 State 必须定义接口

```bash
# 检查代码
npm run lint

# 自动修复
npm run lint:fix
```

---

## 🔍 调试

### 后端调试

使用 Delve 调试器:
```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug main.go
```

### 前端调试

- 使用浏览器开发者工具
- React DevTools 扩展
- 查看网络请求

---

## 📊 数据库操作

### 查看当前数据

```bash
mysql -u root -p ads_creative_platform
```

```sql
-- 查看所有任务
SELECT * FROM creative_tasks ORDER BY created_at DESC LIMIT 10;

-- 查看所有素材
SELECT * FROM creative_assets ORDER BY created_at DESC LIMIT 10;
```

### 重置数据库

```bash
go run cmd/migrate/main.go -action reset
```

---

## 🚀 提交代码

### Git 工作流

```bash
# 创建功能分支
git checkout -b feature/your-feature-name

# 提交更改
git add .
git commit -m "feat: add your feature"

# 推送分支
git push origin feature/your-feature-name
```

### Commit 规范

使用 [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` 新功能
- `fix:` 修复bug
- `docs:` 文档更新
- `style:` 代码格式
- `refactor:` 重构
- `test:` 测试
- `chore:` 构建/工具变动

---

## 🆘 常见问题

**Q: 启动后端报错 "database connection failed"**

检查 MySQL 是否运行:
```bash
mysql -u root -p -e "SELECT 1;"
```

检查配置文件 `config/config.ini` 中的数据库信息是否正确。

**Q: 前端启动报错 "port 3000 already in use"**

```bash
# 查找占用端口的进程
lsof -ti:3000

# 终止进程
kill -9 $(lsof -ti:3000)
```

**Q: Go 依赖下载失败**

```bash
# 设置 Go 代理
go env -w GOPROXY=https://goproxy.cn,direct

# 重新下载
go mod download
```

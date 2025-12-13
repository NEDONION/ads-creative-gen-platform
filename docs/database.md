# 数据库设计文档

## 📊 概览

**数据库名称**: `ads_creative_platform`

**字符集**: `utf8mb4_unicode_ci`

**核心表**: 11+ 张表

---

## 🗂️ ER 关系图

```
┌─────────────┐
│   users     │
└──────┬──────┘
       │ 1
       │ N
┌──────▼──────────┐      ┌──────────────┐
│ creative_tasks  │ 1  N │ creative_    │
│                 ├──────►  assets      │
└─────────────────┘      └──────┬───────┘
                                │ 1
                                │ 1
                         ┌──────▼──────────┐
                         │ creative_scores │
                         └─────────────────┘
```

### 关系说明

1. **用户 - 任务** (1:N): 一个用户可以创建多个任务
2. **任务 - 素材** (1:N): 一个任务生成多个素材
3. **素材 - 评分** (1:1): 每个素材有质量评分

---

## 📐 表结构分类

### 1. 用户与权限 (3张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `users` | 用户表 | username, email, role |
| `projects` | 项目表 | name, owner_id |
| `project_members` | 项目成员 | project_id, user_id, role |

### 2. 创意生成核心 (3张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `creative_tasks` | 任务表 | uuid, title, status, progress |
| `creative_assets` | 素材表 | format, public_url, rank |
| `creative_scores` | 评分表 | ctr_prediction, quality_overall |

### 3. 辅助功能 (5+张表)

- `creative_templates` - 创意模板
- `tags` / `creative_tags` - 标签系统
- `user_quotas` - 用户配额
- `api_keys` - API密钥
- `audit_logs` - 操作审计

---

## 🚀 快速开始

### 1. 创建数据库

```bash
mysql -u root -p -e "CREATE DATABASE ads_creative_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 2. 运行迁移

```bash
# 执行迁移（创建所有表）
go run cmd/migrate/main.go -action migrate

# 初始化种子数据
go run cmd/migrate/main.go -action seed

# 或一次性完成
go run cmd/migrate/main.go -action reset
```

### 3. 验证

```bash
mysql -u root -p ads_creative_platform -e "SHOW TABLES;"
```

---

## 📋 核心表详解

### creative_tasks (任务表)

**用途**: 存储创意生成请求

**核心字段**:
- `uuid` - 任务唯一标识
- `title` - 商品标题
- `selling_points` - 卖点列表 (JSON)
- `requested_formats` - 请求的尺寸 (JSON: ["1:1", "9:16"])
- `status` - 任务状态 (pending/processing/completed/failed)
- `progress` - 进度 (0-100)

**状态流转**:
```
pending → queued → processing → completed
                              ↘ failed
```

**查询示例**:
```sql
-- 获取用户任务
SELECT * FROM creative_tasks
WHERE user_id = 1
ORDER BY created_at DESC;

-- 进行中的任务
SELECT * FROM creative_tasks
WHERE status IN ('queued', 'processing');
```

### creative_assets (素材表)

**用途**: 存储生成的广告图

**核心字段**:
- `format` - 尺寸 (1:1, 9:16, etc.)
- `public_url` - 公开访问URL
- `storage_type` - 存储类型 (qiniu/local)
- `generation_prompt` - 生成提示词
- `rank` - 排名

**查询示例**:
```sql
-- 获取任务素材（按排名）
SELECT * FROM creative_assets
WHERE task_id = 1
ORDER BY rank ASC;

-- 统计各尺寸生成量
SELECT format, COUNT(*) as count
FROM creative_assets
GROUP BY format;
```

### creative_scores (评分表)

**用途**: 质量评分和CTR预测

**核心字段**:
- `quality_overall` - 综合质量 (0-1)
- `ctr_prediction` - CTR预测 (0-1)
- `brightness_score`, `contrast_score` - 各维度评分

---

## 🔍 常用查询

### 获取任务及素材

```sql
SELECT
    ct.uuid,
    ct.title,
    ct.status,
    COUNT(ca.id) as asset_count
FROM creative_tasks ct
LEFT JOIN creative_assets ca ON ct.id = ca.task_id
WHERE ct.user_id = 1
GROUP BY ct.id
ORDER BY ct.created_at DESC;
```

### 获取Top创意

```sql
SELECT
    ca.uuid,
    ca.format,
    ca.public_url,
    cs.ctr_prediction,
    ca.rank
FROM creative_assets ca
JOIN creative_scores cs ON ca.id = cs.creative_id
WHERE ca.task_id = 1
ORDER BY ca.rank ASC
LIMIT 5;
```

---

## 📈 索引策略

### 主要索引

```sql
-- 用户表
INDEX idx_username (username)
INDEX idx_email (email)

-- 任务表
INDEX idx_user (user_id)
INDEX idx_status (status)
INDEX idx_created (created_at)

-- 素材表
INDEX idx_task (task_id)
INDEX idx_format (format)
INDEX idx_rank (rank)
```

---

## 🔐 数据库配置

编辑 `config/config.ini`:

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

---

## 📚 GORM 使用示例

### 创建任务

```go
task := models.CreativeTask{
    UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
    UserID:    1,
    Title:     "商品标题",
    SellingPoints: models.StringArray{"卖点1", "卖点2"},
    RequestedFormats: models.StringArray{"1:1", "9:16"},
    Status:    models.TaskPending,
}
db.Create(&task)
```

### 查询任务及关联数据

```go
var task models.CreativeTask
db.Preload("Assets").Preload("Assets.Score").First(&task, "uuid = ?", taskUUID)
```

### 更新任务状态

```go
db.Model(&task).Updates(map[string]interface{}{
    "status": models.TaskCompleted,
    "progress": 100,
    "completed_at": time.Now(),
})
```

---

## 🔧 维护

### 清理过期任务

```sql
DELETE FROM creative_tasks
WHERE status = 'failed'
AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
```

### 备份数据库

```bash
mysqldump -u root -p ads_creative_platform > backup_$(date +%Y%m%d).sql
```

---

## 🆘 常见问题

**Q: 如何重置数据库？**
```bash
go run cmd/migrate/main.go -action reset
```

**Q: 默认管理员账号？**
- 用户名: `admin`
- 密码: `admin123`
- 邮箱: `admin@example.com`

**Q: 如何查看表结构？**
```bash
mysql -u root -p ads_creative_platform -e "DESC creative_tasks;"
```

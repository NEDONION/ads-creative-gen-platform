# 数据库设计与使用指南

## 📊 数据库概览

**数据库名称**: `ads_creative_platform`

**字符集**: `utf8mb4_unicode_ci`

**表数量**: 11+ 核心表

---

## 🗂️ 表结构分类

### 1. 用户与权限管理 (3张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `users` | 用户表 | username, email, role, status |
| `projects` | 项目/团队表 | name, owner_id, status |
| `project_members` | 项目成员表 | project_id, user_id, role |

### 2. 创意生成核心 (3张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `creative_tasks` | 创意任务表 | title, status, progress |
| `creative_assets` | 创意素材表 | format, file_path, rank |
| `creative_scores` | 评分表 | ctr_prediction, quality_overall |

### 3. 性能与实验 (3张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `creative_performance` | 投放表现表 | impressions, clicks, ctr |
| `ab_experiments` | A/B实验表 | name, status, winner_variant_id |
| `ab_variants` | 实验变体表 | experiment_id, creative_id |

### 4. 辅助功能 (5张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `creative_templates` | 创意模板表 | name, category, layout_config |
| `tags` | 标签表 | name, category, color |
| `creative_tags` | 创意标签关联表 | creative_id, tag_id |
| `user_quotas` | 用户配额表 | max_tasks_per_day, tasks_today |
| `api_keys` | API密钥表 | key_hash, permissions |

### 5. 审计与监控 (3张表)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `audit_logs` | 操作审计表 | action, resource_type, ip_address |
| `system_task_logs` | 系统任务日志 | task_type, status, duration |
| `copy_library` | 文案库表 | category, text, avg_ctr |

---

## 🚀 快速开始

### 1. 配置数据库连接

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

### 2. 创建数据库

```bash
mysql -u root -p -e "CREATE DATABASE ads_creative_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

或者直接执行 SQL 文件：

```bash
mysql -u root -p < docs/database-schema.sql
```

### 3. 运行数据库迁移

```bash
# 安装依赖
go mod download

# 执行迁移（创建所有表）
go run cmd/migrate/main.go -action migrate

# 初始化默认数据
go run cmd/migrate/main.go -action seed

# 或者一次性完成迁移+种子数据
go run cmd/migrate/main.go -action reset
```

### 4. 验证数据库

```bash
mysql -u root -p ads_creative_platform -e "SHOW TABLES;"
```

应该看到类似输出：
```
+-------------------------------+
| Tables_in_ads_creative_platform |
+-------------------------------+
| users                         |
| projects                      |
| creative_tasks                |
| creative_assets               |
| ...                           |
+-------------------------------+
```

---

## 📐 核心表详解

### creative_tasks (创意任务表)

**用途**: 存储用户提交的创意生成请求

**核心字段**:
- `uuid`: 任务唯一标识
- `title`: 商品标题
- `selling_points`: 卖点列表 (JSON)
- `requested_formats`: 请求的尺寸 (JSON: ["1:1", "4:5", "9:16"])
- `status`: 任务状态 (pending → queued → processing → completed)
- `progress`: 进度百分比 (0-100)

**生命周期**:
```
pending → queued → processing → completed
                              ↘ failed
```

**查询示例**:
```sql
-- 获取用户的所有任务
SELECT * FROM creative_tasks WHERE user_id = 1 ORDER BY created_at DESC;

-- 获取进行中的任务
SELECT * FROM creative_tasks WHERE status IN ('queued', 'processing');

-- 统计任务状态分布
SELECT status, COUNT(*) as count FROM creative_tasks GROUP BY status;
```

### creative_assets (创意素材表)

**用途**: 存储生成的每一张广告图

**核心字段**:
- `format`: 尺寸格式 (1:1, 4:5, 9:16, 1200x628)
- `file_path`: 文件存储路径
- `public_url`: 公开访问URL
- `generation_prompt`: 生成时使用的提示词
- `rank`: 排名（基于CTR预测）

**查询示例**:
```sql
-- 获取某任务的所有素材（按排名）
SELECT * FROM creative_assets
WHERE task_id = 1
ORDER BY rank ASC;

-- 统计各尺寸生成量
SELECT format, COUNT(*) as count
FROM creative_assets
GROUP BY format;

-- 获取高分素材
SELECT ca.*, cs.ctr_prediction
FROM creative_assets ca
JOIN creative_scores cs ON ca.id = cs.creative_id
WHERE cs.ctr_prediction > 0.7
ORDER BY cs.ctr_prediction DESC;
```

### creative_scores (评分表)

**用途**: 存储质量评分和CTR预测

**核心字段**:
- `quality_overall`: 综合质量评分 (0-1)
- `ctr_prediction`: CTR预测值 (0-1)
- `brightness_score`, `contrast_score`, `sharpness_score`: 各维度评分

**查询示例**:
```sql
-- 获取平均质量评分
SELECT AVG(quality_overall) as avg_quality FROM creative_scores;

-- 找出低质量素材
SELECT creative_id, quality_overall
FROM creative_scores
WHERE quality_overall < 0.5;
```

---

## 🔍 常用查询

### 1. 获取用户的创意生成历史

```sql
SELECT
    ct.uuid,
    ct.title,
    ct.status,
    ct.created_at,
    COUNT(ca.id) as asset_count
FROM creative_tasks ct
LEFT JOIN creative_assets ca ON ct.id = ca.task_id
WHERE ct.user_id = 1
GROUP BY ct.id
ORDER BY ct.created_at DESC;
```

### 2. 获取Top排名的创意

```sql
SELECT
    ca.uuid,
    ca.format,
    ca.public_url,
    cs.ctr_prediction,
    cs.quality_overall,
    ca.rank
FROM creative_assets ca
JOIN creative_scores cs ON ca.id = cs.creative_id
WHERE ca.task_id = 1
ORDER BY ca.rank ASC
LIMIT 5;
```

### 3. 统计创意性能

```sql
SELECT
    ca.format,
    AVG(cp.ctr) as avg_ctr,
    SUM(cp.impressions) as total_impressions,
    SUM(cp.clicks) as total_clicks
FROM creative_assets ca
JOIN creative_performance cp ON ca.id = cp.creative_id
WHERE cp.date >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
GROUP BY ca.format
ORDER BY avg_ctr DESC;
```

### 4. A/B实验结果

```sql
SELECT
    ab.name as experiment_name,
    av.variant_name,
    av.total_impressions,
    av.total_clicks,
    av.avg_ctr
FROM ab_experiments ab
JOIN ab_variants av ON ab.id = av.experiment_id
WHERE ab.id = 1
ORDER BY av.avg_ctr DESC;
```

---

## 🔧 数据库维护

### 清理过期任务

```sql
-- 删除30天前的失败任务
DELETE FROM creative_tasks
WHERE status = 'failed'
AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
```

### 重置用户每日配额

```sql
-- 每日定时任务执行
UPDATE user_quotas
SET tasks_today = 0,
    last_reset_at = CURDATE()
WHERE last_reset_at < CURDATE();
```

### 更新标签使用计数

```sql
UPDATE tags t
SET usage_count = (
    SELECT COUNT(*)
    FROM creative_tags ct
    WHERE ct.tag_id = t.id
);
```

---

## 📈 索引优化

已创建的关键索引：

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

-- 评分表
INDEX idx_ctr (ctr_prediction)
```

---

## 🔐 安全建议

1. **修改默认管理员密码**
```sql
-- 登录后立即修改
UPDATE users
SET password_hash = '$2a$10$NewHashHere'
WHERE username = 'admin';
```

2. **定期备份数据库**
```bash
mysqldump -u root -p ads_creative_platform > backup_$(date +%Y%m%d).sql
```

3. **使用只读用户进行查询**
```sql
CREATE USER 'readonly'@'%' IDENTIFIED BY 'password';
GRANT SELECT ON ads_creative_platform.* TO 'readonly'@'%';
```

---

## 📚 GORM 模型使用示例

### 创建任务

```go
task := models.CreativeTask{
    UUIDModel: models.UUIDModel{UUID: uuid.New().String()},
    UserID:    1,
    Title:     "户外露营帐篷",
    SellingPoints: models.StringArray{"防水", "三季通用"},
    RequestedFormats: models.StringArray{"1:1", "4:5", "9:16"},
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

## 🆘 常见问题

### Q: 如何重置数据库？

```bash
go run cmd/migrate/main.go -action reset
```

### Q: 如何只添加新表而不影响现有数据？

```bash
go run cmd/migrate/main.go -action migrate
```

### Q: 默认管理员账号是什么？

- **用户名**: admin
- **密码**: admin123
- **邮箱**: admin@example.com

### Q: 如何查看当前数据库版本？

```sql
SELECT version FROM system_config WHERE key = 'schema_version';
```

---

## 📝 更新日志

- **v1.0** (2024-01-15): 初始数据库设计
  - 11+ 核心表
  - 完整的关联关系
  - 索引优化
  - 默认数据初始化

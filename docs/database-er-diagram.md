# 数据库 ER 图

## 核心实体关系

```
┌─────────────┐
│   users     │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼──────────┐      ┌──────────────┐
│ creative_tasks  │ 1  N │ creative_    │
│                 ├──────►  assets      │
└─────────────────┘      └──────┬───────┘
                                │ 1
                                │
                                │ 1
                         ┌──────▼──────────┐
                         │ creative_scores │
                         └─────────────────┘


┌─────────────┐
│  projects   │
└──────┬──────┘
       │ 1
       │
       │ N
┌──────▼───────────┐
│ project_members  │
└──────────────────┘


┌──────────────┐      ┌──────────────┐
│ ab_experiments│ 1  N │ ab_variants  │
└──────────────┘◄─────┤              │
                      └──────┬───────┘
                             │ N
                             │
                             │ 1
                      ┌──────▼──────────┐
                      │ creative_assets │
                      └─────────────────┘


┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│ creative_    │ N  N │ creative_    │ N  1 │   tags       │
│  assets      ├──────►   tags       ├──────►              │
└──────────────┘      └──────────────┘      └──────────────┘


┌──────────────┐
│ creative_    │ 1
│  assets      ├────┐
└──────────────┘    │
                    │ N
             ┌──────▼──────────────┐
             │ creative_performance│
             └─────────────────────┘
```

## 表关系说明

### 1. 用户 - 任务关系 (1:N)
- 一个用户可以创建多个创意任务
- 每个任务属于一个用户

### 2. 任务 - 素材关系 (1:N)
- 一个任务可以生成多个素材（不同尺寸、风格）
- 每个素材属于一个任务

### 3. 素材 - 评分关系 (1:1)
- 每个素材有一条评分记录
- 评分包括质量分和CTR预测

### 4. 素材 - 标签关系 (N:N)
- 一个素材可以有多个标签
- 一个标签可以应用于多个素材
- 通过 `creative_tags` 中间表关联

### 5. 项目 - 成员关系 (1:N)
- 一个项目可以有多个成员
- 通过 `project_members` 表管理成员和权限

### 6. A/B实验 - 变体关系 (1:N)
- 一个实验包含多个变体
- 每个变体对应一个创意素材

### 7. 素材 - 性能数据关系 (1:N)
- 一个素材可以有多条性能记录（按日期、平台）
- 用于统计真实投放效果

## 关键外键约束

```sql
-- 任务外键
creative_tasks.user_id → users.id
creative_tasks.project_id → projects.id

-- 素材外键
creative_assets.task_id → creative_tasks.id

-- 评分外键
creative_scores.creative_id → creative_assets.id

-- 性能外键
creative_performance.creative_id → creative_assets.id

-- A/B实验外键
ab_variants.experiment_id → ab_experiments.id
ab_variants.creative_id → creative_assets.id

-- 项目成员外键
project_members.project_id → projects.id
project_members.user_id → users.id
```

## 索引策略

### 高频查询字段索引

```sql
-- 用户查询
users(username, email, status)

-- 任务查询
creative_tasks(user_id, status, created_at)

-- 素材查询
creative_assets(task_id, format, rank)

-- 评分查询
creative_scores(ctr_prediction, quality_overall)

-- 性能查询
creative_performance(creative_id, date, platform)
```

### 联合索引

```sql
-- 素材查询优化
creative_assets(task_id, rank)

-- 性能数据查询优化
creative_performance(creative_id, date, platform)

-- 项目成员查询优化
project_members(project_id, user_id)
```

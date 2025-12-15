# A/B 实验平台设计文档

## 文档版本
- 版本：v1.0
- 日期：2025-12-12
- 状态：设计已评审，待实现

---

## 1. 目标与范围
- 面向广告创意场景，提供 A/B 实验：同一商品多变体（创意）分流、曝光/点击上报、CTR 计算与显著性判断。
- 支持快速创建实验、管理变体权重、分流接口、埋点上报与结果查询。
- 兼容现有创意任务/素材表（creative_tasks / creative_assets）。

不包含：复杂多臂 bandit、自动调权，仅实现基础 A/B。

---

## 2. 数据模型
### 2.1 表结构（新增）
1) `experiments`
```
id (bigserial PK)
uuid (varchar(36) unique)
name (varchar(128))
product_name (varchar(255))
status (varchar(16)) -- draft/active/paused/archived
start_at (timestamptz, nullable)
end_at   (timestamptz, nullable)
created_at, updated_at, deleted_at
```

2) `experiment_variants`
```
id (bigserial PK)
experiment_id (FK -> experiments.id)
creative_id (FK -> creative_assets.id) -- 变体对应的创意素材
weight (numeric/float) -- 0-1，分流权重
bucket_start (int) -- 0-10000，闭区间
bucket_end   (int) -- 0-10000，闭区间
created_at, updated_at
UNIQUE(experiment_id, creative_id)
```

3) `experiment_metrics`
```
id (bigserial PK)
experiment_id (FK)
creative_id (FK)
impressions (bigint default 0)
clicks      (bigint default 0)
ctr         (numeric) -- 可选缓存
updated_at
UNIQUE(experiment_id, creative_id)
```

> 说明：bucket_start/end 用于静态分桶（0-10000），避免浮点累积误差；权重变更需重算桶区间。

### 2.2 GORM 结构体（示意）
```go
type Experiment struct {
  UUIDModel
  Name        string
  ProductName string
  Status      string // draft/active/paused/archived
  StartAt     *time.Time
  EndAt       *time.Time
  Variants    []ExperimentVariant
}

type ExperimentVariant struct {
  ID           uint
  ExperimentID uint
  CreativeID   uint
  Weight       float64
  BucketStart  int
  BucketEnd    int
}

type ExperimentMetric struct {
  ID           uint
  ExperimentID uint
  CreativeID   uint
  Impressions  int64
  Clicks       int64
  CTR          *float64
}
```

---

## 3. 分流与埋点逻辑
### 3.1 分流（assign）
- 输入：`experiment_id`、`user_key`（可选，若无则使用随机数）。
- 分桶算法：将 `user_key` 哈希映射到 [0,10000)，选择 bucket 覆盖的变体。
- 输出：`creative_id`，以及可选展示信息（图片/CTA/卖点）由前端再调用创意接口获取。
- 约束：仅 status=active 的实验可分流；paused/archived 返回 404/错误。

### 3.2 曝光/点击上报
- 曝光：`POST /experiments/:id/hit { creative_id }`
- 点击：`POST /experiments/:id/click { creative_id }`
- 计数策略：
  - 单次调用累加（可选幂等 token 防重复，MVP 可不做）。
  - 定期任务计算 CTR：`ctr = clicks / impressions`。

### 3.3 显著性计算（简版）
- 使用两比例 z-test 近似：
  - 输入：变体 A/B 的 impressions、clicks。
  - 输出：CTR_A、CTR_B、p-value、置信度（1-p）。
- 仅作为参考提示，不做自动调权。

---

## 4. API 设计
Base URL: `/api/v1`

1) 创建实验  
`POST /experiments`  
请求：
```json
{
  "name": "双十二华为手机测试",
  "product_name": "华为手机",
  "variants": [
    { "creative_id": 101, "weight": 0.5 },
    { "creative_id": 102, "weight": 0.5 }
  ]
}
```
响应：`{ code:0, data:{ experiment_id:"uuid" } }`

2) 更新状态  
`POST /experiments/:id/status` body: `{ "status":"active" }`

3) 分流  
`GET /experiments/:id/assign?user_key=abc123`  
响应：`{ code:0, data:{ creative_id:"..." } }`

4) 上报曝光  
`POST /experiments/:id/hit` body: `{ "creative_id":"..." }`

5) 上报点击  
`POST /experiments/:id/click` body: `{ "creative_id":"..." }`

6) 查询实验结果  
`GET /experiments/:id/metrics`  
响应：
```json
{
  "code":0,
  "data":{
    "experiment_id":"uuid",
    "variants":[
      {"creative_id":"...", "impressions":1000, "clicks":20, "ctr":0.02},
      {"creative_id":"...", "impressions":900, "clicks":30, "ctr":0.033}
    ],
    "winner":"creative_id" // 可选
  }
}
```

7) 列表/详情  
`GET /experiments`（分页、按状态筛选）  
`GET /experiments/:id`

---

## 5. 逻辑与实现要点
- **桶计算**：创建/更新变体时，根据权重累计生成 `[bucket_start, bucket_end]`，范围 0-10000。
- **分流幂等**：同一 `experimentId + user_key` 可选择“固定落桶”策略（哈希值决定桶），保证同用户一致性。
- **校验**：权重之和≈1（误差允许），变体 creative_id 必须存在且状态正常。
- **性能**：assign 接口仅查询 variants（带桶区间），无需 joins；hit/click 直接累加计数。
- **安全**：简单鉴权（可选）；参数校验；防止对非 active 实验上报。

---

## 6. 与现有表的关系
- 变体引用 `creative_assets.id`；可附带任务信息显示给前端。
- 可扩展：若要按任务级分流，可让 creative_id 指向任务首图/默认创意。

---

## 7. 开发步骤（建议）
1) 数据库：新增表 `experiments`、`experiment_variants`、`experiment_metrics`，AutoMigrate。
2) Service 层：
   - 创建/更新实验（含桶计算）
   - 分流 assign
   - hit/click 累加
   - metrics 计算（实时或查询时临时计算）
3) Handler / 路由：按 API 设计暴露。
4) 前端（实验管理页，可选）：创建实验、查看结果；或内嵌到 widget 使用 assign/hit/click。
5) 测试：用脚本模拟曝光/点击，验证计数和分流稳定性。

---

## 8. 后续增强
- 幂等 token 防重复曝光/点击。
- 事件流水表（detail log），用于离线分析。
- 自动调权（bandit），或多变体多臂扩展。
- 分流策略支持地理/设备/时间段定向。

# Ads Creative Gen Platform API 文档

本文档基于当前代码（Gin + `/api/v1` 路由）整理，所有成功响应统一格式：

```json
{
  "code": 0,
  "data": { ... },
  "message": ""   // 可选
}
```

失败会返回 HTTP 4xx/5xx，body 形如：`{"code":400,"message":"错误描述"}`。

- 服务基址：`http://localhost:4000`
- API 前缀：`/api/v1`
- 认证：当前无鉴权（TODO）
- 跨域：允许 `http://localhost:3000`、`http://localhost:3001`，并已开启 `Access-Control-Allow-Credentials: true`。

## 健康检查
- `GET /health` → `{ "status": "ok", "service": "ads-creative-platform" }`
- `GET /api/v1/ping` → `{ "message": "pong" }`

## 文案相关

### 生成文案候选
- `POST /api/v1/copywriting/generate`
- Body：
```json
{ "product_name": "苹果电脑" }
```
- 成功返回：
```json
{
  "code": 0,
  "data": {
    "task_id": "uuid",
    "cta_candidates": ["立即购买", "..."],
    "selling_point_candidates": ["轻薄", "..."]
  }
}
```

### 确认文案并写回任务
- `POST /api/v1/copywriting/confirm`
- Body（至少要有一个卖点或编辑卖点）：
```json
{
  "task_id": "uuid",
  "selected_cta_index": 0,
  "selected_sp_indexes": [0,1],
  "edited_cta": "可选，覆盖候选",
  "edited_sps": ["可选，覆盖选中卖点"],
  "product_image_url": "可选",
  "style": "可选",
  "num_variants": 2,
  "formats": ["1:1"]
}
```
- 返回：`{ "code":0, "data": { "task_id": "...", "status": "queued|draft|..." } }`

## 创意生成（图片/素材）

### 创建生成任务
- `POST /api/v1/creative/generate`
- Body：
```json
{
  "title": "商品标题",
  "selling_points": ["亮点1","亮点2"],
  "product_image_url": "可选，商品图 URL",
  "formats": ["1:1","16:9"],
  "style": "可选",
  "cta_text": "立即购买",
  "num_variants": 2
}
```
- 返回：`{ "task_id": "uuid", "status": "queued|draft|..." }`

### 启动生成（在确认文案后）
- `POST /api/v1/creative/start`
- Body：`{ "task_id": "uuid" }`
- 返回：`{ "task_id": "...", "status": "queued" }`

### 查询任务状态
- `GET /api/v1/creative/task/:id`
- 返回字段：`task_id,status,title,product_name,progress,error,created_at,completed_at,product_image_url,requested_formats,style,cta_text,num_variants,selling_points,creatives[]`
- `creatives` 元素：`{ id, format, image_url, width, height, title?, product_name?, cta_text?, selling_points? }`

### 删除任务
- `DELETE /api/v1/creative/task/:id`
- 返回：`{ "task_id": "...", "status": "deleted" }`

### 列出任务
- `GET /api/v1/creative/tasks?page=1&page_size=20&status=processing|completed|...`
- 返回：`{ tasks: [], total, page, page_size, total_pages }`

### 列出素材
- `GET /api/v1/creative/assets?page=1&page_size=20&format=1:1&task_id=...`
- 返回：`{ assets: [], total, page, page_size, total_pages }`
- `assets` 元素：`{ id/numeric_id?, task_id, format, width, height, storage_type, public_url, image_url?, title?, product_name?, cta_text?, selling_points?, created_at, updated_at }`

## 实验（A/B）

### 创建实验
- `POST /api/v1/experiments`
- Body：
```json
{
  "name": "首页 banner 实验",
  "product_name": "可选",
  "variants": [
    {
      "creative_id": 123,          // 或素材 uuid（前端 types 支持 string/number）
      "weight": 0.5,
      "bucket_start": 0,
      "bucket_end": 49,
      "title": "可选",
      "product_name": "可选",
      "image_url": "可选",
      "cta_text": "可选",
      "selling_points": ["可选"]
    }
  ]
}
```
- 返回：`{ "experiment_id": "uuid", "status": "draft|active|..." }`

### 实验列表
- `GET /api/v1/experiments?page=1&page_size=20&status=active|archived|draft`
- 返回：`{ experiments:[{ experiment_id,name,product_name,status,created_at,start_at,end_at,variants[] }], total, page, page_size }`
- `variants` 同创建时的结构。

### 更新实验状态（激活/停止/归档）
- `POST /api/v1/experiments/:id/status`
- Body：`{ "status": "active|stopped|archived" }`
- 返回：`{ experiment_id, status }`

### 分流获取命中素材（给插件/前端调用）
- `GET /api/v1/experiments/:id/assign?user_key=xxx`
- 返回：
```json
{
  "creative_id": 123,
  "asset_uuid": "可选",
  "task_id": 1,
  "title": "...",
  "product_name": "...",
  "cta_text": "...",
  "selling_points": ["..."],
  "image_url": "素材图 URL"
}
```

### 记录曝光
- `POST /api/v1/experiments/:id/hit`
- Body：`{ "creative_id": 123 }`
- 返回：`{ "status": "ok" }`

### 记录点击
- `POST /api/v1/experiments/:id/click`
- Body：`{ "creative_id": 123 }`
- 返回：`{ "status": "ok" }`

### 实验指标
- `GET /api/v1/experiments/:id/metrics`
- 返回：
```json
{
  "experiment_id": "...",
  "variants": [
    { "creative_id": 123, "impressions": 10, "clicks": 2, "ctr": 0.2 }
  ]
}
```

## 模型调用链路（Trace）

### 列表
- `GET /api/v1/model_traces?page=1&page_size=20&status=running|success|failed&model_name=qwen|tongyi&trace_id=...`
- 返回：`{ traces: [{ trace_id, model_name, model_version, status, duration_ms, start_at, end_at, source, input_preview?, output_preview?, error_message? }], total, page, page_size }`

### 详情
- `GET /api/v1/model_traces/:id`
- 返回：`{ trace_id, model_name, model_version, status, duration_ms, start_at, end_at, source, input_preview?, output_preview?, error_message?, steps: [...] }`
- `steps` 元素：`{ step_name, component, status, duration_ms, start_at, end_at, input_preview?, output_preview?, error_message? }`

---

> 如需补充其它域名的 CORS，请在 `internal/middleware/cors.go` 的 `allowedOrigins` 添加。当前 API 仍未接入鉴权，生产前需要加上认证/限流。 

# API 文档

## 基础信息

- **Base URL**: `http://localhost:4000`
- **版本**: v1
- **前缀**: `/api/v1`

---

## 接口列表

### 1. 健康检查

```http
GET /health
```

**响应**
```json
{
  "status": "ok",
  "service": "ads-creative-platform"
}
```

**测试**
```bash
curl http://localhost:4000/health
```

---

### 2. Ping 测试

```http
GET /api/v1/ping
```

**响应**
```json
{
  "message": "pong"
}
```

**测试**
```bash
curl http://localhost:4000/api/v1/ping
```

---

### 3. 创建创意任务 ✅

```http
POST /api/v1/creative/generate
```

**请求体**
```json
{
  "title": "户外露营帐篷",
  "selling_points": ["防水", "三季通用", "轻量化"],
  "product_image_url": "https://example.com/tent.jpg",
  "style": "modern",
  "num_variants": 1
}
```

**参数说明**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 商品标题 |
| selling_points | array | 否 | 卖点列表 |
| product_image_url | string | 否 | 商品图URL（可选） |
| style | string | 否 | 风格：modern/elegant/vibrant |
| num_variants | int | 否 | 生成数量，默认1 |

**响应**
```json
{
  "code": 0,
  "data": {
    "task_id": "60e293fa-95d1-4d46-931e-dd0fa96b025b",
    "status": "pending"
  }
}
```

**测试**
```bash
curl -X POST http://localhost:4000/api/v1/creative/generate \
  -H "Content-Type: application/json" \
  -d '{
    "title": "露营帐篷",
    "selling_points": ["防水", "轻量"],
    "style": "modern"
  }'
```

---

### 4. 查询任务状态 ✅

```http
GET /api/v1/creative/task/:id
```

**路径参数**
- `id`: 任务ID（UUID）

**响应（处理中）**
```json
{
  "code": 0,
  "data": {
    "task_id": "60e293fa-95d1-4d46-931e-dd0fa96b025b",
    "status": "processing",
    "title": "户外露营帐篷",
    "progress": 70
  }
}
```

**响应（已完成）**
```json
{
  "code": 0,
  "data": {
    "task_id": "60e293fa-95d1-4d46-931e-dd0fa96b025b",
    "status": "completed",
    "title": "户外露营帐篷",
    "progress": 100,
    "creatives": [
      {
        "id": "creative_uuid",
        "format": "1:1",
        "image_url": "https://dashscope.aliyuncs.com/...",
        "width": 1024,
        "height": 1024
      }
    ]
  }
}
```

**状态说明**

| 状态 | 说明 |
|------|------|
| pending | 等待处理 |
| processing | 生成中 |
| completed | 已完成 |
| failed | 失败 |

**测试**
```bash
# 替换 task_id
curl http://localhost:4000/api/v1/creative/task/60e293fa-95d1-4d46-931e-dd0fa96b025b
```

---

## 完整示例

### 1. 创建任务

```bash
TASK_ID=$(curl -s -X POST http://localhost:4000/api/v1/creative/generate \
  -H "Content-Type: application/json" \
  -d '{
    "title": "户外露营帐篷",
    "selling_points": ["防水", "三季通用", "轻量化"],
    "style": "modern"
  }' | jq -r '.data.task_id')

echo "Task ID: $TASK_ID"
```

### 2. 轮询状态

```bash
# 每5秒查询一次
while true; do
  STATUS=$(curl -s http://localhost:4000/api/v1/creative/task/$TASK_ID | jq -r '.data.status')
  PROGRESS=$(curl -s http://localhost:4000/api/v1/creative/task/$TASK_ID | jq -r '.data.progress')
  echo "Status: $STATUS, Progress: $PROGRESS%"

  if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
    break
  fi

  sleep 5
done
```

### 3. 获取结果

```bash
curl -s http://localhost:4000/api/v1/creative/task/$TASK_ID | jq '.data.creatives'
```

---

## 错误码

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 404 | 任务不存在 |
| 500 | 服务器错误 |

**错误响应示例**
```json
{
  "code": 400,
  "message": "Invalid request: title is required"
}
```

---

## 开发进度

- [x] 健康检查
- [x] Ping 测试
- [x] 创意生成 ✅
- [x] 任务查询 ✅
- [ ] 批量生成（Phase 2）
- [ ] 多尺寸支持（Phase 2）
- [ ] A/B 测试（Phase 5）

---

## 注意事项

1. **异步处理**: 创意生成是异步的，创建任务后需轮询状态
2. **超时时间**: 任务超时时间为 60 秒
3. **图片URL**: 返回的是通义万相的临时URL，建议下载保存
4. **用户认证**: 当前默认使用 user_id=1（管理员），后续将支持认证

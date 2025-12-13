# 千问LLM文案生成功能设计文档

## 文档版本
- **版本**: v1.0
- **日期**: 2025-12-12
- **状态**: 设计已批准，待实施

---

## 1. 项目概述

### 1.1 功能描述
为广告创意生成平台增加AI文案生成功能，用户只需输入商品名称，系统自动调用千问LLM生成候选CTA和核心卖点文案，用户选择并编辑后，继续生成广告图片创意。

### 1.2 核心价值
- **降低创作门槛**: 用户无需专业文案能力，AI自动生成创意文案
- **提高效率**: 从商品名称到完整创意一站式完成
- **灵活可控**: 支持用户手动编辑AI生成的文案
- **数据积累**: 保存所有候选文案，支持后续分析和优化

### 1.3 用户旅程
```
步骤1: 输入商品名称（如"智能手表 Pro"）
   ↓
步骤2: AI生成2个CTA + 3个核心卖点 → 用户选择 + 可编辑
   ↓
步骤3: 配置风格、变体数等参数 → 生成广告图片
   ↓
完成: 商品 + 文案 + 图片 = 完整创意
```

---

## 2. 系统架构设计

### 2.1 整体架构
```
┌─────────────────────────────────────────────────────────┐
│                     前端 (React)                         │
│  CreativeGeneratorPage (三步骤工作流)                    │
│    Step 1: 商品输入 → Step 2: 文案选择 → Step 3: 创意配置 │
└─────────────────────────────────────────────────────────┘
                           ↓ HTTP API
┌─────────────────────────────────────────────────────────┐
│                  后端 (Go + Gin)                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │  CreativeHandler                                  │  │
│  │  - GenerateCopywriting()                          │  │
│  │  - ConfirmCopywriting()                           │  │
│  │  - StartCreative()                                │  │
│  └───────────────────────────────────────────────────┘  │
│                           ↓                              │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Services                                         │  │
│  │  ┌──────────────────┐  ┌──────────────────────┐  │  │
│  │  │CopywritingService│  │CreativeService       │  │  │
│  │  │- Generate        │  │- StartGeneration     │  │  │
│  │  │- Confirm         │  │- processTask (async) │  │  │
│  │  └──────────────────┘  └──────────────────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
│                           ↓                              │
│  ┌───────────────────────────────────────────────────┐  │
│  │  External Clients                                 │  │
│  │  ┌──────────────┐    ┌──────────────────────┐    │  │
│  │  │QwenClient    │    │TongyiClient          │    │  │
│  │  │(LLM文案生成)  │    │(图片生成)             │    │  │
│  │  └──────────────┘    └──────────────────────┘    │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│              数据库 (PostgreSQL/MySQL)                   │
│  creative_tasks (存储任务、文案候选、最终文案)            │
│  creative_assets (存储生成的图片素材)                     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│           外部API服务                                     │
│  ┌──────────────────────┐  ┌──────────────────────────┐ │
│  │千问LLM API           │  │通义万相 API              │ │
│  │(文案生成)             │  │(图片生成)                │ │
│  │DashScope             │  │DashScope                 │ │
│  └──────────────────────┘  └──────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 2.2 数据流
```
[用户输入] → [前端验证] → [后端API]
                              ↓
                    [调用千问LLM API]
                              ↓
                    [存储候选文案到DB]
                              ↓
                    [返回候选给前端]
                              ↓
[用户选择编辑] → [前端收集] → [后端API]
                              ↓
                    [更新任务最终文案]
                              ↓
                    [调用通义万相API]
                              ↓
                    [存储图片素材到DB]
                              ↓
                    [返回创意结果]
```

---

## 3. 数据模型设计

### 3.1 CreativeTask表扩展

**现有字段**（保持不变）:
```go
type CreativeTask struct {
    UUIDModel
    UserID          uint
    ProjectID       *uint

    // 输入信息
    Title           string      // 创意标题
    SellingPoints   StringArray // 最终选定的卖点
    ProductImageURL string      // 产品图片URL
    BrandLogoURL    string      // 品牌Logo URL

    // 生成配置
    RequestedFormats StringArray // 请求的格式（如["1:1"]）
    RequestedStyles  StringArray // 请求的风格
    NumVariants      int         // 变体数量
    CTAText          string      // 最终选定的CTA

    // 任务状态
    Status           TaskStatus  // pending/queued/processing/completed/failed
    Progress         int         // 0-100
    ErrorMessage     string      // 错误信息

    // 时间统计
    QueuedAt           *time.Time
    StartedAt          *time.Time
    CompletedAt        *time.Time
    ProcessingDuration *int

    // 关联
    User    *User
    Project *Project
    Assets  []CreativeAsset
}
```

**新增字段**（本次实施）:
```go
// 文案生成相关字段
ProductName            string      `gorm:"type:varchar(255)" json:"product_name,omitempty"`
CTACandidates          StringArray `gorm:"type:json" json:"cta_candidates,omitempty"`
SellingPointCandidates StringArray `gorm:"type:json" json:"selling_point_candidates,omitempty"`
SelectedCTAIndex       *int        `json:"selected_cta_index,omitempty"`
SelectedSPIndexes      StringArray `gorm:"type:json" json:"selected_sp_indexes,omitempty"`
CopywritingGenerated   bool        `gorm:"default:false" json:"copywriting_generated"`
```

**字段说明**:
| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| ProductName | string | 用户输入的商品名称 | "智能手表 Pro" |
| CTACandidates | StringArray | LLM生成的2个CTA候选 | ["立即购买", "了解更多"] |
| SellingPointCandidates | StringArray | LLM生成的3个卖点候选 | ["续航48小时", "心率监测", "防水50米"] |
| SelectedCTAIndex | *int | 用户选择的CTA索引 | 0 (第一个) |
| SelectedSPIndexes | StringArray | 用户选择的卖点索引 | ["0", "2"] (第1和第3个) |
| CopywritingGenerated | bool | 标记是否使用AI生成文案 | true |

**数据流示例**:
```json
// 步骤1: 生成文案后
{
  "product_name": "智能手表 Pro",
  "cta_candidates": ["立即购买", "了解更多"],
  "selling_point_candidates": ["续航48小时", "心率监测", "防水50米"],
  "copywriting_generated": true,
  "status": "pending"
}

// 步骤2: 用户确认后
{
  "selected_cta_index": 0,
  "selected_sp_indexes": ["0", "2"],
  "cta_text": "立即购买",  // 或用户编辑后的值
  "selling_points": ["续航48小时", "防水50米"],  // 或用户编辑后的值
  "product_image_url": "https://...",
  "requested_styles": ["modern"],
  "num_variants": 3,
  "status": "queued"
}
```

### 3.2 数据库迁移
- **迁移方式**: GORM自动迁移（AutoMigrate）
- **兼容性**: 新字段允许NULL，不影响现有数据
- **回滚**: 如需回滚，手动删除新增字段即可

---

## 4. API设计

### 4.1 API端点概览

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| POST | `/api/v1/copywriting/generate` | 生成文案候选 | 需登录 |
| POST | `/api/v1/copywriting/confirm` | 确认文案选择 | 需登录 |
| POST | `/api/v1/creative/start` | 启动创意生成 | 需登录 |
| POST | `/api/v1/creative/generate` | 直接生成创意（旧流程） | 需登录 |

### 4.2 详细API规格

#### 4.2.1 生成文案候选

**请求**:
```http
POST /api/v1/copywriting/generate
Content-Type: application/json

{
  "product_name": "智能手表 Pro"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "cta_candidates": [
      "立即购买",
      "了解更多"
    ],
    "selling_point_candidates": [
      "续航48小时，告别频繁充电",
      "专业心率监测，守护健康",
      "防水50米，运动无忧"
    ]
  }
}
```

**错误响应**:
```json
{
  "code": 500,
  "message": "LLM API调用失败: rate limit exceeded",
  "data": null
}
```

#### 4.2.2 确认文案选择

**请求**:
```http
POST /api/v1/copywriting/confirm
Content-Type: application/json

{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "selected_cta_index": 0,
  "selected_sp_indexes": [0, 2],
  "edited_cta": "立即抢购",  // 可选，用户编辑后的CTA
  "edited_sps": [           // 可选，用户编辑后的卖点
    "超长续航48小时",
    "IPX8级防水"
  ],
  "product_image_url": "https://example.com/product.jpg",
  "style": "modern",
  "num_variants": 3,
  "formats": ["1:1"]
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending"
  }
}
```

#### 4.2.3 启动创意生成

**请求**:
```http
POST /api/v1/creative/start
Content-Type: application/json

{
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "queued"
  }
}
```

### 4.3 错误码定义

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 0 | 成功 | - |
| 400 | 请求参数错误 | 检查请求体格式 |
| 404 | 任务不存在 | 检查task_id是否正确 |
| 500 | 服务器内部错误 | 联系技术支持 |
| 503 | 外部API服务不可用 | 稍后重试 |

---

## 5. 千问LLM集成

### 5.1 API配置
```go
// 阿里云DashScope API
BaseURL: "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
Model:   "qwen-turbo"  // 从config.TongyiConfig.LLMModel读取
APIKey:  从config.TongyiConfig.APIKey读取
Timeout: 30秒
```

### 5.2 Prompt工程

**系统提示词**:
```
You are an expert advertising copywriter. Generate compelling ad copy in JSON format.
```

**用户提示词模板**:
```
生成产品广告文案: "{商品名称}"

请提供以下JSON格式的输出:
{
  "cta_options": ["CTA option 1", "CTA option 2"],
  "selling_point_options": ["Selling point 1", "Selling point 2", "Selling point 3"]
}

要求:
- 生成恰好2个CTA (Call-to-Action) 选项，使用中文
- 生成恰好3个核心卖点选项，使用中文
- CTA应简短有力（3-6个汉字），行动导向（例如: "立即购买", "马上抢购", "了解更多"）
- 卖点应简洁明了（8-15个汉字），突出产品核心优势
- 所有文本必须使用中文
- 只返回有效的JSON格式，不要包含其他文本
```

**示例输入**:
```
商品名称: "智能手表 Pro"
```

**期望输出**:
```json
{
  "cta_options": [
    "立即购买",
    "了解更多"
  ],
  "selling_point_options": [
    "续航48小时，告别频繁充电",
    "专业心率监测，守护健康",
    "防水50米，运动无忧"
  ]
}
```

### 5.3 响应解析策略

**处理流程**:
1. 检查LLM返回是否包含JSON
2. 移除可能的markdown代码块包裹（```json ... ```）
3. 提取JSON内容（查找第一个`{`和最后一个`}`）
4. 解析JSON
5. 验证字段存在性和数量（必须恰好2个CTA + 3个卖点）
6. 返回结构化结果或错误

**错误处理**:
- JSON解析失败 → 返回格式错误
- CTA数量不等于2 → 返回验证错误
- 卖点数量不等于3 → 返回验证错误
- 超时 → 返回超时错误并记录日志

### 5.4 QwenClient实现要点

```go
type QwenClient struct {
    apiKey  string
    model   string
    baseURL string
    client  *http.Client
}

// 核心方法
func (c *QwenClient) GenerateCopywriting(productName string) (*CopywritingResult, error)
func (c *QwenClient) buildPrompt(productName string) string
func (c *QwenClient) callAPI(req CopywritingRequest) (*CopywritingResponse, error)
func (c *QwenClient) parseResponse(response *CopywritingResponse) (*CopywritingResult, error)
func (c *QwenClient) extractJSON(content string) string
```

---

## 6. 服务层设计

### 6.1 CopywritingService

**职责**:
- 调用QwenClient生成文案
- 创建并保存CreativeTask记录
- 验证并更新用户的文案选择

**核心方法**:

```go
type CopywritingService struct {
    qwenClient *QwenClient
}

// 生成文案候选
func (s *CopywritingService) GenerateCopywriting(input GenerateCopywritingInput) (*GenerateCopywritingOutput, error) {
    // 1. 调用QwenClient
    result, err := s.qwenClient.GenerateCopywriting(input.ProductName)
    if err != nil {
        return nil, err
    }

    // 2. 创建任务记录
    task := models.CreativeTask{
        UUID:                   uuid.New().String(),
        UserID:                 input.UserID,
        ProductName:            input.ProductName,
        Title:                  input.ProductName,
        CTACandidates:          result.CTAOptions,
        SellingPointCandidates: result.SellingPointOptions,
        CopywritingGenerated:   true,
        Status:                 models.TaskPending,
        RequestedFormats:       []string{"1:1"},
        NumVariants:            3,
    }

    // 3. 保存到数据库
    database.DB.Create(&task)

    // 4. 返回结果
    return &GenerateCopywritingOutput{
        TaskID:                 task.UUID,
        CTACandidates:          result.CTAOptions,
        SellingPointCandidates: result.SellingPointOptions,
    }, nil
}

// 确认文案选择
func (s *CopywritingService) ConfirmCopywriting(input ConfirmCopywritingInput) (*models.CreativeTask, error) {
    // 1. 查找任务
    var task models.CreativeTask
    database.DB.Where("uuid = ?", input.TaskID).First(&task)

    // 2. 验证索引有效性
    if input.SelectedCTAIndex >= len(task.CTACandidates) {
        return nil, errors.New("invalid CTA index")
    }

    // 3. 准备最终文案
    finalCTA := input.EditedCTA
    if finalCTA == "" {
        finalCTA = task.CTACandidates[input.SelectedCTAIndex]
    }

    var finalSPs []string
    if len(input.EditedSPs) > 0 {
        finalSPs = input.EditedSPs
    } else {
        for _, idx := range input.SelectedSPIndexes {
            finalSPs = append(finalSPs, task.SellingPointCandidates[idx])
        }
    }

    // 4. 更新任务
    database.DB.Model(&task).Updates(map[string]interface{}{
        "cta_text":            finalCTA,
        "selling_points":      finalSPs,
        "selected_cta_index":  input.SelectedCTAIndex,
        "selected_sp_indexes": input.SelectedSPIndexes,
        "product_image_url":   input.ProductImageURL,
        "requested_styles":    []string{input.Style},
        "num_variants":        input.NumVariants,
    })

    return &task, nil
}
```

### 6.2 CreativeService扩展

**新增方法**:
```go
// 启动创意生成（从已确认文案的任务开始）
func (s *CreativeService) StartCreativeGeneration(taskUUID string) error {
    // 1. 查找任务
    var task models.CreativeTask
    if err := database.DB.Where("uuid = ?", taskUUID).First(&task).Error; err != nil {
        return fmt.Errorf("task not found: %w", err)
    }

    // 2. 验证任务包含必要数据
    if task.CTAText == "" || len(task.SellingPoints) == 0 {
        return errors.New("task missing copywriting data")
    }

    // 3. 更新状态为排队
    database.DB.Model(&task).Update("status", models.TaskQueued)

    // 4. 启动异步处理（复用现有逻辑）
    go s.processTask(task.ID)

    return nil
}
```

---

## 7. 前端设计

### 7.1 组件架构

```
CreativeGeneratorPage
├── StepIndicator (步骤指示器)
│   ├── Step 1: 产品输入
│   ├── Step 2: 文案选择
│   └── Step 3: 创意配置
│
├── Step1: ProductInput (商品输入表单)
│   ├── Input: 商品名称
│   └── Button: 生成文案
│
├── Step2: CopywritingSelection (文案选择)
│   ├── CTAOptions (CTA单选组)
│   │   ├── Radio: 选项1
│   │   └── Radio: 选项2
│   ├── EditCTA (编辑CTA)
│   ├── SPOptions (卖点多选组)
│   │   ├── Checkbox: 选项1
│   │   ├── Checkbox: 选项2
│   │   └── Checkbox: 选项3
│   ├── EditSPs (编辑卖点)
│   └── Buttons: 返回 | 下一步
│
└── Step3: CreativeConfig (创意配置)
    ├── Input: 产品图片URL
    ├── Select: 创意风格
    ├── Input: 变体数量
    └── Buttons: 返回 | 生成创意
```

### 7.2 状态管理

```typescript
// 工作流步骤枚举
enum WorkflowStep {
  PRODUCT_INPUT = 1,
  COPYWRITING_SELECTION = 2,
  CREATIVE_CONFIG = 3,
}

// 组件状态
const [currentStep, setCurrentStep] = useState<WorkflowStep>(WorkflowStep.PRODUCT_INPUT);

// 步骤1数据
const [productName, setProductName] = useState('');
const [generatingCopywriting, setGeneratingCopywriting] = useState(false);

// 步骤2数据
const [copywritingData, setCopywritingData] = useState<CopywritingData | null>(null);
const [selectedCTAIndex, setSelectedCTAIndex] = useState<number>(0);
const [selectedSPIndexes, setSelectedSPIndexes] = useState<number[]>([]);
const [editedCTA, setEditedCTA] = useState('');
const [editedSPs, setEditedSPs] = useState<string[]>([]);

// 步骤3数据
const [productImageURL, setProductImageURL] = useState('');
const [style, setStyle] = useState('');
const [numVariants, setNumVariants] = useState(3);
const [submitting, setSubmitting] = useState(false);
```

### 7.3 UI交互流程

**步骤1 → 步骤2 转换**:
```typescript
const handleGenerateCopywriting = async (e: React.FormEvent) => {
  e.preventDefault();
  setGeneratingCopywriting(true);

  try {
    const response = await creativeAPI.generateCopywriting({
      product_name: productName
    });

    if (response.code === 0 && response.data) {
      setCopywritingData(response.data);
      setCurrentStep(WorkflowStep.COPYWRITING_SELECTION);
      // 初始化选择状态
      setSelectedCTAIndex(0);
      setSelectedSPIndexes([0]);
    } else {
      alert('生成文案失败: ' + response.message);
    }
  } catch (err) {
    alert('请求失败: ' + err.message);
  } finally {
    setGeneratingCopywriting(false);
  }
};
```

**步骤2 → 步骤3 转换**:
```typescript
const handleProceedToCreativeConfig = () => {
  if (selectedSPIndexes.length === 0) {
    alert('请至少选择一个卖点');
    return;
  }
  setCurrentStep(WorkflowStep.CREATIVE_CONFIG);
};
```

**步骤3 提交**:
```typescript
const handleSubmitCreative = async (e: React.FormEvent) => {
  e.preventDefault();
  setSubmitting(true);

  try {
    // 先确认文案
    const confirmResponse = await creativeAPI.confirmCopywriting({
      task_id: copywritingData.task_id,
      selected_cta_index: selectedCTAIndex,
      selected_sp_indexes: selectedSPIndexes,
      edited_cta: editedCTA || undefined,
      edited_sps: editedSPs.length > 0 ? editedSPs : undefined,
      product_image_url: productImageURL,
      style: style,
      num_variants: numVariants,
      formats: ['1:1'],
    });

    if (confirmResponse.code !== 0) {
      alert('确认文案失败');
      return;
    }

    // 再启动创意生成
    const startResponse = await creativeAPI.startCreative({
      task_id: copywritingData.task_id
    });

    if (startResponse.code === 0) {
      alert(`创意生成已开始！任务ID: ${copywritingData.task_id}`);
      navigate('/tasks');
    }
  } catch (err) {
    alert('请求失败: ' + err.message);
  } finally {
    setSubmitting(false);
  }
};
```

### 7.4 样式设计

**步骤指示器**:
```css
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2rem;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0.4;
}

.step.active {
  opacity: 1;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.step.active .step-number {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}
```

**选项卡**:
```css
.radio-option, .checkbox-option {
  padding: 1rem;
  background: white;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.radio-option:hover, .checkbox-option:hover {
  border-color: #667eea;
  background: #f5f7ff;
}

.radio-option input:checked + .radio-label {
  color: #667eea;
  font-weight: 600;
}
```

---

## 8. 实施步骤

### 8.1 后端实施（按顺序）

1. **数据模型** (`internal/models/creative.go`)
   - 添加6个新字段到CreativeTask结构体
   - 验证GORM tag正确性

2. **QwenClient** (`internal/services/qwen_client.go` - 新建)
   - 实现HTTP客户端和API调用
   - 实现Prompt构建
   - 实现响应解析和验证
   - 添加单元测试

3. **CopywritingService** (`internal/services/copywriting_service.go` - 新建)
   - 实现GenerateCopywriting方法
   - 实现ConfirmCopywriting方法
   - 添加业务逻辑验证

4. **CreativeService扩展** (`internal/services/creative_service.go`)
   - 添加StartCreativeGeneration方法

5. **DTOs** (`internal/handlers/dto.go`)
   - 添加3个新DTO结构体
   - 添加JSON tag和validation tag

6. **Handler** (`internal/handlers/creative_handler.go`)
   - 添加CopywritingService依赖
   - 实现3个新handler方法
   - 添加错误处理

7. **路由** (`main.go`)
   - 注册3个新API端点

8. **测试**
   - 使用Postman/curl测试每个端点
   - 验证数据库数据正确性

### 8.2 前端实施（按顺序）

9. **类型定义** (`web/src/types/index.ts`)
   - 添加3个新接口

10. **API方法** (`web/src/services/api.ts`)
    - 添加3个新API方法

11. **页面重构** (`web/src/pages/CreativeGeneratorPage.tsx`)
    - 实现三步骤工作流
    - 实现状态管理
    - 实现表单验证
    - 实现API调用

12. **样式** (`web/src/App.css`)
    - 添加步骤指示器样式
    - 添加选项卡样式

13. **测试**
    - 端到端流程测试
    - 边界情况测试
    - 错误处理测试

### 8.3 集成测试

14. **完整流程验证**
    - 输入商品名称 → 生成文案 → 选择编辑 → 生成图片 → 查看结果
    - 验证所有候选文案正确保存
    - 验证最终选定文案正确使用

---

## 9. 技术细节

### 9.1 并发处理
- 文案生成：同步调用（用户等待）
- 图片生成：异步处理（goroutine）
- 数据库写入：事务保护

### 9.2 错误处理策略
| 场景 | 策略 |
|------|------|
| LLM API超时 | 返回503错误，提示用户重试 |
| LLM返回格式错误 | 记录原始响应，返回解析错误 |
| 数据库写入失败 | 回滚事务，返回500错误 |
| 前端网络错误 | Toast提示，保留用户输入 |

### 9.3 性能优化
- LLM调用超时：30秒
- 前端防抖：避免重复点击
- 数据库索引：task_id, status, copywriting_generated
- 缓存策略：暂不缓存（后续可考虑Redis）

### 9.4 安全考虑
- API认证：JWT token（复用现有认证）
- 输入验证：商品名称长度限制（255字符）
- SQL注入防护：GORM参数化查询
- XSS防护：React自动转义

---

## 10. 部署清单

### 10.1 环境配置
```bash
# .env文件
TONGYI_API_KEY=sk-xxxxxxxxxxxxx
TONGYI_LLM_MODEL=qwen-turbo
TONGYI_IMAGE_MODEL=wanx-v1
```

### 10.2 数据库迁移
```bash
# GORM会自动迁移，无需手动操作
# 但建议先在测试环境验证
go run main.go  # 自动执行AutoMigrate
```

### 10.3 前端构建
```bash
cd web
npm install
npm run build
```

### 10.4 验收标准

**后端**:
- [ ] 千问LLM API调用成功返回2+3文案
- [ ] 候选文案正确存储到creative_tasks表
- [ ] 文案确认后能触发图片生成
- [ ] 旧流程仍可用（向后兼容）

**前端**:
- [ ] 三步骤流程顺畅切换
- [ ] 用户可选择和编辑文案
- [ ] 生成后跳转到任务页
- [ ] 所有表单验证正常

**集成**:
- [ ] 完整流程可顺利执行
- [ ] TasksPage显示文案候选信息
- [ ] 错误场景有友好提示

---

## 11. 风险与应对

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| LLM API限流 | 用户无法生成文案 | 添加重试机制，提示用户稍后重试 |
| LLM返回格式异常 | 解析失败 | 增强JSON提取逻辑，记录原始响应 |
| 前端状态丢失 | 用户需重新操作 | 使用localStorage临时保存 |
| 数据库字段冲突 | 迁移失败 | 使用GORM的nullable字段 |

---

## 12. 未来优化方向

1. **智能推荐**: 基于历史数据，推荐更高转化率的文案
2. **A/B测试**: 支持多套文案组合生成，自动测试最优方案
3. **多语言支持**: 扩展支持英文、日文等语言文案生成
4. **文案模板**: 提供行业文案模板供用户选择
5. **批量生成**: 支持批量导入商品列表，批量生成文案
6. **WebSocket实时反馈**: 文案生成过程实时推送进度

---

## 附录

### A. 文件清单

**后端文件** (7个):
1. `internal/models/creative.go` - 修改
2. `internal/services/qwen_client.go` - 新建
3. `internal/services/copywriting_service.go` - 新建
4. `internal/services/creative_service.go` - 修改
5. `internal/handlers/dto.go` - 修改
6. `internal/handlers/creative_handler.go` - 修改
7. `main.go` - 修改

**前端文件** (4个):
8. `web/src/types/index.ts` - 修改
9. `web/src/services/api.ts` - 修改
10. `web/src/pages/CreativeGeneratorPage.tsx` - 重构
11. `web/src/App.css` - 修改

### B. 工作量预估
- 后端开发：4-5小时
- 前端开发：3-4小时
- 测试调试：2小时
- **总计**：9-11小时

### C. 参考资料
- [阿里云DashScope API文档](https://help.aliyun.com/zh/dashscope/)
- [千问LLM模型说明](https://help.aliyun.com/zh/dashscope/developer-reference/model-introduction)
- [GORM官方文档](https://gorm.io/zh_CN/docs/)
- [React Hooks最佳实践](https://react.dev/reference/react)

---

**设计文档完成，准备实施。**

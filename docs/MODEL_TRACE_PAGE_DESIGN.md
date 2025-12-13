# 模型调用链路展示页面设计方案

## 目标
- 在前端提供一个页面，清晰呈现单次模型调用的链路与中间过程，方便排障与性能分析。
- 支持列表 + 详情视图：列表查看调用概要，详情查看步骤拆解、耗时、输入输出摘要、错误栈。

## 主要需求
1) 列表视图
   - 字段：trace_id / request_id、模型名称/版本、业务来源（如实验ID/任务ID）、状态（success/failed/running）、总耗时、开始时间。
   - 过滤/搜索：按 trace_id、状态、时间范围、模型名称。
   - 分页：默认 page=1，page_size=20，可调。

2) 详情视图
   - 基本信息：trace_id、模型/版本、入口参数摘要（截断）、调用耗时、状态、开始/结束时间、关联对象（实验ID/任务ID/用户ID等）。
   - 步骤列表（timeline）：每个步骤包含
     - step_name（如 preprocess、rerank、llm_generate 等）
     - component/model 名称与版本
     - duration_ms，start_ts / end_ts
     - 输入摘要 / 输出摘要（长度受限，支持展开）
     - 日志或错误（异常时）
   - 元数据：节点机器/区域、重试次数、缓存命中信息（可选）。

3) 性能与观测
   - 需要后端落库或实时查询 trace 数据；若已有 APM/Tracing 系统（如 OpenTelemetry），前端可调用后端封装的接口转发/聚合。

## 后端接口设计（建议）
### 数据模型（建议表）
- `model_traces`
  - id (auto)
  - trace_id (string, index)
  - model_name (string)
  - model_version (string)
  - status (enum: running/success/failed)
  - duration_ms (int)
  - start_at (datetime)
  - end_at (datetime)
  - request_meta (json) // 包含来源：experiment_id / task_id / user_id 等
  - input_preview (text) // 截断后的输入
  - output_preview (text) // 截断后的输出
  - error_message (text)

- `model_trace_steps`
  - id (auto)
  - trace_id (string, index)
  - step_name (string)
  - component (string) // 模块/模型名称
  - status (enum)
  - duration_ms (int)
  - start_at (datetime)
  - end_at (datetime)
  - input_preview (text)
  - output_preview (text)
  - error_message (text)
  - extra (json) // 机器、区域、缓存命中、重试等

### API（v1 示例）
- `GET /api/v1/model_traces`  
  params: page, page_size, status, model_name, start_at, end_at, trace_id  
  resp: list (基础字段) + total/page/page_size
- `GET /api/v1/model_traces/:trace_id`  
  resp: 基本信息 + steps 数组

> 若已有 OpenTelemetry/Jaeger：后端提供以上两个接口，内部去 trace 系统查询并做字段映射。

## 前端页面方案
路由建议：`/traces`

1) 列表区
   - 顶部过滤：trace_id 输入、状态下拉、模型下拉、时间范围。
   - 表格列：trace_id（可复制）、模型/版本、状态徽章、耗时、开始时间、来源（如 experiment_id/ task_id）。
   - 行操作：查看详情。

2) 详情区
   - 基本信息卡：trace_id 复制、状态徽章、耗时、时间、模型、来源。
   - 步骤时间线（timeline）：每步显示名称、component、耗时、状态、输入/输出摘要、错误提示，支持展开查看全量。
   - 错误显示：当 status=failed 时突出 error_message。

3) 交互与体验
   - 轻量紧凑布局，固定侧边栏复用现有 Sidebar。
   - 状态颜色：running（蓝）、success（绿）、failed（红）。
   - 复制：trace_id、来源 ID 可一键复制。

## 实现优先级
MVP（推荐先做）：
- 后端：实现两个接口（列表 + 详情），数据可来自落库或 trace 系统映射。
- 前端：新增 `/traces` 页面：过滤 + 表格 + 详情抽屉/区块。

后续增强：
- 时间线可视化（节点耗时条形图）。
- 输入/输出全文查看与下载。
- 导出 CSV/JSON。
- 与实验/任务页打通：从实验详情跳转到关联的 trace 列表（按 experiment_id 过滤）。

## 依赖与注意
- 需要后端能返回 step 级数据；若没有，需先落库或接 OpenTelemetry/Jaeger。
- 输入输出截断处理，避免过大响应；错误信息需注意脱敏。
- 时间、耗时统一时区/格式；状态枚举前后端对齐。

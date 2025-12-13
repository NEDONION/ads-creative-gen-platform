# 广告创意生成平台 - 前端

基于 React + TypeScript + Vite 构建的现代化前端应用。

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **React Router** - 路由管理
- **Axios** - HTTP 客户端

## 功能特性

- ✅ **仪表盘** - 查看平台统计信息
- ✅ **任务管理** - 查看和管理所有创意生成任务(完全对接后端API,无模拟数据)
- ✅ **素材管理** - 浏览和管理生成的广告素材
- ✅ **创意生成** - 创建新的广告创意生成任务

## 开发

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

前端将运行在 `http://localhost:3000`

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## API 配置

前端通过 Vite 代理连接后端 API:

- API 基础路径: `/api/v1`
- 后端服务: `http://localhost:4000`

代理配置在 `vite.config.js` 中。

## 项目结构

```
web/
├── src/
│   ├── components/     # 可复用组件
│   │   ├── Layout.tsx
│   │   ├── Sidebar.tsx
│   │   └── Header.tsx
│   ├── pages/          # 页面组件
│   │   ├── DashboardPage.tsx
│   │   ├── TasksPage.tsx       # 任务管理(无模拟数据)
│   │   ├── AssetsPage.tsx
│   │   └── CreativeGeneratorPage.tsx
│   ├── services/       # API 服务
│   │   └── api.ts
│   ├── types/          # TypeScript 类型定义
│   │   └── index.ts
│   ├── App.tsx         # 主应用组件
│   ├── App.css         # 全局样式
│   └── main.tsx        # 应用入口
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.js
```

## 重要说明

### 任务管理页面

任务管理页面已完全接入后端 API,所有数据都从以下真实接口获取:

- **GET /api/v1/creative/tasks** - 获取任务列表
- **GET /api/v1/creative/task/:id** - 获取任务详情

页面中**没有任何模拟数据**,所有展示的信息都来自后端数据库。

## 后端 API 接口

- `GET /health` - 健康检查
- `GET /api/v1/ping` - Ping 测试
- `POST /api/v1/creative/generate` - 创建创意生成任务
- `GET /api/v1/creative/task/:id` - 查询任务详情
- `GET /api/v1/creative/tasks` - 获取任务列表 (新增)
- `GET /api/v1/creative/assets` - 获取素材列表

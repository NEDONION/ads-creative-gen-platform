# ==========================================
# 多阶段构建 Dockerfile
# 用于 Railway 和通用云平台部署
# ==========================================

# ==========================================
# 阶段 1: 构建前端
# ==========================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package*.json ./

# 安装依赖（使用 npm ci 更快更可靠）
RUN npm ci

# 复制前端源代码
COPY web/ ./

# 构建前端
RUN npm run build

# ==========================================
# 阶段 2: 构建后端
# ==========================================
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源代码
COPY . .

# 从前一阶段复制构建好的前端文件
COPY --from=frontend-builder /app/web/dist ./web/dist

# 构建 Go 应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ==========================================
# 阶段 3: 运行时环境
# ==========================================
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/main .

# 从构建阶段复制前端静态文件
COPY --from=frontend-builder /app/web/dist ./web/dist

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# 暴露端口（Railway 会自动注入 PORT 环境变量）
EXPOSE 4000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-4000}/health || exit 1

# 启动应用
CMD ["./main"]

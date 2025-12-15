# ==========================================
# 阶段 1: 构建前端
# ==========================================
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package*.json ./

# 安装前端依赖
RUN npm install

# 复制前端源代码
COPY web/ ./

# 构建前端
RUN npm run build

# ==========================================
# 阶段 2: 构建 Go 后端
# ==========================================
FROM golang:1.21-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache git

WORKDIR /app

# 复制 Go 依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /app/web/dist ./web/dist

# 构建 Go 应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# ==========================================
# 阶段 3: 最终运行时镜像
# ==========================================
FROM alpine:latest

# 安装 CA 证书（用于 HTTPS 请求）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件和前端静态文件
COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/web/dist ./web/dist

# 暴露端口（fly.io 会自动映射）
EXPOSE 8080

# 运行应用
CMD ["./main"]

# 部署指南

## 🚀 部署方式

本项目支持以下部署方式:

1. **本地开发部署** - 直接运行
2. **Docker 部署** - 推荐用于开发和测试
3. **生产环境部署** - 使用二进制文件

---

## 📦 方式一: 本地开发部署

### 前置要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+

### 步骤

1. **启动后端**
```bash
# 配置环境
cp config/config.ini.example config/config.ini
vim config/config.ini

# 初始化数据库
go run cmd/migrate/main.go -action reset

# 启动服务
./scripts/start.sh
```

2. **启动前端**
```bash
cd web
npm install
npm run dev
```

3. **访问应用**
- 前端: http://localhost:3000
- 后端: http://localhost:4000

---

## 🐳 方式二: Docker 部署 (推荐)

### 包含服务

| 服务 | 端口 | 管理界面 |
|------|------|---------|
| MySQL 8.0 | 3306 | phpMyAdmin :8081 |
| Redis 7 | 6379 | - |
| MinIO | 9000 | :9001 |
| phpMyAdmin | 8081 | http://localhost:8081 |

### 启动步骤

```bash
# 启动所有服务
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f mysql
```

### 验证服务

**MySQL**:
```bash
# phpMyAdmin
浏览器访问: http://localhost:8081
用户名: root
密码: root

# 命令行
docker exec -it ads_creative_mysql mysql -uroot -proot ads_creative_platform
```

**Redis**:
```bash
docker exec -it ads_creative_redis redis-cli ping
# 输出: PONG
```

**MinIO**:
```bash
浏览器访问: http://localhost:9001
用户名: minioadmin
密码: minioadmin
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

---

## 🏭 方式三: 生产环境部署

### 1. 构建二进制文件

```bash
# 构建后端
go build -o bin/server main.go

# 构建前端
cd web
npm run build
```

### 2. 配置环境

```bash
# 复制配置文件
cp config/config.ini.example config/config.ini

# 编辑配置
vim config/config.ini
```

配置示例:
```ini
[app]
AppMode = release
HttpPort = :4000

[mysql]
DbHost = your_mysql_host
DbPort = 3306
DbUser = your_user
DbPassWord = your_password
DbName = ads_creative_platform

[tongyi]
ApiKey = your_api_key

[qiniu]
AccessKey = your_access_key
SecretKey = your_secret_key
Bucket = your_bucket
Domain = your_domain
```

### 3. 初始化数据库

```bash
# 创建数据库
mysql -h your_host -u root -p -e "CREATE DATABASE ads_creative_platform CHARACTER SET utf8mb4;"

# 运行迁移
./bin/server migrate
```

### 4. 使用 systemd 管理服务

创建 `/etc/systemd/system/ads-creative.service`:

```ini
[Unit]
Description=Ads Creative Platform
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/ads-creative-platform
ExecStart=/opt/ads-creative-platform/bin/server
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务:
```bash
sudo systemctl daemon-reload
sudo systemctl enable ads-creative
sudo systemctl start ads-creative
sudo systemctl status ads-creative
```

### 5. 配置 Nginx

创建 `/etc/nginx/sites-available/ads-creative`:

```nginx
server {
    listen 80;
    server_name your_domain.com;

    # 前端
    location / {
        root /opt/ads-creative-platform/web/dist;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API
    location /api {
        proxy_pass http://localhost:4000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # 健康检查
    location /health {
        proxy_pass http://localhost:4000;
    }
}
```

启用配置:
```bash
sudo ln -s /etc/nginx/sites-available/ads-creative /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 🔧 使用脚本管理

项目提供了便捷的管理脚本（位于 `scripts/` 目录）:

### 启动服务

```bash
./scripts/start.sh
```

功能:
- 检查 MySQL 连接
- 自动迁移数据库
- 启动后端服务

### 停止服务

```bash
./scripts/stop.sh
```

功能:
- 优雅关闭服务
- 杀死残留进程

### 查看状态

```bash
./scripts/status.sh
```

功能:
- 显示服务运行状态
- 显示端口占用情况
- 显示最近的日志

---

## 📊 监控与日志

### 查看日志

```bash
# 后端日志
tail -f logs/app.log

# Nginx 日志
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

### 监控指标

建议监控:
- 服务健康: `GET /health`
- 数据库连接数
- API 响应时间
- 错误率

---

## 🔐 安全建议

### 1. 修改默认密码

```bash
# 修改数据库 root 密码
mysql -u root -p
ALTER USER 'root'@'localhost' IDENTIFIED BY 'new_strong_password';

# 修改应用管理员密码
# 登录后在用户设置中修改
```

### 2. 配置防火墙

```bash
# 只允许必要的端口
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

### 3. 使用 HTTPS

```bash
# 使用 Let's Encrypt 获取免费证书
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d your_domain.com
```

### 4. 定期备份

```bash
# 数据库备份脚本
#!/bin/bash
BACKUP_DIR="/var/backups/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
mysqldump -u root -p ads_creative_platform > $BACKUP_DIR/backup_$DATE.sql
find $BACKUP_DIR -name "backup_*.sql" -mtime +7 -delete
```

设置定时任务:
```bash
crontab -e
# 每天凌晨 2 点备份
0 2 * * * /path/to/backup.sh
```

---

## 🆘 故障排查

### 服务无法启动

1. 检查端口是否被占用:
```bash
lsof -i:4000
```

2. 查看日志:
```bash
tail -f logs/app.log
```

3. 检查配置文件:
```bash
cat config/config.ini
```

### 数据库连接失败

1. 检查 MySQL 是否运行:
```bash
systemctl status mysql
```

2. 测试连接:
```bash
mysql -h 127.0.0.1 -u root -p
```

3. 检查配置文件中的连接信息

### 前端无法访问

1. 检查 Nginx 状态:
```bash
systemctl status nginx
```

2. 检查 Nginx 配置:
```bash
nginx -t
```

3. 查看 Nginx 错误日志:
```bash
tail -f /var/log/nginx/error.log
```

---

## 📈 性能优化

### 1. 数据库优化

- 添加适当的索引
- 定期清理过期数据
- 使用连接池

### 2. 缓存优化

- 启用 Redis 缓存
- 缓存频繁访问的数据
- 设置合理的过期时间

### 3. 静态资源优化

- 使用 CDN 分发静态资源
- 开启 Gzip 压缩
- 配置浏览器缓存

Nginx 配置示例:
```nginx
# Gzip 压缩
gzip on;
gzip_types text/css application/javascript application/json;

# 静态资源缓存
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

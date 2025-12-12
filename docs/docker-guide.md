# Docker 部署指南

## 🐳 为什么使用 Docker？

### 优势

1. **环境一致性**：所有开发者使用相同的数据库、缓存、消息队列版本
2. **一键启动**：无需手动安装 MySQL、Redis、RabbitMQ 等
3. **隔离性**：不污染宿主机环境
4. **易于重建**：出问题可以快速删除并重建
5. **接近生产**：开发环境与生产环境一致

---

## 📦 包含的服务

| 服务 | 端口 | 用途 | 管理界面 |
|------|------|------|---------|
| **MySQL 8.0** | 3306 | 主数据库 | phpMyAdmin :8081 |
| **Redis 7** | 6379 | 缓存、任务队列 | - |
| **RabbitMQ 3** | 5672 | 消息队列 | :15672 (guest/guest) |
| **MinIO** | 9000 | 对象存储 | :9001 (minioadmin/minioadmin) |
| **phpMyAdmin** | 8081 | MySQL 管理 | http://localhost:8081 |

---

## 🚀 快速开始

### 1. 安装 Docker

**Mac**:
```bash
# 下载 Docker Desktop for Mac
https://www.docker.com/products/docker-desktop
```

**Windows**:
```bash
# 下载 Docker Desktop for Windows
https://www.docker.com/products/docker-desktop
```

**Linux**:
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 docker-compose
sudo apt-get install docker-compose-plugin
```

### 2. 启动所有服务

```bash
# 在项目根目录下执行
docker-compose up -d
```

第一次运行会下载镜像，需要几分钟。

### 3. 查看服务状态

```bash
docker-compose ps
```

你应该看到：

```
NAME                      STATUS              PORTS
ads_creative_mysql        Up (healthy)        0.0.0.0:3306->3306/tcp
ads_creative_redis        Up (healthy)        0.0.0.0:6379->6379/tcp
ads_creative_rabbitmq     Up (healthy)        0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp
ads_creative_minio        Up (healthy)        0.0.0.0:9000-9001->9000-9001/tcp
ads_creative_phpmyadmin   Up                  0.0.0.0:8081->80/tcp
```

### 4. 验证服务

#### MySQL
```bash
# 方式一：使用 phpMyAdmin
打开浏览器访问: http://localhost:8081
用户名: root
密码: root

# 方式二：命令行连接
docker exec -it ads_creative_mysql mysql -uroot -proot ads_creative_platform

# 方式三：本地客户端连接
mysql -h 127.0.0.1 -P 3306 -uroot -proot ads_creative_platform
```

#### Redis
```bash
docker exec -it ads_creative_redis redis-cli ping
# 输出: PONG
```

#### RabbitMQ
```bash
# 打开管理界面
http://localhost:15672
用户名: guest
密码: guest
```

#### MinIO
```bash
# 打开控制台
http://localhost:9001
用户名: minioadmin
密码: minioadmin
```

---

## 🔧 常用命令

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 只启动 MySQL
docker-compose up -d mysql

# 启动 MySQL + Redis
docker-compose up -d mysql redis
```

### 停止服务

```bash
# 停止所有服务
docker-compose stop

# 停止并删除容器（数据保留）
docker-compose down

# 停止并删除容器和数据卷（⚠️ 会删除所有数据）
docker-compose down -v
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs

# 查看 MySQL 日志
docker-compose logs mysql

# 实时跟踪日志
docker-compose logs -f mysql

# 查看最近 100 行日志
docker-compose logs --tail=100 mysql
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启 MySQL
docker-compose restart mysql
```

### 进入容器

```bash
# 进入 MySQL 容器
docker exec -it ads_creative_mysql bash

# 进入 Redis 容器
docker exec -it ads_creative_redis sh

# 直接执行 MySQL 命令
docker exec -it ads_creative_mysql mysql -uroot -proot
```

---

## 🗂️ 数据持久化

所有数据都会持久化到 Docker 卷中，即使删除容器也不会丢失数据。

### 查看数据卷

```bash
docker volume ls | grep ads_creative
```

你会看到：
```
ads-creative-gen-platform_mysql_data
ads-creative-gen-platform_redis_data
ads-creative-gen-platform_rabbitmq_data
ads-creative-gen-platform_minio_data
```

### 备份数据

```bash
# 备份 MySQL
docker exec ads_creative_mysql mysqldump -uroot -proot ads_creative_platform > backup.sql

# 恢复 MySQL
docker exec -i ads_creative_mysql mysql -uroot -proot ads_creative_platform < backup.sql
```

### 清除所有数据（⚠️ 危险操作）

```bash
# 停止并删除所有容器和数据
docker-compose down -v

# 重新启动
docker-compose up -d
```

---

## ⚙️ 配置修改

### 修改 MySQL 端口

编辑 `docker-compose.yml`:

```yaml
mysql:
  ports:
    - "3307:3306"  # 改为 3307
```

然后更新 `config/config.ini`:

```ini
DbPort = 3307
```

### 修改 MySQL 密码

编辑 `docker-compose.yml`:

```yaml
mysql:
  environment:
    MYSQL_ROOT_PASSWORD: your_new_password
```

然后更新 `config/config.ini`:

```ini
DbPassWord = your_new_password
```

**重要**：修改配置后需要重建容器：

```bash
docker-compose down -v
docker-compose up -d
```

---

## 🔍 故障排查

### MySQL 无法启动

```bash
# 查看 MySQL 日志
docker-compose logs mysql

# 常见问题：端口被占用
# 方案1：修改 docker-compose.yml 中的端口
# 方案2：停止本地 MySQL 服务
```

### 无法连接 MySQL

```bash
# 1. 检查容器是否健康
docker-compose ps

# 2. 检查端口是否开放
netstat -an | grep 3306

# 3. 测试连接
docker exec -it ads_creative_mysql mysql -uroot -proot -e "SELECT 1"

# 4. 检查防火墙
# Mac: 一般不需要
# Windows: 检查 Windows Defender 防火墙
# Linux: sudo ufw allow 3306
```

### 数据库连接慢

```bash
# 查看容器资源使用
docker stats

# 重启 Docker
# Mac: Docker Desktop -> Restart
# Linux: sudo systemctl restart docker
```

---

## 📊 资源限制

如果你的机器配置有限，可以限制容器资源使用：

编辑 `docker-compose.yml`，添加：

```yaml
mysql:
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 1G
      reservations:
        memory: 512M
```

---

## 🎯 最佳实践

### 1. 仅启动需要的服务

**Phase 1 (MVP)**：只需要 MySQL
```bash
docker-compose up -d mysql phpmyadmin
```

**Phase 4 (生产化)**：启动全部
```bash
docker-compose up -d
```

### 2. 定期备份

```bash
# 每天备份一次
docker exec ads_creative_mysql mysqldump -uroot -proot ads_creative_platform \
  > backup_$(date +%Y%m%d).sql
```

### 3. 监控资源使用

```bash
# 查看容器资源占用
docker stats --no-stream

# 查看磁盘使用
docker system df
```

### 4. 清理未使用的资源

```bash
# 清理未使用的镜像、容器、网络
docker system prune -a

# 清理未使用的数据卷
docker volume prune
```

---

## 📝 配置文件对照

### config/config.ini

使用 Docker 时的配置：

```ini
[mysql]
Db = mysql
DbHost = 127.0.0.1        # 使用本地回环地址
DbPort = 3306              # Docker 映射的端口
DbUser = root
DbPassWord = root
DbName = ads_creative_platform
Charset = utf8mb4
```

### .env

```env
# 如果项目也在 Docker 中运行，使用容器名
# DB_HOST=mysql

# 如果项目在宿主机运行，使用 localhost
DB_HOST=127.0.0.1
```

---

## 🚢 生产环境部署

### 1. 使用 docker-compose.prod.yml

创建生产环境配置：

```yaml
version: '3.8'
services:
  mysql:
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}  # 从环境变量读取
    volumes:
      - /data/mysql:/var/lib/mysql  # 使用宿主机路径
```

### 2. 使用外部数据库

生产环境建议使用云数据库（阿里云 RDS、腾讯云 TencentDB）：

- 更高可用性
- 自动备份
- 自动监控
- 更好的性能

---

## 🆘 常见问题

### Q: Docker Desktop 启动慢？

A:
1. 检查 Docker Desktop 资源设置（Settings -> Resources）
2. 减少启动时自动启动的容器
3. 升级到最新版本

### Q: 端口冲突？

A: 修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "3307:3306"  # 主机端口:容器端口
```

### Q: 容器一直重启？

A:
```bash
docker-compose logs mysql  # 查看错误日志
docker-compose down -v     # 删除数据卷重新开始
```

---

## 📚 相关资源

- [Docker 官方文档](https://docs.docker.com/)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [MySQL Docker Hub](https://hub.docker.com/_/mysql)
- [Redis Docker Hub](https://hub.docker.com/_/redis)

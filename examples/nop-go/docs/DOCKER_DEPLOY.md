# nop-go Docker 部署指南

## 快速开始

### Windows 用户

```cmd
# 完整部署（构建 + 启动）
scripts\deploy.bat all

# 查看服务状态
scripts\deploy.bat status

# 停止服务
scripts\deploy.bat down
```

### Linux/Mac 用户

```bash
# 添加执行权限
chmod +x scripts/deploy.sh

# 完整部署（构建 + 启动）
./scripts/deploy.sh all

# 查看服务状态
./scripts/deploy.sh status

# 停止服务
./scripts/deploy.sh down
```

## 可用命令

| 命令 | 说明 |
|------|------|
| `all` | 完整部署（构建镜像 + 启动服务） |
| `build` | 仅构建服务镜像 |
| `up` | 启动所有服务 |
| `down` | 停止所有服务 |
| `restart` | 重启所有服务 |
| `status` | 查看服务状态 |
| `logs [服务名]` | 查看日志 |
| `clean` | 清理所有容器和镜像 |
| `infra` | 仅启动基础设施（MySQL、Redis） |
| `swagger` | 仅启动 Swagger UI |

## 访问地址

### Swagger 文档

- **聚合 Swagger UI**: http://localhost:8080
- **网关 Swagger**: http://localhost:8000/swagger

### 各服务 Swagger

| 服务 | 端口 | Swagger 地址 |
|------|------|-------------|
| API 网关 | 8000 | http://localhost:8000/swagger |
| 后台管理 | 8001 | http://localhost:8001/swagger/index.html |
| 客户服务 | 8002 | http://localhost:8002/swagger/index.html |
| 商品目录 | 8003 | http://localhost:8003/swagger/index.html |
| 购物车 | 8004 | http://localhost:8004/swagger/index.html |
| 订单服务 | 8005 | http://localhost:8005/swagger/index.html |
| CMS | 8007 | http://localhost:8007/swagger/index.html |
| 库存 | 8008 | http://localhost:8008/swagger/index.html |
| 通知 | 8009 | http://localhost:8009/swagger/index.html |
| 支付 | 8010 | http://localhost:8010/swagger/index.html |
| 价格 | 8011 | http://localhost:8011/swagger/index.html |
| 物流 | 8012 | http://localhost:8012/swagger/index.html |
| 多店铺 | 8013 | http://localhost:8013/swagger/index.html |
| 媒体 | 8014 | http://localhost:8014/swagger/index.html |
| 多语言 | 8015 | http://localhost:8015/swagger/index.html |
| 联盟营销 | 8016 | http://localhost:8016/swagger/index.html |
| SEO | 8017 | http://localhost:8017/swagger/index.html |
| 导入导出 | 8018 | http://localhost:8018/swagger/index.html |
| 主题 | 8019 | http://localhost:8019/swagger/index.html |
| AI | 8020 | http://localhost:8020/swagger/index.html |

### 数据库

- **MySQL**: localhost:13306 (root / nop123456)
- **Redis**: localhost:16379

## 环境配置

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

主要配置项：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| MYSQL_ROOT_PASSWORD | MySQL root 密码 | nop123456 |
| JWT_SECRET | JWT 密钥 | your-jwt-secret-key |
| AI_API_KEY | AI 服务 API Key | (空) |
| ENABLE_SWAGGER | 是否启用 Swagger | true |

## 单独操作某个服务

```bash
# 仅启动某个服务
docker-compose up -d customer

# 查看某个服务日志
docker-compose logs -f customer

# 重启某个服务
docker-compose restart customer

# 停止某个服务
docker-compose stop customer
```

## 数据持久化

以下目录会持久化到 Docker Volume：

- `mysql_data`: MySQL 数据
- `redis_data`: Redis 数据
- `media_uploads`: 媒体文件上传目录
- `import_temp`: 导入导出临时目录
- `theme_themes`: 主题文件目录

## 故障排查

### 查看服务日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f customer
```

### 进入容器调试

```bash
# 进入服务容器
docker exec -it nop-customer sh

# 进入 MySQL 容器
docker exec -it nop-mysql mysql -uroot -p
```

### 重置环境

```bash
# 停止并删除所有容器、网络、卷
docker-compose down -v

# 重新完整部署
./scripts/deploy.sh all
```

## 生产环境注意事项

1. 修改 `.env` 中的密码和密钥
2. 关闭 Swagger: `ENABLE_SWAGGER=false`
3. 使用外部 MySQL 和 Redis
4. 配置 HTTPS/TLS
5. 设置资源限制（CPU/内存）
.PHONY: all build clean test start-all stop-all swagger-install swagger-gen swagger-gen-% swagger-aggregate swagger-clean swagger-help docker-build docker-up docker-down docker-logs docker-status docker-clean docker-all

# 服务列表
SERVICES := catalog-service customer-service inventory-service price-service \
            order-service payment-service shipping-service cart-service \
            cms-service notification-service gateway-service admin-service \
            store-service media-service localization-service affiliate-service \
            seo-service import-service theme-service ai-service

# 默认目标
all: build

# 编译所有服务
build:
	@echo "Building all services..."
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		if [ -d "services/$$svc" ]; then \
			cd services/$$svc && go build -o bin/$$svc ./cmd && cd ../..; \
		fi; \
	done
	@echo "Build complete!"

# 清理编译产物
clean:
	@echo "Cleaning..."
	@for svc in $(SERVICES); do \
		rm -f services/$$svc/bin/$$svc; \
	done
	@echo "Clean complete!"

# 运行测试
test:
	@echo "Running tests..."
	go test ./... -v

# 启动所有服务（需要先启动基础设施）
start-all:
	@echo "Starting all services..."
	@for svc in $(SERVICES); do \
		echo "Starting $$svc..."; \
		cd services/$$svc && go run ./cmd & cd ../..; \
	done

# 停止所有服务
stop-all:
	@echo "Stopping all services..."
	@killall main 2>/dev/null || true

# 启动单个服务
start-%:
	@echo "Starting $*..."
	cd services/$* && go run ./cmd

# 数据库初始化
init-db:
	@echo "Initializing database..."
	go run infra/init_db.go

# 生成 Proto
proto:
	@echo "Generating proto files..."
	protoc --go_out=. --go-grpc_out=. infra/proto/*.proto

# Docker 启动基础设施
infra-up:
	docker-compose up -d mysql redis

# Docker 停止基础设施
infra-down:
	docker-compose down

# ========== Swagger 文档生成 ==========

# 安装 swag 工具
swagger-install:
	@echo "安装 swag 工具..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ swag 安装完成"

# 为所有服务生成 swagger 文档
swagger-gen:
	@echo "=== 生成各服务 Swagger 文档 ==="
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ] && [ -f "services/$$svc/cmd/main.go" ]; then \
			echo "处理 $$svc..."; \
			cd services/$$svc && \
			mkdir -p docs && \
			swag init -g ./cmd/main.go -o ./docs --parseInternal --parseDependency 2>/dev/null || echo "  跳过 $$svc (缺少注解)"; \
			cd ../..; \
		fi; \
	done
	@echo "=== 完成 ==="

# 为单个服务生成 swagger
swagger-gen-%:
	@echo "生成 $* swagger..."
	@if [ -d "services/$*" ]; then \
		cd services/$* && mkdir -p docs && swag init -g ./cmd/main.go -o ./docs --parseInternal --parseDependency; \
	else \
		echo "服务 $* 不存在"; \
	fi

# 生成聚合文档
swagger-aggregate:
	@echo "生成聚合文档..."
	@mkdir -p docs/swagger
	@cd scripts/gen-swagger && go run main.go
	@echo "✓ 聚合文档生成完成: docs/swagger/index.html"

# 清理所有 swagger 文档
swagger-clean:
	@echo "清理 swagger 文档..."
	@for svc in $(SERVICES); do \
		rm -rf services/$$svc/docs; \
	done
	@rm -rf docs/swagger
	@echo "✓ 清理完成"

# Swagger 帮助信息
swagger-help:
	@echo "Swagger 文档生成命令:"
	@echo "  make swagger-install      安装 swag 工具"
	@echo "  make swagger-gen          为所有服务生成 swagger 文档"
	@echo "  make swagger-gen-<svc>    为单个服务生成 swagger 文档"
	@echo "  make swagger-aggregate    生成聚合文档"
	@echo "  make swagger-clean        清理所有 swagger 文档"
	@echo ""
	@echo "示例:"
	@echo "  make swagger-gen-customer-service"
	@echo ""
	@echo "访问方式:"
	@echo "  各服务: http://localhost:{port}/swagger/index.html"
	@echo "  网关聚合: http://localhost:8000/swagger"

# ========== Docker 部署命令 ==========

# Docker 构建镜像
docker-build:
	@echo "构建 Docker 镜像..."
	docker-compose build --parallel

# Docker 启动所有服务
docker-up:
	@echo "启动所有 Docker 服务..."
	docker-compose up -d
	@$(MAKE) docker-status

# Docker 停止所有服务
docker-down:
	@echo "停止所有 Docker 服务..."
	docker-compose down

# Docker 查看日志
docker-logs:
	docker-compose logs -f

# Docker 查看状态
docker-status:
	@echo "=== Docker 服务状态 ==="
	@docker-compose ps
	@echo ""
	@echo "访问地址:"
	@echo "  API 网关:     http://localhost:8000"
	@echo "  Swagger UI:   http://localhost:8080"
	@echo "  MySQL:        localhost:3306 (root / nop123456)"
	@echo "  Redis:        localhost:6379"

# Docker 清理
docker-clean:
	@echo "清理 Docker 容器和镜像..."
	docker-compose down -v --rmi local
	@echo "✓ 清理完成"

# Docker 完整部署
docker-all: swagger-gen docker-build
	@echo "启动基础设施..."
	docker-compose up -d mysql redis
	@sleep 15
	@echo "启动所有微服务..."
	docker-compose up -d
	@$(MAKE) docker-status

# 查看服务状态
status:
	@echo "Service status:"
	@curl -s http://localhost:8000/healthz 2>/dev/null && echo "gateway: OK" || echo "gateway: DOWN"
	@for port in 8001 8002 8003 8004 8005 8007 8008 8009 8010 8011 8012 8013 8014 8015 8016 8017 8018 8019 8020; do \
		result=$$(curl -s http://localhost:$$port/healthz 2>/dev/null); \
		if [ -n "$$result" ]; then \
			echo "Port $$port: OK"; \
		else \
			echo "Port $$port: DOWN"; \
		fi; \
	done
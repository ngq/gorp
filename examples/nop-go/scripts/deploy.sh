#!/bin/bash
# nop-go 微服务一键部署脚本
# 中文说明：构建并部署所有微服务到本地 Docker

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 配置
COMPOSE_FILE="${PROJECT_DIR}/docker-compose.yml"
IMAGE_PREFIX="nop-go"

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║          nop-go 微服务一键部署脚本                          ║"
echo "║          nopCommerce Go Version Docker Deploy              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# 显示帮助
show_help() {
    echo "使用方法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  build       构建所有服务镜像"
    echo "  up          启动所有服务"
    echo "  down        停止所有服务"
    echo "  restart     重启所有服务"
    echo "  logs        查看日志"
    echo "  status      查看服务状态"
    echo "  clean       清理所有容器和镜像"
    echo "  all         构建 + 启动（完整部署）"
    echo "  infra       仅启动基础设施（MySQL、Redis）"
    echo "  swagger     仅启动 Swagger UI"
    echo ""
    echo "示例:"
    echo "  $0 all      # 完整部署"
    echo "  $0 logs gateway  # 查看网关日志"
    echo "  $0 status   # 查看状态"
}

# 检查依赖
check_dependencies() {
    echo -e "${YELLOW}检查依赖...${NC}"

    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: Docker 未安装${NC}"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}错误: Docker Compose 未安装${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Docker 已安装${NC}"
    echo -e "${GREEN}✓ Docker Compose 已安装${NC}"
}

# 构建 Swagger 文档
build_swagger_docs() {
    echo -e "${YELLOW}生成 Swagger 文档...${NC}"

    cd "$PROJECT_DIR"

    # 检查 swag 是否安装
    if ! command -v swag &> /dev/null; then
        echo -e "${YELLOW}安装 swag 工具...${NC}"
        go install github.com/swaggo/swag/cmd/swag@latest
    fi

    # 生成各服务的 swagger 文档
    for service_dir in services/*-service; do
        if [ -d "$service_dir" ] && [ -f "$service_dir/cmd/main.go" ]; then
            service_name=$(basename "$service_dir")
            echo -e "  处理 ${service_name}..."
            cd "$service_dir"
            mkdir -p docs
            swag init -g ./cmd/main.go -o ./docs --parseInternal --parseDependency 2>/dev/null || true
            cd "$PROJECT_DIR"
        fi
    done

    echo -e "${GREEN}✓ Swagger 文档生成完成${NC}"
}

# 构建镜像
build_images() {
    echo -e "${YELLOW}构建服务镜像...${NC}"
    cd "$PROJECT_DIR"

    # 使用 docker-compose 构建
    docker-compose -f "$COMPOSE_FILE" build --parallel

    echo -e "${GREEN}✓ 镜像构建完成${NC}"
}

# 启动基础设施
start_infra() {
    echo -e "${YELLOW}启动基础设施 (MySQL, Redis)...${NC}"
    cd "$PROJECT_DIR"

    docker-compose -f "$COMPOSE_FILE" up -d mysql redis

    echo -e "${YELLOW}等待数据库就绪...${NC}"
    sleep 10

    # 等待 MySQL 就绪
    until docker-compose -f "$COMPOSE_FILE" exec -T mysql mysqladmin ping -h localhost --silent; do
        echo "  等待 MySQL 启动..."
        sleep 2
    done

    # 等待 Redis 就绪
    until docker-compose -f "$COMPOSE_FILE" exec -T redis redis-cli ping | grep -q PONG; do
        echo "  等待 Redis 启动..."
        sleep 2
    done

    echo -e "${GREEN}✓ 基础设施启动完成${NC}"
}

# 启动所有服务
start_services() {
    echo -e "${YELLOW}启动所有微服务...${NC}"
    cd "$PROJECT_DIR"

    docker-compose -f "$COMPOSE_FILE" up -d

    echo -e "${GREEN}✓ 服务启动完成${NC}"
    show_status
}

# 停止所有服务
stop_services() {
    echo -e "${YELLOW}停止所有服务...${NC}"
    cd "$PROJECT_DIR"

    docker-compose -f "$COMPOSE_FILE" down

    echo -e "${GREEN}✓ 服务已停止${NC}"
}

# 重启所有服务
restart_services() {
    stop_services
    start_services
}

# 查看日志
view_logs() {
    local service=$1
    cd "$PROJECT_DIR"

    if [ -n "$service" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f "$service"
    else
        docker-compose -f "$COMPOSE_FILE" logs -f
    fi
}

# 查看状态
show_status() {
    echo -e "${YELLOW}服务状态:${NC}"
    echo ""

    cd "$PROJECT_DIR"

    # 显示容器状态
    docker-compose -f "$COMPOSE_FILE" ps

    echo ""
    echo -e "${BLUE}访问地址:${NC}"
    echo "  API 网关:     http://localhost:8000"
    echo "  Swagger UI:   http://localhost:8080"
    echo "  后台管理:     http://localhost:8001/swagger/index.html"
    echo "  客户服务:     http://localhost:8002/swagger/index.html"
    echo "  商品目录:     http://localhost:8003/swagger/index.html"
    echo ""
    echo -e "${BLUE}数据库:${NC}"
    echo "  MySQL:        localhost:13306 (root / nop123456)"
    echo "  Redis:        localhost:16379"
}

# 清理
clean_all() {
    echo -e "${RED}清理所有容器和镜像...${NC}"
    cd "$PROJECT_DIR"

    docker-compose -f "$COMPOSE_FILE" down -v --rmi local

    echo -e "${GREEN}✓ 清理完成${NC}"
}

# 完整部署
deploy_all() {
    check_dependencies
    build_swagger_docs
    build_images
    start_infra
    start_services

    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                  部署成功完成！                            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}访问 Swagger 文档:${NC}"
    echo "  http://localhost:8080"
    echo ""
    echo -e "${BLUE}各服务 Swagger:${NC}"
    echo "  网关:    http://localhost:8000/swagger"
    echo "  客户:    http://localhost:8002/swagger/index.html"
    echo "  商品:    http://localhost:8003/swagger/index.html"
    echo "  订单:    http://localhost:8005/swagger/index.html"
}

# 主入口
case "$1" in
    build)
        check_dependencies
        build_swagger_docs
        build_images
        ;;
    up)
        start_services
        ;;
    down)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        view_logs "$2"
        ;;
    status)
        show_status
        ;;
    clean)
        clean_all
        ;;
    all)
        deploy_all
        ;;
    infra)
        check_dependencies
        start_infra
        ;;
    swagger)
        cd "$PROJECT_DIR"
        docker-compose -f "$COMPOSE_FILE" up -d swagger-ui
        echo -e "${GREEN}Swagger UI: http://localhost:8080${NC}"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        if [ -n "$1" ]; then
            echo -e "${RED}未知命令: $1${NC}"
        fi
        show_help
        exit 1
        ;;
esac
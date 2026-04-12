#!/bin/bash
# Swagger 文档生成脚本
# 中文说明：遍历所有微服务生成 swagger 文档

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NOP_GO_DIR="$(dirname "$SCRIPT_DIR")"

# 服务列表
SERVICES=(
    "admin-service"
    "affiliate-service"
    "ai-service"
    "cart-service"
    "catalog-service"
    "cms-service"
    "customer-service"
    "import-service"
    "inventory-service"
    "localization-service"
    "media-service"
    "notification-service"
    "order-service"
    "payment-service"
    "price-service"
    "seo-service"
    "shipping-service"
    "store-service"
    "theme-service"
)

echo "=== Swagger 文档生成 ==="
echo "工作目录: $NOP_GO_DIR"
echo ""

# 检查 swag 是否安装
if ! command -v swag &> /dev/null; then
    echo "错误: swag 未安装"
    echo "请运行: go install github.com/swaggo/swag/cmd/swag@latest"
    exit 1
fi

# 生成各服务的 swagger 文档
for service in "${SERVICES[@]}"; do
    service_dir="$NOP_GO_DIR/services/$service"
    if [ -d "$service_dir" ]; then
        echo "生成 $service swagger..."
        cd "$service_dir"

        # 创建 docs 目录
        mkdir -p docs

        # 生成 swagger 文档
        # -g 指定入口文件
        # -o 指定输出目录
        # --parseInternal 解析内部包
        # --parseDependency 解析依赖
        swag init \
            -g ./cmd/main.go \
            -o ./docs \
            --parseInternal \
            --parseDependency \
            --outputTypes json,yaml \
            2>/dev/null || echo "  警告: $service swagger 生成跳过（可能缺少注解）"
    fi
done

# 生成聚合文档
echo ""
echo "生成聚合 swagger 文档..."
AGGREGATED_DIR="$NOP_GO_DIR/docs/swagger"
mkdir -p "$AGGREGATED_DIR"

# 创建聚合索引
cat > "$AGGREGATED_DIR/index.json" << 'EOF'
{
  "swagger": "2.0",
  "info": {
    "title": "nop-go 微服务 API 文档",
    "description": "nopCommerce Go 版本微服务聚合 API 文档",
    "version": "1.0.0"
  },
  "apis": [
EOF

first=true
for service in "${SERVICES[@]}"; do
    swagger_file="$NOP_GO_DIR/services/$service/docs/swagger.json"
    if [ -f "$swagger_file" ]; then
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$AGGREGATED_DIR/index.json"
        fi
        cat >> "$AGGREGATED_DIR/index.json" << EOF
    {
      "name": "$service",
      "url": "/services/$service/swagger.json"
    }
EOF
    fi
done

cat >> "$AGGREGATED_DIR/index.json" << 'EOF'
  ]
}
EOF

echo ""
echo "=== 完成 ==="
echo "各服务 swagger: http://localhost:{port}/swagger/index.html"
echo "聚合文档索引: $AGGREGATED_DIR/index.json"
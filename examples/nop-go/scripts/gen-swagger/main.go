// Package main Swagger 文档生成工具
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 服务列表
var services = []string{
	"admin-service",
	"affiliate-service",
	"ai-service",
	"cart-service",
	"catalog-service",
	"cms-service",
	"customer-service",
	"import-service",
	"inventory-service",
	"localization-service",
	"media-service",
	"notification-service",
	"order-service",
	"payment-service",
	"price-service",
	"seo-service",
	"shipping-service",
	"store-service",
	"theme-service",
}

// ServiceSwaggerInfo 服务 swagger 信息
type ServiceSwaggerInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Port        int    `json:"port"`
}

// AggregatedIndex 聚合索引
type AggregatedIndex struct {
	Swagger string               `json:"swagger"`
	Info    AggregatedInfo       `json:"info"`
	Apis    []ServiceSwaggerInfo `json:"apis"`
}

// AggregatedInfo 聚合信息
type AggregatedInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// 服务端口映射
var servicePorts = map[string]int{
	"admin-service":      8001,
	"affiliate-service":  8016,
	"ai-service":         8020,
	"cart-service":       8004,
	"catalog-service":    8003,
	"cms-service":        8007,
	"customer-service":   8002,
	"import-service":     8018,
	"inventory-service":  8008,
	"localization-service": 8015,
	"media-service":      8014,
	"notification-service": 8009,
	"order-service":      8005,
	"payment-service":    8010,
	"price-service":      8011,
	"seo-service":        8017,
	"shipping-service":   8012,
	"store-service":      8013,
	"theme-service":      8019,
}

func main() {
	// 获取项目根目录
	cwd, _ := os.Getwd()
	nopGoDir := cwd

	// 检查是否在 scripts 目录运行
	if filepath.Base(cwd) == "gen-swagger" {
		nopGoDir = filepath.Join(cwd, "..", "..")
	}

	// 确认是正确的目录
	if _, err := os.Stat(filepath.Join(nopGoDir, "services")); os.IsNotExist(err) {
		fmt.Printf("错误: 找不到 services 目录，当前工作目录: %s\n", cwd)
		fmt.Println("请在 nop-go 根目录或 scripts/gen-swagger 目录运行此脚本")
		os.Exit(1)
	}

	fmt.Println("=== Swagger 文档生成工具 ===")
	fmt.Printf("工作目录: %s\n\n", nopGoDir)

	// 创建聚合文档目录
	aggregatedDir := filepath.Join(nopGoDir, "docs", "swagger")
	os.MkdirAll(aggregatedDir, 0755)

	var apiInfos []ServiceSwaggerInfo

	// 为每个服务创建 swagger 注解文件
	for _, service := range services {
		serviceDir := filepath.Join(nopGoDir, "services", service)
		if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
			fmt.Printf("跳过 %s (目录不存在)\n", service)
			continue
		}

		fmt.Printf("处理 %s...\n", service)

		// 创建 docs 目录
		docsDir := filepath.Join(serviceDir, "docs")
		os.MkdirAll(docsDir, 0755)

		// 创建 swagger 注解文件
		createSwaggerAnnotations(service, serviceDir)

		// 添加到聚合列表
		port := servicePorts[service]
		if port == 0 {
			port = 8000
		}

		title := getServiceTitle(service)
		apiInfos = append(apiInfos, ServiceSwaggerInfo{
			Name:        service,
			Title:       title,
			Description: getServiceDescription(service),
			Version:     "1.0",
			Port:        port,
		})
	}

	// 生成聚合索引
	createAggregatedIndex(aggregatedDir, apiInfos)

	fmt.Println("\n=== 完成 ===")
	fmt.Println("\n使用说明:")
	fmt.Println("1. 安装 swag: go install github.com/swaggo/swag/cmd/swag@latest")
	fmt.Println("2. 在各服务目录运行: swag init -g ./cmd/main.go -o ./docs")
	fmt.Println("3. 各服务访问: http://localhost:{port}/swagger/index.html")
	fmt.Println("4. 网关聚合: http://localhost:8000/swagger/aggregated")
}

// getServiceTitle 获取服务标题
func getServiceTitle(service string) string {
	titles := map[string]string{
		"admin-service":        "后台管理服务",
		"affiliate-service":    "联盟营销服务",
		"ai-service":           "AI智能服务",
		"cart-service":         "购物车服务",
		"catalog-service":      "商品目录服务",
		"cms-service":          "内容管理服务",
		"customer-service":     "客户服务",
		"import-service":       "导入导出服务",
		"inventory-service":    "库存服务",
		"localization-service": "多语言服务",
		"media-service":        "媒体服务",
		"notification-service": "通知服务",
		"order-service":        "订单服务",
		"payment-service":      "支付服务",
		"price-service":        "价格服务",
		"seo-service":          "SEO服务",
		"shipping-service":     "物流服务",
		"store-service":        "多店铺服务",
		"theme-service":        "主题服务",
	}
	if title, ok := titles[service]; ok {
		return title
	}
	return service
}

// getServiceDescription 获取服务描述
func getServiceDescription(service string) string {
	descs := map[string]string{
		"admin-service":        "后台管理、用户权限、系统配置",
		"affiliate-service":    "联盟营销、佣金管理、推广追踪",
		"ai-service":           "AI聊天、商品推荐、内容生成",
		"cart-service":         "购物车、优惠券、促销",
		"catalog-service":      "商品管理、分类、品牌、规格",
		"cms-service":          "页面、博客、菜单、投票",
		"customer-service":     "用户注册登录、地址管理、GDPR合规",
		"import-service":       "数据导入导出、Excel处理",
		"inventory-service":    "库存管理、仓储、库存预警",
		"localization-service": "多语言、货币、时区管理",
		"media-service":        "图片、文件上传与管理",
		"notification-service": "邮件、短信、站内通知",
		"order-service":        "订单流程、退款、状态管理",
		"payment-service":      "支付方式、交易、退款",
		"price-service":        "价格规则、折扣、会员价",
		"seo-service":          "URL优化、元数据、站点地图",
		"shipping-service":     "配送方式、运费计算、物流追踪",
		"store-service":        "多店铺、多供应商管理",
		"theme-service":        "主题管理、样式变量、模板",
	}
	if desc, ok := descs[service]; ok {
		return desc
	}
	return service
}

// createSwaggerAnnotations 创建 swagger 注解文件
func createSwaggerAnnotations(service, serviceDir string) {
	title := getServiceTitle(service)
	desc := getServiceDescription(service)
	port := servicePorts[service]
	if port == 0 {
		port = 8000
	}

	// 创建 main.go 顶部注解
	annotations := fmt.Sprintf(`// Package main %s
//
// %s API 文档
//
// @title           %s
// @version         1.0
// @description     %s
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:%d
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main
`, title, title, title, desc, port)

	// 写入注解文件（供参考，实际需要合并到 main.go）
	annoFile := filepath.Join(serviceDir, "docs", "annotations.go")
	os.WriteFile(annoFile, []byte(annotations), 0644)
}

// createAggregatedIndex 创建聚合索引
func createAggregatedIndex(dir string, apis []ServiceSwaggerInfo) {
	index := AggregatedIndex{
		Swagger: "2.0",
		Info: AggregatedInfo{
			Title:       "nop-go 微服务 API 文档",
			Description: "nopCommerce Go 版本微服务聚合 API 文档",
			Version:     "1.0.0",
		},
		Apis: apis,
	}

	data, _ := json.MarshalIndent(index, "", "  ")
	indexFile := filepath.Join(dir, "index.json")
	os.WriteFile(indexFile, data, 0644)

	fmt.Printf("\n聚合索引: %s\n", indexFile)

	// 创建 HTML 聚合页面
	htmlContent := generateSwaggerHTML(apis)
	htmlFile := filepath.Join(dir, "index.html")
	os.WriteFile(htmlFile, []byte(htmlContent), 0644)
}

// generateSwaggerHTML 生成聚合 HTML 页面
func generateSwaggerHTML(apis []ServiceSwaggerInfo) string {
	var links strings.Builder
	for _, api := range apis {
		links.WriteString(fmt.Sprintf(`
        <div class="service-card">
            <h3>%s</h3>
            <p>%s</p>
            <p><strong>端口:</strong> %d</p>
            <a href="http://localhost:%d/swagger/index.html" target="_blank">打开 API 文档</a>
        </div>
`, api.Title, api.Description, api.Port, api.Port))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>nop-go API 文档</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            padding: 40px 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            color: white;
            text-align: center;
            margin-bottom: 40px;
            font-size: 2.5em;
        }
        .services {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
        }
        .service-card {
            background: white;
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            transition: transform 0.2s;
        }
        .service-card:hover {
            transform: translateY(-4px);
        }
        .service-card h3 {
            color: #333;
            margin-bottom: 8px;
        }
        .service-card p {
            color: #666;
            font-size: 14px;
            margin-bottom: 16px;
        }
        .service-card a {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 10px 20px;
            border-radius: 6px;
            text-decoration: none;
            font-weight: 500;
        }
        .service-card a:hover {
            background: #5a6fd6;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>nop-go 微服务 API 文档</h1>
        <div class="services">
%s
        </div>
    </div>
</body>
</html>`, links.String())
}
// Package service 网关服务HTTP层
package service

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Route 路由配置
type Route struct {
	Path   string
	Target string
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Port        int    `json:"port"`
	SwaggerURL  string `json:"swagger_url"`
}

// GatewayService 网关服务
type GatewayService struct {
	routes   []Route
	services []ServiceInfo
}

// NewGatewayService 创建网关服务
func NewGatewayService(routes []Route) *GatewayService {
	gs := &GatewayService{
		routes: routes,
		services: getServiceList(),
	}
	return gs
}

// RegisterRoutes 注册路由
func (s *GatewayService) RegisterRoutes(r *gin.Engine) {
	// 聚合 Swagger 文档页面
	r.GET("/swagger", s.swaggerIndex)
	r.GET("/swagger/aggregated", s.aggregatedSwagger)

	// API 文档聚合 JSON
	r.GET("/api-docs", s.apiDocs)

	// 业务代理统一收口到 /api 下，避免与 swagger 等固定路由冲突
	r.Any("/api/*path", s.proxy)
}

// swaggerIndex 聚合 Swagger 首页
func (s *GatewayService) swaggerIndex(c *gin.Context) {
	html := s.generateSwaggerIndexHTML()
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// aggregatedSwagger 返回聚合的服务列表
func (s *GatewayService) aggregatedSwagger(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"swagger": "2.0",
		"info": gin.H{
			"title":       "nop-go 微服务 API 文档",
			"description": "nopCommerce Go 版本微服务聚合 API 文档",
			"version":     "1.0.0",
		},
		"services": s.services,
	})
}

// apiDocs 返回 OpenAPI 兼容的聚合文档
func (s *GatewayService) apiDocs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       "nop-go API Gateway",
			"description": "API Gateway for nop-go microservices",
			"version":     "1.0.0",
		},
		"services": s.services,
	})
}

// proxy 代理请求到后端服务
func (s *GatewayService) proxy(c *gin.Context) {
	path := c.Request.URL.Path

	// 排除 swagger 相关路径
	if strings.HasPrefix(path, "/swagger") || path == "/api-docs" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	for _, route := range s.routes {
		if strings.HasPrefix(path, route.Path) {
			targetURL, err := url.Parse(route.Target)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target URL"})
				return
			}

			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.ServeHTTP(c.Writer, c.Request)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
}

// generateSwaggerIndexHTML 生成聚合 Swagger 首页 HTML
func (s *GatewayService) generateSwaggerIndexHTML() string {
	var serviceCards string
	for _, svc := range s.services {
		serviceCards += `
        <div class="service-card">
            <h3>` + svc.Title + `</h3>
            <p>` + svc.Description + `</p>
            <p><strong>端口:</strong> ` + fmt.Sprintf("%d", svc.Port) + `</p>
            <a href="http://localhost:` + fmt.Sprintf("%d", svc.Port) + `/swagger/index.html" target="_blank">打开 API 文档</a>
            <a href="/proxy` + svc.SwaggerURL + `" target="_blank" class="secondary">通过网关访问</a>
        </div>
`
	}

	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>nop-go API 网关 - Swagger 文档</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            padding: 40px 20px;
            color: #fff;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            margin-bottom: 16px;
            font-size: 2.5em;
            background: linear-gradient(90deg, #667eea, #764ba2);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .subtitle {
            text-align: center;
            color: #888;
            margin-bottom: 40px;
        }
        .services {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
            gap: 24px;
        }
        .service-card {
            background: rgba(255,255,255,0.05);
            border: 1px solid rgba(255,255,255,0.1);
            border-radius: 16px;
            padding: 24px;
            transition: all 0.3s ease;
        }
        .service-card:hover {
            transform: translateY(-4px);
            background: rgba(255,255,255,0.08);
            border-color: rgba(102,126,234,0.5);
        }
        .service-card h3 {
            color: #fff;
            margin-bottom: 8px;
            font-size: 1.2em;
        }
        .service-card p {
            color: #aaa;
            font-size: 14px;
            margin-bottom: 16px;
        }
        .service-card a {
            display: inline-block;
            background: linear-gradient(90deg, #667eea, #764ba2);
            color: white;
            padding: 10px 20px;
            border-radius: 8px;
            text-decoration: none;
            font-weight: 500;
            margin-right: 8px;
            margin-bottom: 8px;
        }
        .service-card a:hover {
            opacity: 0.9;
        }
        .service-card a.secondary {
            background: transparent;
            border: 1px solid #667eea;
        }
        .footer {
            text-align: center;
            margin-top: 60px;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>nop-go API 文档中心</h1>
        <p class="subtitle">nopCommerce Go 版本微服务 API 文档聚合</p>
        <div class="services">
` + serviceCards + `
        </div>
        <div class="footer">
            <p>共 ` + fmt.Sprintf("%d", len(s.services)) + ` 个微服务 | <a href="/swagger/aggregated" style="color:#667eea">JSON API</a></p>
        </div>
    </div>
</body>
</html>`
}

// getServiceList 获取服务列表
func getServiceList() []ServiceInfo {
	return []ServiceInfo{
		{Name: "admin-service", Title: "后台管理服务", Description: "后台管理、用户权限、系统配置", Port: 8001, SwaggerURL: "/api/admin/swagger.json"},
		{Name: "customer-service", Title: "客户服务", Description: "用户注册登录、地址管理、GDPR合规", Port: 8002, SwaggerURL: "/api/customer/swagger.json"},
		{Name: "catalog-service", Title: "商品目录服务", Description: "商品管理、分类、品牌、规格", Port: 8003, SwaggerURL: "/api/catalog/swagger.json"},
		{Name: "cart-service", Title: "购物车服务", Description: "购物车、优惠券、促销", Port: 8004, SwaggerURL: "/api/cart/swagger.json"},
		{Name: "order-service", Title: "订单服务", Description: "订单流程、退款、状态管理", Port: 8005, SwaggerURL: "/api/order/swagger.json"},
		{Name: "cms-service", Title: "内容管理服务", Description: "页面、博客、菜单、投票", Port: 8007, SwaggerURL: "/api/cms/swagger.json"},
		{Name: "inventory-service", Title: "库存服务", Description: "库存管理、仓储、库存预警", Port: 8008, SwaggerURL: "/api/inventory/swagger.json"},
		{Name: "notification-service", Title: "通知服务", Description: "邮件、短信、站内通知", Port: 8009, SwaggerURL: "/api/notification/swagger.json"},
		{Name: "payment-service", Title: "支付服务", Description: "支付方式、交易、退款", Port: 8010, SwaggerURL: "/api/payment/swagger.json"},
		{Name: "price-service", Title: "价格服务", Description: "价格规则、折扣、会员价", Port: 8011, SwaggerURL: "/api/price/swagger.json"},
		{Name: "shipping-service", Title: "物流服务", Description: "配送方式、运费计算、物流追踪", Port: 8012, SwaggerURL: "/api/shipping/swagger.json"},
		{Name: "store-service", Title: "多店铺服务", Description: "多店铺、多供应商管理", Port: 8013, SwaggerURL: "/api/store/swagger.json"},
		{Name: "media-service", Title: "媒体服务", Description: "图片、文件上传与管理", Port: 8014, SwaggerURL: "/api/media/swagger.json"},
		{Name: "localization-service", Title: "多语言服务", Description: "多语言、货币、时区管理", Port: 8015, SwaggerURL: "/api/localization/swagger.json"},
		{Name: "affiliate-service", Title: "联盟营销服务", Description: "联盟营销、佣金管理、推广追踪", Port: 8016, SwaggerURL: "/api/affiliate/swagger.json"},
		{Name: "seo-service", Title: "SEO服务", Description: "URL优化、元数据、站点地图", Port: 8017, SwaggerURL: "/api/seo/swagger.json"},
		{Name: "import-service", Title: "导入导出服务", Description: "数据导入导出、Excel处理", Port: 8018, SwaggerURL: "/api/import/swagger.json"},
		{Name: "theme-service", Title: "主题服务", Description: "主题管理、样式变量、模板", Port: 8019, SwaggerURL: "/api/theme/swagger.json"},
		{Name: "ai-service", Title: "AI智能服务", Description: "AI聊天、商品推荐、内容生成", Port: 8020, SwaggerURL: "/api/ai/swagger.json"},
	}
}
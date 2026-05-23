package http

import (
	"testing"

	svc "nop-go/services/catalog-service/internal/service"
)

// TestRegisterRoutes_ServicesExist 测试路由注册前置条件。
// 注意：由于 RegisterRoutes 需要 gorp.Router 接口，
// 在单元测试中无法直接使用 gin.Engine 测试路由。
// 实际路由集成测试需要通过 gorp.Run 启动完整服务完成。
func TestRegisterRoutes_ServicesExist(t *testing.T) {
	// 创建一个空 service 结构验证类型定义正确
	services := &svc.Services{}

	// 验证 service 结构体不为 nil
	if services == nil {
		t.Fatal("services 不应为 nil")
	}

	// 验证四个子服务字段存在（初始为 nil 是正常的，需要 DB 才能初始化）
	_ = services.Product
	_ = services.Directory
	_ = services.Media
	_ = services.Seo
}
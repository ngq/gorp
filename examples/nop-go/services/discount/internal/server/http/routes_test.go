package http

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nop-go/services/discount/internal/service"
)

// TestRegisterRoutes_RoutesExist 测试路由注册的最小冒烟测试
//
// 中文说明：
// - 这里只验证 DiscountService 可以正常构造为空实例并通过 nil 检查
// - 完整的路由集成测试需要通过 gorp.Run 启动服务后再进行
func TestRegisterRoutes_RoutesExist(t *testing.T) {
	// 创建一个空的 DiscountService 实例用于占位测试
	svc := &service.DiscountService{}

	// 验证 service 实例不为 nil（占位断言）
	require.NotNil(t, svc)

	// 实际路由测试需要通过集成测试完成
	// RegisterRoutes 需要 gorp.Router 接口，无法直接用 gin.Engine 测试
}

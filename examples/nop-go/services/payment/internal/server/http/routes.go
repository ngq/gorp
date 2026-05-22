// Package http 提供 payment 服务的 HTTP 路由注册
package http

import (
	"nop-go/services/payment/internal/server/http/handler"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册支付服务的所有 HTTP 路由
// 路由分组: /api/v1/payment
//
// 中文说明：
// - Gin-first 模式：直接使用原生 Gin Engine 注册路由；
// - 抽象契约模式：使用 gorp.Router 抽象接口；
func RegisterRoutes(r gorp.Router, h *handler.PaymentHandler) {
	// 支付方式相关路由
	r.GET("/api/v1/payment/methods", h.ListPaymentMethods)         // 支付方式列表
	r.PUT("/api/v1/payment/methods/:id", h.UpdatePaymentMethod)    // 更新支付方式

	// 支付方式限制相关路由
	r.GET("/api/v1/payment/method-restrictions", h.ListMethodRestrictions)    // 支付方式限制列表
	r.PUT("/api/v1/payment/method-restrictions", h.UpdateMethodRestrictions)  // 更新支付方式限制
}

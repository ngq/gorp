// Package http 注册 HTTP 路由。
//
// 本文件负责将所有订单服务的路由端点映射到对应 handler 方法。
// 路由分组：
//   - /api/v1/orders        — 订单 CRUD 及操作（取消、退款、PDF、重新下单等）
//   - /api/v1/shopping-cart — 购物车（查看、添加、更新、折扣码、礼品卡）
//   - /api/v1/wishlist      — 愿望清单（查看、添加）
//   - /api/v1/checkout      — 结账流程（地址、配送、支付、确认）
//   - /api/v1/return-requests — 退货请求（列表、提交）
package http

import (
	"nop-go/services/order/internal/server/http/handler"
	"nop-go/services/order/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册 HTTP 路由。
//
// 使用 gorp.Router 抽象接口注册所有订单服务的路由端点。
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	// 创建 handler，注入所有子服务
	h := handler.NewOrderHandler(
		services.Order,
		services.Cart,
		services.Wishlist,
		services.Checkout,
		services.ReturnRequest,
	)

	// ========================================================================
	// 订单路由组 /api/v1/orders
	// ========================================================================
	orders := r.Group("/api/v1/orders")
	{
		orders.GET("", h.List)                          // 订单列表
		orders.GET("/:id", h.GetByID)                   // 订单详情
		orders.POST("", h.Create)                       // 创建订单
		orders.DELETE("/:id", h.Delete)                 // 删除订单
		orders.POST("/:id/cancel", h.CancelOrder)       // 取消订单
		orders.POST("/:id/refund", h.RefundOrder)       // 退款
		orders.GET("/:id/pdf", h.GetPDFInvoice)         // PDF发票
		orders.POST("/:id/repost-payment", h.RePostPayment) // 重新提交支付
		orders.GET("/:id/shipment/:shipmentId", h.GetShipmentDetail) // 配送详情
		orders.POST("/:id/reorder", h.Reorder)          // 重新下单
	}

	// ========================================================================
	// 购物车路由组 /api/v1/shopping-cart
	// ========================================================================
	cart := r.Group("/api/v1/shopping-cart")
	{
		cart.GET("", h.GetShoppingCart)                    // 购物车
		cart.POST("/update", h.UpdateCart)                // 更新购物车
		cart.POST("/add", h.AddToCart)                    // 添加到购物车
		cart.POST("/apply-coupon", h.ApplyCoupon)         // 应用折扣码
		cart.POST("/apply-gift-card", h.ApplyGiftCard)    // 应用礼品卡
	}

	// ========================================================================
	// 愿望清单路由组 /api/v1/wishlist
	// ========================================================================
	wishlist := r.Group("/api/v1/wishlist")
	{
		wishlist.GET("", h.GetWishlist)          // 愿望清单
		wishlist.POST("/add", h.AddToWishlist)   // 添加到愿望清单
	}

	// ========================================================================
	// 结账路由组 /api/v1/checkout
	// ========================================================================
	checkout := r.Group("/api/v1/checkout")
	{
		checkout.GET("", h.GetCheckout)                          // 结账信息
		checkout.POST("/billing-address", h.SetBillingAddress)   // 账单地址
		checkout.POST("/shipping-address", h.SetShippingAddress) // 配送地址
		checkout.POST("/shipping-method", h.SetShippingMethod)  // 配送方式
		checkout.POST("/payment-method", h.SetPaymentMethod)    // 支付方式
		checkout.POST("/confirm", h.ConfirmCheckout)            // 确认订单
	}

	// ========================================================================
	// 退货请求路由组 /api/v1/return-requests
	// ========================================================================
	returnRequests := r.Group("/api/v1/return-requests")
	{
		returnRequests.GET("", h.ListReturnRequests)    // 退货请求列表
		returnRequests.POST("", h.CreateReturnRequest)  // 提交退货请求
	}
}
// Package http 注册交易服务 HTTP 路由
// 路由分组：
//   /api/v1/orders          — 订单 CRUD
//   /api/v1/shopping-cart   — 购物车
//   /api/v1/wishlist        — 心愿单
//   /api/v1/checkout        — 结账
//   /api/v1/return-requests — 退换货
//   /api/v1/payment         — 支付
//   /api/v1/shipping        — 物流
//   /api/v1/tax             — 税务
package http

import (
	"nop-go/services/trade-service/internal/server/http/handler"
	"nop-go/services/trade-service/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册交易服务的所有 HTTP 路由
// 所有 handler 通过 service.Services 容器注入，统一模式
func RegisterRoutes(r gorp.Router, services *service.Services) {
	// 创建所有 handler，统一注入 Services 容器
	orderH := handler.NewOrderHandler(services)
	paymentH := handler.NewPaymentHandler(services)
	shippingH := handler.NewShippingHandler(services)
	taxH := handler.NewTaxHandler(services)

	// ========================================================================
	// 订单路由组 /api/v1/orders
	// ========================================================================
	orders := r.Group("/api/v1/orders")
	{
		orders.GET("", orderH.ListOrders)            // 订单列表
		orders.GET("/:id", orderH.GetOrder)          // 订单详情
		orders.POST("", orderH.CreateOrder)          // 创建订单
		orders.DELETE("/:id", orderH.DeleteOrder)    // 删除订单
	}

	// ========================================================================
	// 购物车路由组 /api/v1/shopping-cart
	// ========================================================================
	cart := r.Group("/api/v1/shopping-cart")
	{
		cart.GET("", orderH.GetCart)                         // 购物车
		cart.POST("/add", orderH.AddToCart)                  // 添加到购物车
		cart.POST("/update", orderH.UpdateCartItem)          // 更新购物车项
		cart.DELETE("/:productId", orderH.RemoveFromCart)    // 移除购物车项
	}

	// ========================================================================
	// 心愿单路由组 /api/v1/wishlist
	// ========================================================================
	wishlist := r.Group("/api/v1/wishlist")
	{
		wishlist.GET("", orderH.GetWishlist)                     // 心愿单
		wishlist.POST("/add", orderH.AddToWishlist)              // 添加到心愿单
		wishlist.DELETE("/:productId", orderH.RemoveFromWishlist) // 移除心愿单项
	}

	// ========================================================================
	// 结账路由组 /api/v1/checkout
	// ========================================================================
	r.POST("/api/v1/checkout", orderH.Checkout)               // 结账

	// ========================================================================
	// 退换货路由组 /api/v1/return-requests
	// ========================================================================
	returnRequests := r.Group("/api/v1/return-requests")
	{
		returnRequests.GET("", orderH.ListReturnRequests)      // 退换货列表
		returnRequests.POST("", orderH.CreateReturnRequest)    // 提交退换货
	}

	// ========================================================================
	// 支付路由组 /api/v1/payment
	// ========================================================================
	payment := r.Group("/api/v1/payment")
	{
		payment.POST("", paymentH.CreatePayment)               // 创建支付
		payment.GET("/:id", paymentH.GetPayment)               // 支付详情
		payment.GET("/order/:orderId", paymentH.ListPaymentsByOrder) // 按订单查询支付
		payment.GET("/methods", paymentH.ListPaymentMethods)   // 支付方式列表
		payment.POST("/methods", paymentH.CreatePaymentMethod) // 创建支付方式
		payment.DELETE("/methods/:id", paymentH.DeletePaymentMethod) // 删除支付方式
	}

	// ========================================================================
	// 物流路由组 /api/v1/shipping
	// ========================================================================
	shipping := r.Group("/api/v1/shipping")
	{
		// 物流服务商
		shipping.GET("/providers", shippingH.ListShippingProviders)         // 物流服务商列表
		shipping.POST("/providers", shippingH.CreateShippingProvider)       // 创建物流服务商
		shipping.PUT("/providers/:id", shippingH.UpdateShippingProvider)    // 更新物流服务商
		shipping.DELETE("/providers/:id", shippingH.DeleteShippingProvider) // 删除物流服务商

		// 物流订单
		shipping.POST("/orders", shippingH.CreateShippingOrder)             // 创建物流订单
		shipping.GET("/orders/:id", shippingH.GetShippingOrder)             // 物流订单详情
		shipping.PUT("/orders/:id", shippingH.UpdateShippingOrder)          // 更新物流订单

		// 物流事件
		shipping.POST("/events", shippingH.CreateShippingEvent)             // 创建物流事件
		shipping.GET("/events/:shippingOrderId", shippingH.ListShippingEvents) // 物流事件列表
	}

	// ========================================================================
	// 税务路由组 /api/v1/tax
	// ========================================================================
	tax := r.Group("/api/v1/tax")
	{
		// 税务服务商
		tax.GET("/providers", taxH.ListTaxProviders)               // 税务服务商列表
		tax.POST("/providers", taxH.CreateTaxProvider)             // 创建税务服务商
		tax.PUT("/providers/:id", taxH.UpdateTaxProvider)          // 更新税务服务商
		tax.DELETE("/providers/:id", taxH.DeleteTaxProvider)       // 删除税务服务商

		// 税种分类
		tax.GET("/categories", taxH.ListTaxCategories)             // 税种分类列表
		tax.POST("/categories", taxH.CreateTaxCategory)            // 创建税种分类
		tax.PUT("/categories/:id", taxH.UpdateTaxCategory)         // 更新税种分类
		tax.DELETE("/categories/:id", taxH.DeleteTaxCategory)      // 删除税种分类

		// 税率
		tax.GET("/rates/:categoryId", taxH.ListTaxRates)           // 税率列表
		tax.POST("/rates", taxH.CreateTaxRate)                     // 创建税率
	}
}
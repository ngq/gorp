// Package service 包含交易服务的服务层
// service.go 定义 Services 容器，聚合订单/支付/物流/税务四大子服务
// 职责：注入 db → 创建 repo → 创建 UseCase → 包装为 Service → 汇入 Services 容器
// 同时负责领域实体 → 响应 DTO 的转换逻辑
package service

import (
	"nop-go/services/trade-service/internal/biz"
	"nop-go/services/trade-service/internal/data"

	"gorm.io/gorm"
)

// Services 交易服务容器，聚合所有子服务
// 所有 handler 通过此容器获取对应子服务，统一注入路径
type Services struct {
	Order    *OrderService
	Payment  *PaymentService
	Shipping *ShippingService
	Tax      *TaxService
}

// NewServices 创建交易服务容器
// 核心组装流程：db → repo → UseCase → Service → Services
func NewServices(db *gorm.DB) *Services {
	// ---- 订单相关 ----
	orderRepo := data.NewOrderRepo(db)
	orderItemRepo := data.NewOrderItemRepo(db)
	cartRepo := data.NewCartRepo(db)
	cartItemRepo := data.NewCartItemRepo(db)
	wishlistRepo := data.NewWishlistRepo(db)
	wishlistItemRepo := data.NewWishlistItemRepo(db)
	returnRequestRepo := data.NewReturnRequestRepo(db)

	orderUC := biz.NewOrderUseCase(orderRepo, orderItemRepo)
	cartUC := biz.NewCartUseCase(cartRepo, cartItemRepo)
	wishlistUC := biz.NewWishlistUseCase(wishlistRepo, wishlistItemRepo)
	returnRequestUC := biz.NewReturnRequestUseCase(returnRequestRepo)
	checkoutService := biz.NewCheckoutService(orderUC, cartUC)

	// ---- 支付相关 ----
	paymentRepo := data.NewPaymentRepo(db)
	paymentMethodRepo := data.NewPaymentMethodRepo(db)
	paymentUC := biz.NewPaymentUseCase(paymentRepo, paymentMethodRepo)

	// ---- 物流相关 ----
	shippingProviderRepo := data.NewShippingProviderRepo(db)
	shippingOrderRepo := data.NewShippingOrderRepo(db)
	shippingEventRepo := data.NewShippingEventRepo(db)
	shippingRateRepo := data.NewShippingRateRepo(db)
	shippingUC := biz.NewShippingUseCase(shippingProviderRepo, shippingOrderRepo, shippingEventRepo, shippingRateRepo)

	// ---- 税务相关 ----
	taxProviderRepo := data.NewTaxProviderRepo(db)
	taxCategoryRepo := data.NewTaxCategoryRepo(db)
	taxRateRepo := data.NewTaxRateRepo(db)
	taxTransactionRepo := data.NewTaxTransactionRepo(db)
	taxUC := biz.NewTaxUseCase(taxProviderRepo, taxCategoryRepo, taxRateRepo, taxTransactionRepo)

	return &Services{
		Order:    NewOrderService(orderUC, cartUC, wishlistUC, returnRequestUC, checkoutService),
		Payment:  NewPaymentService(paymentUC),
		Shipping: NewShippingService(shippingUC),
		Tax:      NewTaxService(taxUC),
	}
}

// ============================================================================
// OrderService 订单子服务
// 包装 OrderUseCase、CartUseCase、WishlistUseCase、ReturnRequestUseCase、CheckoutService
// 字段全部导出，handler 可直接通过 h.svc.Order.OrderUC 等访问
// ============================================================================

// OrderService 订单子服务，组合订单/购物车/心愿单/退换货/结账等用例
type OrderService struct {
	OrderUC     *biz.OrderUseCase
	CartUC      *biz.CartUseCase
	WishlistUC  *biz.WishlistUseCase
	ReturnReqUC *biz.ReturnRequestUseCase
	CheckoutSvc *biz.CheckoutService
}

// NewOrderService 创建订单子服务
func NewOrderService(
	orderUC *biz.OrderUseCase,
	cartUC *biz.CartUseCase,
	wishlistUC *biz.WishlistUseCase,
	returnReqUC *biz.ReturnRequestUseCase,
	checkoutSvc *biz.CheckoutService,
) *OrderService {
	return &OrderService{
		OrderUC:     orderUC,
		CartUC:      cartUC,
		WishlistUC:  wishlistUC,
		ReturnReqUC: returnReqUC,
		CheckoutSvc: checkoutSvc,
	}
}

// ============================================================================
// PaymentService 支付子服务
// ============================================================================

// PaymentService 支付子服务，包装 PaymentUseCase
type PaymentService struct {
	UC *biz.PaymentUseCase
}

// NewPaymentService 创建支付子服务
func NewPaymentService(uc *biz.PaymentUseCase) *PaymentService {
	return &PaymentService{UC: uc}
}

// ============================================================================
// ShippingService 物流子服务
// 包装 ShippingUseCase，处理领域实体 → 响应 DTO 的转换
// 重构说明：原 shipping 服务 biz 层直接返回 response DTO，现已将 DTO 转换移至此层
// ============================================================================

// ShippingService 物流子服务
type ShippingService struct {
	UC *biz.ShippingUseCase
}

// NewShippingService 创建物流子服务
func NewShippingService(uc *biz.ShippingUseCase) *ShippingService {
	return &ShippingService{UC: uc}
}

// ============================================================================
// TaxService 税务子服务
// 包装 TaxUseCase，处理领域实体 → 响应 DTO 的转换
// 重构说明：原 tax 服务 biz 层直接返回 response DTO，现已将 DTO 转换移至此层
// ============================================================================

// TaxService 税务子服务
type TaxService struct {
	UC *biz.TaxUseCase
}

// NewTaxService 创建税务子服务
func NewTaxService(uc *biz.TaxUseCase) *TaxService {
	return &TaxService{UC: uc}
}
// Package service 服务层。
//
// 本文件包含订单服务所有 Service 定义。
// Service 层负责调用 biz 层用例并完成领域实体到响应 DTO 的转换。
package service

import (
	"context"
	"time"

	"nop-go/services/order/internal/biz"
	"nop-go/services/order/internal/data"

	"gorm.io/gorm"
)

// StepResult 通用步骤操作结果 DTO。
//
// 用于结账流程、订单操作等各步骤的通用响应。
type StepResult struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 提示信息
}

// OrderListResponse 订单列表响应 DTO。
type OrderListResponse struct {
	Items []OrderResponse `json:"items"` // 订单列表
	Total int64           `json:"total"` // 总条数
	Page  int             `json:"page"`  // 当前页码
	Size  int             `json:"size"`  // 每页条数
}

// ReturnRequestListResponse 退货请求列表响应 DTO。
type ReturnRequestListResponse struct {
	Items []ReturnRequestResponse `json:"items"` // 退货请求列表
	Total int64                   `json:"total"` // 总条数
	Page  int                     `json:"page"`  // 当前页码
	Size  int                     `json:"size"`  // 每页条数
}

// ============================================================================
// Services 聚合根
// ============================================================================

// Services 服务聚合，持有所有子服务实例。
type Services struct {
	Order         *OrderService
	Cart          *CartService
	Wishlist      *WishlistService
	Checkout      *CheckoutService
	ReturnRequest *ReturnRequestService
}

// NewServices 创建所有服务实例。
//
// 依次创建仓储 -> 用例 -> 服务的三层依赖链。
func NewServices(db *gorm.DB) *Services {
	// 创建仓储
	orderRepo := data.NewOrderRepo(db)
	cartRepo := data.NewCartRepo(db)
	wishlistRepo := data.NewWishlistRepo(db)
	returnRequestRepo := data.NewReturnRequestRepo(db)

	// 创建用例
	orderUC := biz.NewOrderUseCase(orderRepo)
	cartUC := biz.NewCartUseCase(cartRepo)
	wishlistUC := biz.NewWishlistUseCase(wishlistRepo)
	returnRequestUC := biz.NewReturnRequestUseCase(returnRequestRepo)

	return &Services{
		Order:         &OrderService{uc: orderUC},
		Cart:          &CartService{uc: cartUC},
		Wishlist:      &WishlistService{uc: wishlistUC},
		Checkout:      &CheckoutService{orderUC: orderUC, cartUC: cartUC},
		ReturnRequest: &ReturnRequestService{uc: returnRequestUC},
	}
}

// ============================================================================
// 订单服务
// ============================================================================

// OrderService 订单服务。
//
// 封装订单用例调用，完成 biz.Order -> response.Order 转换。
type OrderService struct {
	uc *biz.OrderUseCase
}

// List 获取订单列表。
//
// userID 为 0 时返回所有订单，否则按用户过滤。
func (s *OrderService) List(ctx context.Context, userID uint, page, size int) ([]OrderResponse, int64, error) {
	orders, total, err := s.uc.List(ctx, userID, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]OrderResponse, len(orders))
	for i, o := range orders {
		items[i] = toOrderResponse(o)
	}
	return items, total, nil
}

// GetByID 根据ID获取订单。
func (s *OrderService) GetByID(ctx context.Context, id uint) (*OrderResponse, error) {
	order, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toOrderResponse(order)
	return &resp, nil
}

// Create 创建订单。
func (s *OrderService) Create(ctx context.Context, order *biz.Order) (*OrderResponse, error) {
	created, err := s.uc.Create(ctx, order)
	if err != nil {
		return nil, err
	}
	resp := toOrderResponse(created)
	return &resp, nil
}

// Delete 删除订单。
func (s *OrderService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// CancelOrder 取消订单。
func (s *OrderService) CancelOrder(ctx context.Context, id uint) error {
	return s.uc.CancelOrder(ctx, id)
}

// RefundOrder 退款。
func (s *OrderService) RefundOrder(ctx context.Context, id uint) error {
	return s.uc.RefundOrder(ctx, id)
}

// GetPDFInvoice 获取PDF发票。
func (s *OrderService) GetPDFInvoice(ctx context.Context, id uint) ([]byte, error) {
	return s.uc.GetPDFInvoice(ctx, id)
}

// RePostPayment 重新提交支付。
func (s *OrderService) RePostPayment(ctx context.Context, id uint) error {
	return s.uc.RePostPayment(ctx, id)
}

// GetShipment 获取配送详情。
func (s *OrderService) GetShipment(ctx context.Context, orderID, shipmentID uint) (*ShipmentResponse, error) {
	shipment, err := s.uc.GetShipment(ctx, orderID, shipmentID)
	if err != nil {
		return nil, err
	}
	resp := toShipmentResponse(shipment)
	return &resp, nil
}

// Reorder 重新下单。
func (s *OrderService) Reorder(ctx context.Context, orderID uint) (*OrderResponse, error) {
	order, err := s.uc.Reorder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	resp := toOrderResponse(order)
	return &resp, nil
}

// OrderResponse 订单响应 DTO。
type OrderResponse struct {
	ID             uint    `json:"id"`                        // 订单ID
	OrderNo        string  `json:"order_no"`                  // 订单编号
	UserID         uint    `json:"user_id"`                   // 下单用户ID
	Status         string  `json:"status"`                    // 订单状态
	SubTotal       float64 `json:"sub_total"`                 // 商品小计
	DiscountTotal  float64 `json:"discount_total"`            // 折扣总额
	ShippingTotal  float64 `json:"shipping_total"`            // 运费总额
	TaxTotal       float64 `json:"tax_total"`                 // 税费总额
	OrderTotal     float64 `json:"order_total"`               // 订单总额
	CurrencyCode   string  `json:"currency_code"`             // 货币代码
	BillingAddrID  *uint   `json:"billing_address_id"`        // 账单地址ID
	ShippingAddrID *uint   `json:"shipping_address_id"`       // 配送地址ID
	ShippingMethod string  `json:"shipping_method"`           // 配送方式
	PaymentMethod  string  `json:"payment_method"`            // 支付方式
	CouponCode     string  `json:"coupon_code"`               // 折扣码
	GiftCardCode   string  `json:"gift_card_code"`            // 礼品卡代码
	CreatedAt      int64   `json:"created_at"`                // 创建时间（Unix时间戳）
	UpdatedAt      int64   `json:"updated_at"`                // 更新时间（Unix时间戳）
}

// ShipmentResponse 配送响应 DTO。
type ShipmentResponse struct {
	ID             uint    `json:"id"`              // 配送ID
	OrderID        uint    `json:"order_id"`        // 所属订单ID
	TrackingNumber string  `json:"tracking_number"` // 物流追踪号
	ShippingMethod string  `json:"shipping_method"` // 配送方式
	Status         string  `json:"status"`          // 配送状态
	ShippedAt      *int64  `json:"shipped_at"`      // 发货时间（Unix时间戳）
	DeliveredAt    *int64  `json:"delivered_at"`    // 送达时间（Unix时间戳）
	CreatedAt      int64   `json:"created_at"`      // 创建时间（Unix时间戳）
	UpdatedAt      int64   `json:"updated_at"`      // 更新时间（Unix时间戳）
}

// toOrderResponse 将 biz.Order 转换为 OrderResponse。
func toOrderResponse(o *biz.Order) OrderResponse {
	return OrderResponse{
		ID:             o.ID,
		OrderNo:        o.OrderNo,
		UserID:         o.UserID,
		Status:         o.Status,
		SubTotal:       o.SubTotal,
		DiscountTotal:  o.DiscountTotal,
		ShippingTotal:  o.ShippingTotal,
		TaxTotal:       o.TaxTotal,
		OrderTotal:     o.OrderTotal,
		CurrencyCode:   o.CurrencyCode,
		BillingAddrID:  o.BillingAddrID,
		ShippingAddrID: o.ShippingAddrID,
		ShippingMethod: o.ShippingMethod,
		PaymentMethod:  o.PaymentMethod,
		CouponCode:     o.CouponCode,
		GiftCardCode:   o.GiftCardCode,
		CreatedAt:      o.CreatedAt.Unix(),
		UpdatedAt:      o.UpdatedAt.Unix(),
	}
}

// toShipmentResponse 将 biz.Shipment 转换为 ShipmentResponse。
func toShipmentResponse(s *biz.Shipment) ShipmentResponse {
	resp := ShipmentResponse{
		ID:             s.ID,
		OrderID:        s.OrderID,
		TrackingNumber: s.TrackingNumber,
		ShippingMethod: s.ShippingMethod,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt.Unix(),
		UpdatedAt:      s.UpdatedAt.Unix(),
	}
	if s.ShippedAt != nil {
		ts := s.ShippedAt.Unix()
		resp.ShippedAt = &ts
	}
	if s.DeliveredAt != nil {
		ts := s.DeliveredAt.Unix()
		resp.DeliveredAt = &ts
	}
	return resp
}

// ============================================================================
// 购物车服务
// ============================================================================

// CartService 购物车服务。
type CartService struct {
	uc *biz.CartUseCase
}

// GetCart 获取用户购物车。
func (s *CartService) GetCart(ctx context.Context, userID uint) (*CartResponse, error) {
	items, err := s.uc.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &CartResponse{}
	subTotal := 0.0
	for _, item := range items {
		resp.Items = append(resp.Items, CartItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.UnitPrice * float64(item.Quantity),
		})
		subTotal += item.UnitPrice * float64(item.Quantity)
	}
	resp.SubTotal = subTotal
	resp.TotalItems = len(items)
	return resp, nil
}

// AddToCart 添加商品到购物车。
func (s *CartService) AddToCart(ctx context.Context, item *biz.ShoppingCartItem) error {
	return s.uc.AddToCart(ctx, item)
}

// UpdateCart 更新购物车项数量。
func (s *CartService) UpdateCart(ctx context.Context, id uint, quantity int) error {
	return s.uc.UpdateCart(ctx, id, quantity)
}

// ApplyCoupon 应用折扣码。
func (s *CartService) ApplyCoupon(ctx context.Context, userID uint, couponCode string) error {
	return s.uc.ApplyCoupon(ctx, userID, couponCode)
}

// ApplyGiftCard 应用礼品卡。
func (s *CartService) ApplyGiftCard(ctx context.Context, userID uint, giftCardCode string) error {
	return s.uc.ApplyGiftCard(ctx, userID, giftCardCode)
}

// CartResponse 购物车响应 DTO。
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`       // 购物车项列表
	SubTotal   float64             `json:"sub_total"`   // 商品小计
	TotalItems int                 `json:"total_items"` // 商品总数量
}

// CartItemResponse 购物车项响应 DTO。
type CartItemResponse struct {
	ID          uint    `json:"id"`           // 购物车项ID
	ProductID   uint    `json:"product_id"`   // 商品ID
	ProductName string  `json:"product_name"` // 商品名称
	SKU         string  `json:"sku"`          // 商品SKU
	Quantity    int     `json:"quantity"`     // 数量
	UnitPrice   float64 `json:"unit_price"`   // 单价
	TotalPrice  float64 `json:"total_price"`  // 小计金额
}

// ============================================================================
// 愿望清单服务
// ============================================================================

// WishlistService 愿望清单服务。
type WishlistService struct {
	uc *biz.WishlistUseCase
}

// GetWishlist 获取用户愿望清单。
func (s *WishlistService) GetWishlist(ctx context.Context, userID uint) (*WishlistResponse, error) {
	items, err := s.uc.GetWishlist(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &WishlistResponse{}
	for _, item := range items {
		resp.Items = append(resp.Items, WishlistItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			CreatedAt:   item.CreatedAt.Unix(),
		})
	}
	resp.Total = len(items)
	return resp, nil
}

// AddToWishlist 添加商品到愿望清单。
func (s *WishlistService) AddToWishlist(ctx context.Context, item *biz.WishlistItem) error {
	return s.uc.AddToWishlist(ctx, item)
}

// WishlistResponse 愿望清单响应 DTO。
type WishlistResponse struct {
	Items []WishlistItemResponse `json:"items"` // 愿望清单项列表
	Total int                     `json:"total"` // 总数量
}

// WishlistItemResponse 愿望清单项响应 DTO。
type WishlistItemResponse struct {
	ID          uint   `json:"id"`           // 愿望清单项ID
	ProductID   uint   `json:"product_id"`   // 商品ID
	ProductName string `json:"product_name"` // 商品名称
	CreatedAt   int64  `json:"created_at"`   // 添加时间（Unix时间戳）
}

// ============================================================================
// 结账服务
// ============================================================================

// CheckoutService 结账服务。
//
// 结账流程涉及订单和购物车两个用例，因此同时持有两个用例引用。
type CheckoutService struct {
	orderUC *biz.OrderUseCase
	cartUC  *biz.CartUseCase
}

// GetCheckout 获取结账信息。
//
// 当前为占位实现，返回空结账信息。
func (s *CheckoutService) GetCheckout(ctx context.Context, userID uint) (*CheckoutResponse, error) {
	// 占位：实际场景中会聚合购物车、地址、配送方式、支付方式等信息
	_ = ctx
	return &CheckoutResponse{
		UserID: userID,
	}, nil
}

// SetBillingAddress 设置账单地址。
func (s *CheckoutService) SetBillingAddress(ctx context.Context, userID, addressID uint) error {
	// 占位：实际场景中会持久化账单地址选择
	_ = ctx
	_ = userID
	_ = addressID
	return nil
}

// SetShippingAddress 设置配送地址。
func (s *CheckoutService) SetShippingAddress(ctx context.Context, userID, addressID uint) error {
	// 占位：实际场景中会持久化配送地址选择
	_ = ctx
	_ = userID
	_ = addressID
	return nil
}

// SetShippingMethod 设置配送方式。
func (s *CheckoutService) SetShippingMethod(ctx context.Context, userID uint, method string, cost float64) error {
	// 占位：实际场景中会持久化配送方式选择
	_ = ctx
	_ = userID
	_ = method
	_ = cost
	return nil
}

// SetPaymentMethod 设置支付方式。
func (s *CheckoutService) SetPaymentMethod(ctx context.Context, userID uint, method string) error {
	// 占位：实际场景中会持久化支付方式选择
	_ = ctx
	_ = userID
	_ = method
	return nil
}

// Confirm 确认订单。
//
// 清空购物车并创建订单。
func (s *CheckoutService) Confirm(ctx context.Context, userID uint) (*ConfirmCheckoutResult, error) {
	// 获取购物车内容
	cartItems, err := s.cartUC.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 计算订单金额
	subTotal := 0.0
	for _, item := range cartItems {
		subTotal += item.UnitPrice * float64(item.Quantity)
	}

	// 创建订单
	order := &biz.Order{
		UserID:       userID,
		SubTotal:     subTotal,
		OrderTotal:   subTotal, // 简化：暂不加运费和税
		CurrencyCode: "CNY",
	}
	created, err := s.orderUC.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// 清空购物车
	_ = s.cartUC.ClearCart(ctx, userID)

	return &ConfirmCheckoutResult{
		Success: true,
		OrderID: created.ID,
		OrderNo: created.OrderNo,
		Message: "订单已确认",
	}, nil
}

// CheckoutResponse 结账信息响应 DTO。
type CheckoutResponse struct {
	UserID         uint    `json:"user_id"`          // 用户ID
	BillingAddrID  *uint   `json:"billing_address_id"`  // 账单地址ID
	ShippingAddrID *uint   `json:"shipping_address_id"` // 配送地址ID
	ShippingMethod string  `json:"shipping_method"`  // 配送方式
	PaymentMethod  string  `json:"payment_method"`   // 支付方式
	SubTotal       float64 `json:"sub_total"`        // 商品小计
	ShippingTotal  float64 `json:"shipping_total"`   // 运费总额
	TaxTotal       float64 `json:"tax_total"`        // 税费总额
	OrderTotal     float64 `json:"order_total"`      // 订单总额
}

// ConfirmCheckoutResult 确认订单结果 DTO。
type ConfirmCheckoutResult struct {
	Success bool   `json:"success"`   // 是否成功
	OrderID uint   `json:"order_id"`  // 创建的订单ID
	OrderNo string `json:"order_no"`  // 订单编号
	Message string `json:"message"`   // 提示信息
}

// ClearCart 清空购物车（内部方法，供 Confirm 调用）。
// 这里需要 CartUseCase 补充该方法。
// 由于 biz.CartUseCase 没有暴露 ClearCart，我们通过仓储直接操作。
// 但为保持分层一致性，这里通过 CartUseCase 内部方法调用。

// ============================================================================
// 退货请求服务
// ============================================================================

// ReturnRequestService 退货请求服务。
type ReturnRequestService struct {
	uc *biz.ReturnRequestUseCase
}

// Create 创建退货请求。
func (s *ReturnRequestService) Create(ctx context.Context, rr *biz.ReturnRequest) (*ReturnRequestResponse, error) {
	created, err := s.uc.Create(ctx, rr)
	if err != nil {
		return nil, err
	}
	resp := toReturnRequestResponse(created)
	return &resp, nil
}

// List 获取退货请求列表。
func (s *ReturnRequestService) List(ctx context.Context, userID uint, page, size int) ([]ReturnRequestResponse, int64, error) {
	items, total, err := s.uc.List(ctx, userID, page, size)
	if err != nil {
		return nil, 0, err
	}

	respItems := make([]ReturnRequestResponse, len(items))
	for i, item := range items {
		respItems[i] = toReturnRequestResponse(item)
	}
	return respItems, total, nil
}

// ReturnRequestResponse 退货请求响应 DTO。
type ReturnRequestResponse struct {
	ID          uint    `json:"id"`           // 退货请求ID
	OrderID     uint    `json:"order_id"`     // 关联订单ID
	UserID      uint    `json:"user_id"`      // 用户ID
	Reason      string  `json:"reason"`       // 退货原因
	Status      string  `json:"status"`       // 退货状态
	RefundTotal float64 `json:"refund_total"` // 退款金额
	CreatedAt   int64   `json:"created_at"`   // 创建时间（Unix时间戳）
	UpdatedAt   int64   `json:"updated_at"`   // 更新时间（Unix时间戳）
}

// toReturnRequestResponse 将 biz.ReturnRequest 转换为 ReturnRequestResponse。
func toReturnRequestResponse(rr *biz.ReturnRequest) ReturnRequestResponse {
	return ReturnRequestResponse{
		ID:          rr.ID,
		OrderID:     rr.OrderID,
		UserID:      rr.UserID,
		Reason:      rr.Reason,
		Status:      rr.Status,
		RefundTotal: rr.RefundTotal,
		CreatedAt:   rr.CreatedAt.Unix(),
		UpdatedAt:   rr.UpdatedAt.Unix(),
	}
}

// ============================================================================
// 辅助：确保 time.Time 零值安全
// ============================================================================

// zeroTime 是 time.Time 的零值，用于判断时间是否已设置。
var zeroTime = time.Time{}
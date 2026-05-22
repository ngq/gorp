// Package biz 业务逻辑层。
//
// 本文件包含订单服务所有领域实体、仓储接口及用例实现。
// 涵盖订单、购物车、愿望清单、退货请求四大业务域。
package biz

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// 订单领域
// ============================================================================

// Order 订单领域实体。
//
// 存储订单主信息，包含金额、状态、地址、支付方式等。
type Order struct {
	ID             uint       // 订单ID
	OrderNo        string     // 订单编号
	UserID         uint       // 下单用户ID
	Status         string     // 订单状态：pending/processing/cancelled/refunded/completed
	SubTotal       float64    // 商品小计
	DiscountTotal  float64    // 折扣总额
	ShippingTotal  float64    // 运费总额
	TaxTotal       float64    // 税费总额
	OrderTotal     float64    // 订单总额
	CurrencyCode   string     // 货币代码
	BillingAddrID  *uint      // 账单地址ID
	ShippingAddrID *uint      // 配送地址ID
	ShippingMethod string     // 配送方式
	PaymentMethod  string     // 支付方式
	CouponCode     string     // 使用的折扣码
	GiftCardCode   string     // 使用的礼品卡代码
	CustomerIP     string     // 下单客户IP
	CreatedAt      time.Time  // 创建时间
	UpdatedAt      time.Time  // 更新时间
}

// OrderItem 订单项领域实体。
//
// 存储订单中的商品明细，包含商品快照信息。
type OrderItem struct {
	ID          uint    // 订单项ID
	OrderID     uint    // 所属订单ID
	ProductID   uint    // 商品ID
	ProductName string  // 商品名称（快照）
	SKU         string  // 商品SKU
	Quantity    int     // 数量
	UnitPrice   float64 // 单价
	TotalPrice  float64 // 小计金额
}

// Shipment 配送领域实体。
type Shipment struct {
	ID             uint       // 配送ID
	OrderID        uint       // 所属订单ID
	TrackingNumber string     // 物流追踪号
	ShippingMethod string     // 配送方式
	ShippedAt      *time.Time // 发货时间
	DeliveredAt    *time.Time // 送达时间
	Status         string     // 配送状态：pending/shipped/delivered
	CreatedAt      time.Time  // 创建时间
	UpdatedAt      time.Time  // 更新时间
}

// OrderRepository 订单仓储接口。
//
// 定义订单持久化的核心操作契约。
type OrderRepository interface {
	// Create 创建订单
	Create(ctx context.Context, order *Order) error
	// GetByID 根据ID获取订单
	GetByID(ctx context.Context, id uint) (*Order, error)
	// List 获取订单列表（userID=0 时不过滤用户）
	List(ctx context.Context, userID uint, page, size int) ([]*Order, int64, error)
	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// Delete 删除订单
	Delete(ctx context.Context, id uint) error
	// GetOrderItems 获取订单项列表
	GetOrderItems(ctx context.Context, orderID uint) ([]*OrderItem, error)
	// GetShipment 获取配送信息
	GetShipment(ctx context.Context, orderID, shipmentID uint) (*Shipment, error)
}

// OrderUseCase 订单用例。
//
// 封装订单相关的业务逻辑，包括创建、取消、退款、PDF发票等。
type OrderUseCase struct {
	repo OrderRepository
}

// NewOrderUseCase 创建订单用例。
func NewOrderUseCase(repo OrderRepository) *OrderUseCase {
	return &OrderUseCase{repo: repo}
}

// Create 创建订单。
//
// 生成订单编号并设置初始状态为 pending。
func (uc *OrderUseCase) Create(ctx context.Context, order *Order) (*Order, error) {
	// 生成订单编号：ORD-时间戳
	order.OrderNo = fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	order.Status = "pending"
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// GetByID 根据ID获取订单。
func (uc *OrderUseCase) GetByID(ctx context.Context, id uint) (*Order, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取订单列表。
//
// userID 为 0 时返回所有订单，否则按用户过滤。
func (uc *OrderUseCase) List(ctx context.Context, userID uint, page, size int) ([]*Order, int64, error) {
	return uc.repo.List(ctx, userID, page, size)
}

// CancelOrder 取消订单。
//
// 仅 pending/processing 状态的订单可以取消。
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id uint) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 校验订单状态是否允许取消
	if order.Status != "pending" && order.Status != "processing" {
		return fmt.Errorf("订单状态为 %s，无法取消", order.Status)
	}
	return uc.repo.UpdateStatus(ctx, id, "cancelled")
}

// RefundOrder 退款。
//
// 仅已支付（processing/completed）的订单可以退款。
func (uc *OrderUseCase) RefundOrder(ctx context.Context, id uint) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 校验订单状态是否允许退款
	if order.Status != "processing" && order.Status != "completed" {
		return fmt.Errorf("订单状态为 %s，无法退款", order.Status)
	}
	return uc.repo.UpdateStatus(ctx, id, "refunded")
}

// Delete 删除订单。
func (uc *OrderUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// GetOrderItems 获取订单项列表。
func (uc *OrderUseCase) GetOrderItems(ctx context.Context, orderID uint) ([]*OrderItem, error) {
	return uc.repo.GetOrderItems(ctx, orderID)
}

// GetShipment 获取配送详情。
func (uc *OrderUseCase) GetShipment(ctx context.Context, orderID, shipmentID uint) (*Shipment, error) {
	return uc.repo.GetShipment(ctx, orderID, shipmentID)
}

// Reorder 重新下单。
//
// 根据原订单信息创建新订单，状态重置为 pending。
func (uc *OrderUseCase) Reorder(ctx context.Context, orderID uint) (*Order, error) {
	original, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	newOrder := &Order{
		UserID:         original.UserID,
		SubTotal:       original.SubTotal,
		ShippingTotal:  original.ShippingTotal,
		TaxTotal:       original.TaxTotal,
		OrderTotal:     original.OrderTotal,
		CurrencyCode:   original.CurrencyCode,
		ShippingMethod: original.ShippingMethod,
		PaymentMethod:  original.PaymentMethod,
	}
	return uc.Create(ctx, newOrder)
}

// RePostPayment 重新提交支付。
//
// 仅 pending 状态的订单可以重新提交支付。
func (uc *OrderUseCase) RePostPayment(ctx context.Context, id uint) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != "pending" {
		return fmt.Errorf("订单状态为 %s，无法重新提交支付", order.Status)
	}
	// 实际场景中这里会调用支付网关重新发起支付
	return nil
}

// GetPDFInvoice 获取PDF发票。
//
// 返回订单的PDF发票数据（当前为占位实现）。
func (uc *OrderUseCase) GetPDFInvoice(ctx context.Context, id uint) ([]byte, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 占位：实际场景中会生成 PDF 文件
	_ = order
	return []byte("PDF invoice placeholder for order " + fmt.Sprintf("%d", id)), nil
}

// ============================================================================
// 购物车领域
// ============================================================================

// ShoppingCartItem 购物车项领域实体。
type ShoppingCartItem struct {
	ID          uint      // 购物车项ID
	UserID      uint      // 用户ID
	ProductID   uint      // 商品ID
	ProductName string    // 商品名称
	SKU         string    // 商品SKU
	Quantity    int       // 数量
	UnitPrice   float64   // 单价
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// CartRepository 购物车仓储接口。
type CartRepository interface {
	// GetCart 获取用户购物车
	GetCart(ctx context.Context, userID uint) ([]*ShoppingCartItem, error)
	// AddCartItem 添加购物车项
	AddCartItem(ctx context.Context, item *ShoppingCartItem) error
	// UpdateCartItem 更新购物车项数量
	UpdateCartItem(ctx context.Context, id uint, quantity int) error
	// DeleteCartItem 删除购物车项
	DeleteCartItem(ctx context.Context, id uint) error
	// ClearCart 清空用户购物车
	ClearCart(ctx context.Context, userID uint) error
}

// CartUseCase 购物车用例。
type CartUseCase struct {
	repo CartRepository
}

// NewCartUseCase 创建购物车用例。
func NewCartUseCase(repo CartRepository) *CartUseCase {
	return &CartUseCase{repo: repo}
}

// GetCart 获取用户购物车。
func (uc *CartUseCase) GetCart(ctx context.Context, userID uint) ([]*ShoppingCartItem, error) {
	return uc.repo.GetCart(ctx, userID)
}

// AddToCart 添加商品到购物车。
func (uc *CartUseCase) AddToCart(ctx context.Context, item *ShoppingCartItem) error {
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	return uc.repo.AddCartItem(ctx, item)
}

// UpdateCart 更新购物车项数量。
func (uc *CartUseCase) UpdateCart(ctx context.Context, id uint, quantity int) error {
	return uc.repo.UpdateCartItem(ctx, id, quantity)
}

// ApplyCoupon 应用折扣码。
//
// 实际场景中会校验折扣码有效性并计算折扣金额。
func (uc *CartUseCase) ApplyCoupon(ctx context.Context, userID uint, couponCode string) error {
	// 占位：实际场景中会校验折扣码并更新购物车折扣
	_ = userID
	_ = couponCode
	return nil
}

// ApplyGiftCard 应用礼品卡。
//
// 实际场景中会校验礼品卡有效性并扣除余额。
func (uc *CartUseCase) ApplyGiftCard(ctx context.Context, userID uint, giftCardCode string) error {
	// 占位：实际场景中会校验礼品卡并更新购物车
	_ = userID
	_ = giftCardCode
	return nil
}

// ClearCart 清空用户购物车。
//
// 通常在确认订单后调用。
func (uc *CartUseCase) ClearCart(ctx context.Context, userID uint) error {
	return uc.repo.ClearCart(ctx, userID)
}

// ============================================================================
// 愿望清单领域
// ============================================================================

// WishlistItem 愿望清单项领域实体。
type WishlistItem struct {
	ID          uint      // 愿望清单项ID
	UserID      uint      // 用户ID
	ProductID   uint      // 商品ID
	ProductName string    // 商品名称
	CreatedAt   time.Time // 创建时间
}

// WishlistRepository 愿望清单仓储接口。
type WishlistRepository interface {
	// GetWishlist 获取用户愿望清单
	GetWishlist(ctx context.Context, userID uint) ([]*WishlistItem, error)
	// AddWishlistItem 添加愿望清单项
	AddWishlistItem(ctx context.Context, item *WishlistItem) error
	// DeleteWishlistItem 删除愿望清单项
	DeleteWishlistItem(ctx context.Context, id uint) error
}

// WishlistUseCase 愿望清单用例。
type WishlistUseCase struct {
	repo WishlistRepository
}

// NewWishlistUseCase 创建愿望清单用例。
func NewWishlistUseCase(repo WishlistRepository) *WishlistUseCase {
	return &WishlistUseCase{repo: repo}
}

// GetWishlist 获取用户愿望清单。
func (uc *WishlistUseCase) GetWishlist(ctx context.Context, userID uint) ([]*WishlistItem, error) {
	return uc.repo.GetWishlist(ctx, userID)
}

// AddToWishlist 添加商品到愿望清单。
func (uc *WishlistUseCase) AddToWishlist(ctx context.Context, item *WishlistItem) error {
	item.CreatedAt = time.Now()
	return uc.repo.AddWishlistItem(ctx, item)
}

// ============================================================================
// 退货请求领域
// ============================================================================

// ReturnRequest 退货请求领域实体。
type ReturnRequest struct {
	ID          uint      // 退货请求ID
	OrderID     uint      // 关联订单ID
	UserID      uint      // 用户ID
	Reason      string    // 退货原因
	Status      string    // 退货状态：pending/approved/rejected/completed
	RefundTotal float64   // 退款金额
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// ReturnRequestRepository 退货请求仓储接口。
type ReturnRequestRepository interface {
	// Create 创建退货请求
	Create(ctx context.Context, rr *ReturnRequest) error
	// List 获取退货请求列表
	List(ctx context.Context, userID uint, page, size int) ([]*ReturnRequest, int64, error)
	// GetByID 根据ID获取退货请求
	GetByID(ctx context.Context, id uint) (*ReturnRequest, error)
}

// ReturnRequestUseCase 退货请求用例。
type ReturnRequestUseCase struct {
	repo ReturnRequestRepository
}

// NewReturnRequestUseCase 创建退货请求用例。
func NewReturnRequestUseCase(repo ReturnRequestRepository) *ReturnRequestUseCase {
	return &ReturnRequestUseCase{repo: repo}
}

// Create 创建退货请求。
//
// 初始状态为 pending，退款金额默认为 0。
func (uc *ReturnRequestUseCase) Create(ctx context.Context, rr *ReturnRequest) (*ReturnRequest, error) {
	rr.Status = "pending"
	rr.CreatedAt = time.Now()
	rr.UpdatedAt = time.Now()
	if err := uc.repo.Create(ctx, rr); err != nil {
		return nil, err
	}
	return rr, nil
}

// List 获取退货请求列表。
func (uc *ReturnRequestUseCase) List(ctx context.Context, userID uint, page, size int) ([]*ReturnRequest, int64, error) {
	return uc.repo.List(ctx, userID, page, size)
}

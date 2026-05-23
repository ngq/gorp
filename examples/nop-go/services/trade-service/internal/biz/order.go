// Package biz 包含交易服务的业务逻辑层
// order.go 整合了订单、购物车、心愿单、结账、退换货等核心业务实体与用例
package biz

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// 订单领域实体
// ============================================================================

// Order 订单实体，表示用户的一次购买行为
type Order struct {
	ID            string         `json:"id"`
	UserID        string         `json:"userId"`
	Status        string         `json:"status"` // pending, confirmed, shipped, delivered, cancelled
	TotalAmount   float64        `json:"totalAmount"`
	Currency      string         `json:"currency"`
	ShippingAddr  string         `json:"shippingAddress"`
	PaymentMethod string         `json:"paymentMethod"`
	Items         []OrderItem    `json:"items" gorm:"-"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// OrderItem 订单行项目，对应订单中的单个商品
type OrderItem struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"orderId"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unitPrice"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ============================================================================
// 购物车领域实体
// ============================================================================

// Cart 购物车实体，包含用户当前待结算的商品
type Cart struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	Items     []CartItem `json:"items" gorm:"-"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// CartItem 购物车行项目
type CartItem struct {
	ID        string    `json:"id"`
	CartID    string    `json:"cartId"`
	ProductID string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"addedAt"`
}

// ============================================================================
// 心愿单领域实体
// ============================================================================

// Wishlist 心愿单实体，用户收藏的商品列表
type Wishlist struct {
	ID        string          `json:"id"`
	UserID    string          `json:"userId"`
	Items     []WishlistItem  `json:"items" gorm:"-"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// WishlistItem 心愿单行项目
type WishlistItem struct {
	ID          string    `json:"id"`
	WishlistID  string    `json:"wishlistId"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	AddedAt     time.Time `json:"addedAt"`
}

// ============================================================================
// 退换货领域实体
// ============================================================================

// ReturnRequest 退换货请求实体
type ReturnRequest struct {
	ID          string         `json:"id"`
	OrderID     string         `json:"orderId"`
	UserID      string         `json:"userId"`
	Reason      string         `json:"reason"`
	Status      string         `json:"status"` // pending, approved, rejected, completed
	RefundAmt   float64        `json:"refundAmount"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ============================================================================
// 仓库接口定义
// ============================================================================

// OrderRepo 订单数据仓库接口
type OrderRepo interface {
	// Create 创建新订单
	Create(ctx context.Context, order *Order) error
	// GetByID 根据ID获取订单
	GetByID(ctx context.Context, id string) (*Order, error)
	// Update 更新订单信息
	Update(ctx context.Context, order *Order) error
	// Delete 软删除订单
	Delete(ctx context.Context, id string) error
	// ListByUserID 按用户ID分页查询订单
	ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*Order, int64, error)
}

// OrderItemRepo 订单行项目数据仓库接口
type OrderItemRepo interface {
	// BatchCreate 批量创建订单行项目
	BatchCreate(ctx context.Context, items []OrderItem) error
	// ListByOrderID 按订单ID查询行项目
	ListByOrderID(ctx context.Context, orderID string) ([]OrderItem, error)
}

// CartRepo 购物车数据仓库接口
type CartRepo interface {
	// GetByUserID 根据用户ID获取购物车
	GetByUserID(ctx context.Context, userID string) (*Cart, error)
	// Create 创建购物车
	Create(ctx context.Context, cart *Cart) error
	// Update 更新购物车
	Update(ctx context.Context, cart *Cart) error
	// Delete 删除购物车
	Delete(ctx context.Context, userID string) error
}

// CartItemRepo 购物车行项目数据仓库接口
type CartItemRepo interface {
	// Create 添加商品到购物车
	Create(ctx context.Context, item *CartItem) error
	// Update 更新购物车商品数量
	Update(ctx context.Context, item *CartItem) error
	// Delete 从购物车移除商品
	Delete(ctx context.Context, cartID, productID string) error
	// ListByCartID 按购物车ID查询行项目
	ListByCartID(ctx context.Context, cartID string) ([]CartItem, error)
}

// WishlistRepo 心愿单数据仓库接口
type WishlistRepo interface {
	// GetByUserID 根据用户ID获取心愿单
	GetByUserID(ctx context.Context, userID string) (*Wishlist, error)
	// Create 创建心愿单
	Create(ctx context.Context, wishlist *Wishlist) error
	// Update 更新心愿单
	Update(ctx context.Context, wishlist *Wishlist) error
}

// WishlistItemRepo 心愿单行项目数据仓库接口
type WishlistItemRepo interface {
	// Create 添加商品到心愿单
	Create(ctx context.Context, item *WishlistItem) error
	// Delete 从心愿单移除商品
	Delete(ctx context.Context, wishlistID, productID string) error
	// ListByWishlistID 按心愿单ID查询行项目
	ListByWishlistID(ctx context.Context, wishlistID string) ([]WishlistItem, error)
}

// ReturnRequestRepo 退换货请求数据仓库接口
type ReturnRequestRepo interface {
	// Create 创建退换货请求
	Create(ctx context.Context, req *ReturnRequest) error
	// GetByID 根据ID获取退换货请求
	GetByID(ctx context.Context, id string) (*ReturnRequest, error)
	// Update 更新退换货请求
	Update(ctx context.Context, req *ReturnRequest) error
	// ListByUserID 按用户ID分页查询退换货请求
	ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*ReturnRequest, int64, error)
	// ListByOrderID 按订单ID查询退换货请求
	ListByOrderID(ctx context.Context, orderID string) ([]*ReturnRequest, error)
}

// ============================================================================
// 用例（UseCase）
// ============================================================================

// OrderUseCase 订单业务用例，封装订单的增删改查逻辑
type OrderUseCase struct {
	orderRepo     OrderRepo
	orderItemRepo OrderItemRepo
}

// NewOrderUseCase 创建订单用例实例
func NewOrderUseCase(orderRepo OrderRepo, orderItemRepo OrderItemRepo) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
	}
}

// CreateOrder 创建订单
func (uc *OrderUseCase) CreateOrder(ctx context.Context, order *Order) error {
	return uc.orderRepo.Create(ctx, order)
}

// GetOrder 获取订单详情
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}
	// 加载订单行项目
	items, err := uc.orderItemRepo.ListByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取订单行项目失败: %w", err)
	}
	order.Items = items
	return order, nil
}

// UpdateOrder 更新订单
func (uc *OrderUseCase) UpdateOrder(ctx context.Context, order *Order) error {
	return uc.orderRepo.Update(ctx, order)
}

// DeleteOrder 删除订单
func (uc *OrderUseCase) DeleteOrder(ctx context.Context, id string) error {
	return uc.orderRepo.Delete(ctx, id)
}

// ListOrders 按用户ID分页查询订单
func (uc *OrderUseCase) ListOrders(ctx context.Context, userID string, page, pageSize int) ([]*Order, int64, error) {
	return uc.orderRepo.ListByUserID(ctx, userID, page, pageSize)
}

// CartUseCase 购物车业务用例
type CartUseCase struct {
	cartRepo     CartRepo
	cartItemRepo CartItemRepo
}

// NewCartUseCase 创建购物车用例实例
func NewCartUseCase(cartRepo CartRepo, cartItemRepo CartItemRepo) *CartUseCase {
	return &CartUseCase{
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
	}
}

// GetCart 获取用户购物车（含行项目）
func (uc *CartUseCase) GetCart(ctx context.Context, userID string) (*Cart, error) {
	cart, err := uc.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取购物车失败: %w", err)
	}
	if cart == nil {
		return nil, nil
	}
	// 加载购物车行项目
	items, err := uc.cartItemRepo.ListByCartID(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("获取购物车行项目失败: %w", err)
	}
	cart.Items = items
	return cart, nil
}

// AddToCart 添加商品到购物车
func (uc *CartUseCase) AddToCart(ctx context.Context, item *CartItem) error {
	return uc.cartItemRepo.Create(ctx, item)
}

// UpdateCartItem 更新购物车商品数量
func (uc *CartUseCase) UpdateCartItem(ctx context.Context, item *CartItem) error {
	return uc.cartItemRepo.Update(ctx, item)
}

// RemoveFromCart 从购物车移除商品
func (uc *CartUseCase) RemoveFromCart(ctx context.Context, cartID, productID string) error {
	return uc.cartItemRepo.Delete(ctx, cartID, productID)
}

// WishlistUseCase 心愿单业务用例
type WishlistUseCase struct {
	wishlistRepo     WishlistRepo
	wishlistItemRepo WishlistItemRepo
}

// NewWishlistUseCase 创建心愿单用例实例
func NewWishlistUseCase(wishlistRepo WishlistRepo, wishlistItemRepo WishlistItemRepo) *WishlistUseCase {
	return &WishlistUseCase{
		wishlistRepo:     wishlistRepo,
		wishlistItemRepo: wishlistItemRepo,
	}
}

// GetWishlist 获取用户心愿单（含行项目）
func (uc *WishlistUseCase) GetWishlist(ctx context.Context, userID string) (*Wishlist, error) {
	wishlist, err := uc.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取心愿单失败: %w", err)
	}
	if wishlist == nil {
		return nil, nil
	}
	// 加载心愿单行项目
	items, err := uc.wishlistItemRepo.ListByWishlistID(ctx, wishlist.ID)
	if err != nil {
		return nil, fmt.Errorf("获取心愿单行项目失败: %w", err)
	}
	wishlist.Items = items
	return wishlist, nil
}

// AddToWishlist 添加商品到心愿单
func (uc *WishlistUseCase) AddToWishlist(ctx context.Context, item *WishlistItem) error {
	return uc.wishlistItemRepo.Create(ctx, item)
}

// RemoveFromWishlist 从心愿单移除商品
func (uc *WishlistUseCase) RemoveFromWishlist(ctx context.Context, wishlistID, productID string) error {
	return uc.wishlistItemRepo.Delete(ctx, wishlistID, productID)
}

// ReturnRequestUseCase 退换货请求业务用例
type ReturnRequestUseCase struct {
	returnRequestRepo ReturnRequestRepo
}

// NewReturnRequestUseCase 创建退换货请求用例实例
func NewReturnRequestUseCase(returnRequestRepo ReturnRequestRepo) *ReturnRequestUseCase {
	return &ReturnRequestUseCase{
		returnRequestRepo: returnRequestRepo,
	}
}

// CreateReturnRequest 创建退换货请求
func (uc *ReturnRequestUseCase) CreateReturnRequest(ctx context.Context, req *ReturnRequest) error {
	return uc.returnRequestRepo.Create(ctx, req)
}

// GetReturnRequest 获取退换货请求详情
func (uc *ReturnRequestUseCase) GetReturnRequest(ctx context.Context, id string) (*ReturnRequest, error) {
	return uc.returnRequestRepo.GetByID(ctx, id)
}

// UpdateReturnRequest 更新退换货请求
func (uc *ReturnRequestUseCase) UpdateReturnRequest(ctx context.Context, req *ReturnRequest) error {
	return uc.returnRequestRepo.Update(ctx, req)
}

// ListReturnRequests 按用户ID分页查询退换货请求
func (uc *ReturnRequestUseCase) ListReturnRequests(ctx context.Context, userID string, page, pageSize int) ([]*ReturnRequest, int64, error) {
	return uc.returnRequestRepo.ListByUserID(ctx, userID, page, pageSize)
}

// ListReturnRequestsByOrder 按订单ID查询退换货请求
func (uc *ReturnRequestUseCase) ListReturnRequestsByOrder(ctx context.Context, orderID string) ([]*ReturnRequest, error) {
	return uc.returnRequestRepo.ListByOrderID(ctx, orderID)
}

// ============================================================================
// CheckoutService 结账服务（组合 OrderUseCase + CartUseCase）
// 结账流程：购物车 → 创建订单 → 清空购物车
// ============================================================================

// CheckoutService 结账服务，封装从购物车到订单的完整结账流程
type CheckoutService struct {
	orderUC *OrderUseCase
	cartUC  *CartUseCase
}

// NewCheckoutService 创建结账服务实例
func NewCheckoutService(orderUC *OrderUseCase, cartUC *CartUseCase) *CheckoutService {
	return &CheckoutService{
		orderUC: orderUC,
		cartUC:  cartUC,
	}
}

// Checkout 执行结账流程：验证购物车 → 创建订单 → 清空购物车
func (s *CheckoutService) Checkout(ctx context.Context, userID string, shippingAddr, paymentMethod string) (*Order, error) {
	// 1. 获取用户购物车
	cart, err := s.cartUC.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("结账-获取购物车失败: %w", err)
	}
	if cart == nil || len(cart.Items) == 0 {
		return nil, fmt.Errorf("购物车为空，无法结账")
	}

	// 2. 从购物车构建订单
	var totalAmount float64
	var orderItems []OrderItem
	for _, ci := range cart.Items {
		subtotal := float64(ci.Quantity) * 0.0 // 实际应查询商品价格，此处为示例
		orderItems = append(orderItems, OrderItem{
			ProductID: ci.ProductID,
			Quantity:  ci.Quantity,
			Subtotal:  subtotal,
		})
		totalAmount += subtotal
	}

	order := &Order{
		UserID:        userID,
		Status:        "pending",
		TotalAmount:   totalAmount,
		ShippingAddr:  shippingAddr,
		PaymentMethod: paymentMethod,
		Items:         orderItems,
	}

	// 3. 创建订单
	if err := s.orderUC.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("结账-创建订单失败: %w", err)
	}

	// 4. 清空购物车
	if err := s.cartUC.RemoveFromCart(ctx, cart.ID, ""); err != nil {
		// 注意：清空购物车失败不影响订单已创建，但需要日志记录
		_ = err
	}

	return order, nil
}

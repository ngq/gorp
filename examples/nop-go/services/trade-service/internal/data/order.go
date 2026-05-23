// Package data 包含交易服务的数据访问层
// order.go 定义订单相关 PO 及仓库实现
package data

import (
	"context"
	"fmt"
	"strconv"

	"nop-go/services/trade-service/internal/biz"

	"gorm.io/gorm"
)

// ============================================================================
// PO 定义
// ============================================================================

// OrderPO 订单持久化对象
type OrderPO struct {
	gorm.Model
	UserID        string  `gorm:"index;not null;size:64;column:user_id" db:"user_id"`
	Status        string  `gorm:"size:32;default:'pending';column:status" db:"status"`
	TotalAmount   float64 `gorm:"type:decimal(12,2);not null;column:total_amount" db:"total_amount"`
	Currency      string  `gorm:"size:8;default:'CNY';column:currency" db:"currency"`
	ShippingAddr  string  `gorm:"size:512;column:shipping_address" db:"shipping_address"`
	PaymentMethod string  `gorm:"size:128;column:payment_method" db:"payment_method"`
}

// TableName 指定订单表名
func (OrderPO) TableName() string { return "orders" }

// ToEntity 转换为订单领域实体
func (po *OrderPO) ToEntity() *biz.Order {
	return &biz.Order{
		ID:            fmtID(po.ID),
		UserID:        po.UserID,
		Status:        po.Status,
		TotalAmount:   po.TotalAmount,
		Currency:      po.Currency,
		ShippingAddr:  po.ShippingAddr,
		PaymentMethod: po.PaymentMethod,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}
}

// OrderItemPO 订单行项目持久化对象
type OrderItemPO struct {
	gorm.Model
	OrderID     string  `gorm:"index;not null;size:64;column:order_id" db:"order_id"`
	ProductID   string  `gorm:"not null;size:64;column:product_id" db:"product_id"`
	ProductName string  `gorm:"size:256;column:product_name" db:"product_name"`
	Quantity    int     `gorm:"not null;column:quantity" db:"quantity"`
	UnitPrice   float64 `gorm:"type:decimal(12,2);not null;column:unit_price" db:"unit_price"`
	Subtotal    float64 `gorm:"type:decimal(12,2);not null;column:subtotal" db:"subtotal"`
}

// TableName 指定订单行项目表名
func (OrderItemPO) TableName() string { return "order_items" }

// ToEntity 转换为订单行项目领域实体
func (po *OrderItemPO) ToEntity() biz.OrderItem {
	return biz.OrderItem{
		ID:          fmtID(po.ID),
		OrderID:     po.OrderID,
		ProductID:   po.ProductID,
		ProductName: po.ProductName,
		Quantity:    po.Quantity,
		UnitPrice:   po.UnitPrice,
		Subtotal:    po.Subtotal,
		CreatedAt:   po.CreatedAt,
	}
}

// CartItemPO 购物车行项目持久化对象
type CartItemPO struct {
	gorm.Model
	CartID    string `gorm:"index;not null;size:64;column:cart_id" db:"cart_id"`
	ProductID string `gorm:"not null;size:64;column:product_id" db:"product_id"`
	Quantity  int    `gorm:"not null;default:1;column:quantity" db:"quantity"`
}

// TableName 指定购物车行项目表名
func (CartItemPO) TableName() string { return "cart_items" }

// ToEntity 转换为购物车行项目领域实体
func (po *CartItemPO) ToEntity() biz.CartItem {
	return biz.CartItem{
		ID:        fmtID(po.ID),
		CartID:    po.CartID,
		ProductID: po.ProductID,
		Quantity:  po.Quantity,
		AddedAt:   po.CreatedAt,
	}
}

// WishlistItemPO 心愿单行项目持久化对象
type WishlistItemPO struct {
	gorm.Model
	WishlistID string `gorm:"index;not null;size:64;column:wishlist_id" db:"wishlist_id"`
	ProductID  string `gorm:"not null;size:64;column:product_id" db:"product_id"`
	ProductName string `gorm:"size:256;not null;column:product_name" db:"product_name"`
}

// TableName 指定心愿单行项目表名
func (WishlistItemPO) TableName() string { return "wishlist_items" }

// ToEntity 转换为心愿单行项目领域实体
func (po *WishlistItemPO) ToEntity() biz.WishlistItem {
	return biz.WishlistItem{
		ID:          fmtID(po.ID),
		WishlistID:  po.WishlistID,
		ProductID:   po.ProductID,
		ProductName: po.ProductName,
		AddedAt:     po.CreatedAt,
	}
}

// ReturnRequestPO 退换货请求持久化对象
type ReturnRequestPO struct {
	gorm.Model
	OrderID   string  `gorm:"index;not null;size:64;column:order_id" db:"order_id"`
	UserID    string  `gorm:"index;not null;size:64;column:user_id" db:"user_id"`
	Reason    string  `gorm:"size:512;column:reason" db:"reason"`
	Status    string  `gorm:"size:32;default:'pending';column:status" db:"status"`
	RefundAmt float64 `gorm:"type:decimal(12,2);default:0;column:refund_amount" db:"refund_amount"`
}

// TableName 指定退换货请求表名
func (ReturnRequestPO) TableName() string { return "return_requests" }

// ToEntity 转换为退换货请求领域实体
func (po *ReturnRequestPO) ToEntity() *biz.ReturnRequest {
	return &biz.ReturnRequest{
		ID:        fmtID(po.ID),
		OrderID:   po.OrderID,
		UserID:    po.UserID,
		Reason:    po.Reason,
		Status:    po.Status,
		RefundAmt: po.RefundAmt,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// ============================================================================
// 仓库实现
// ============================================================================

// --- 订单 ---

type orderRepo struct{ db *gorm.DB }

// NewOrderRepo 创建订单仓储
func NewOrderRepo(db *gorm.DB) biz.OrderRepo { return &orderRepo{db: db} }

func (r *orderRepo) Create(ctx context.Context, order *biz.Order) error {
	po := &OrderPO{
		UserID:        order.UserID,
		Status:        order.Status,
		TotalAmount:   order.TotalAmount,
		Currency:      order.Currency,
		ShippingAddr:  order.ShippingAddr,
		PaymentMethod: order.PaymentMethod,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	order.ID = fmtID(po.ID)
	return nil
}

func (r *orderRepo) GetByID(ctx context.Context, id string) (*biz.Order, error) {
	var po OrderPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

func (r *orderRepo) Update(ctx context.Context, order *biz.Order) error {
	return r.db.WithContext(ctx).Model(&OrderPO{}).Where("id = ?", parseID(order.ID)).Updates(map[string]interface{}{
		"status":           order.Status,
		"total_amount":     order.TotalAmount,
		"shipping_address": order.ShippingAddr,
		"payment_method":   order.PaymentMethod,
	}).Error
}

func (r *orderRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&OrderPO{}, parseID(id)).Error
}

func (r *orderRepo) ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*biz.Order, int64, error) {
	var pos []OrderPO
	var total int64
	r.db.WithContext(ctx).Model(&OrderPO{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Order, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// --- 订单行项目 ---

type orderItemRepo struct{ db *gorm.DB }

// NewOrderItemRepo 创建订单行项目仓储
func NewOrderItemRepo(db *gorm.DB) biz.OrderItemRepo { return &orderItemRepo{db: db} }

func (r *orderItemRepo) BatchCreate(ctx context.Context, items []biz.OrderItem) error {
	pos := make([]OrderItemPO, len(items))
	for i, item := range items {
		pos[i] = OrderItemPO{
			OrderID:     item.OrderID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		}
	}
	return r.db.WithContext(ctx).Create(&pos).Error
}

func (r *orderItemRepo) ListByOrderID(ctx context.Context, orderID string) ([]biz.OrderItem, error) {
	var pos []OrderItemPO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]biz.OrderItem, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// --- 购物车 ---

type cartRepo struct{ db *gorm.DB }

// NewCartRepo 创建购物车仓储
func NewCartRepo(db *gorm.DB) biz.CartRepo { return &cartRepo{db: db} }

func (r *cartRepo) GetByUserID(ctx context.Context, userID string) (*biz.Cart, error) {
	// 购物车通过虚拟 ID 关联用户，无需单独存储
	return &biz.Cart{ID: "cart-" + userID, UserID: userID}, nil
}

func (r *cartRepo) Create(ctx context.Context, cart *biz.Cart) error { return nil }
func (r *cartRepo) Update(ctx context.Context, cart *biz.Cart) error { return nil }
func (r *cartRepo) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("cart_id LIKE ?", "cart-"+userID+"%").Delete(&CartItemPO{}).Error
}

// --- 购物车行项目 ---

type cartItemRepo struct{ db *gorm.DB }

// NewCartItemRepo 创建购物车行项目仓储
func NewCartItemRepo(db *gorm.DB) biz.CartItemRepo { return &cartItemRepo{db: db} }

func (r *cartItemRepo) Create(ctx context.Context, item *biz.CartItem) error {
	po := &CartItemPO{
		CartID:    item.CartID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	item.ID = fmtID(po.ID)
	return nil
}

func (r *cartItemRepo) Update(ctx context.Context, item *biz.CartItem) error {
	return r.db.WithContext(ctx).Model(&CartItemPO{}).Where("id = ?", parseID(item.ID)).Update("quantity", item.Quantity).Error
}

func (r *cartItemRepo) Delete(ctx context.Context, cartID, productID string) error {
	if productID != "" {
		return r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&CartItemPO{}).Error
	}
	// productID 为空时清空整个购物车
	return r.db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&CartItemPO{}).Error
}

func (r *cartItemRepo) ListByCartID(ctx context.Context, cartID string) ([]biz.CartItem, error) {
	var pos []CartItemPO
	if err := r.db.WithContext(ctx).Where("cart_id = ?", cartID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]biz.CartItem, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// --- 心愿单 ---

type wishlistRepo struct{ db *gorm.DB }

// NewWishlistRepo 创建心愿单仓储
func NewWishlistRepo(db *gorm.DB) biz.WishlistRepo { return &wishlistRepo{db: db} }

func (r *wishlistRepo) GetByUserID(ctx context.Context, userID string) (*biz.Wishlist, error) {
	return &biz.Wishlist{ID: "wl-" + userID, UserID: userID}, nil
}

func (r *wishlistRepo) Create(ctx context.Context, wishlist *biz.Wishlist) error { return nil }
func (r *wishlistRepo) Update(ctx context.Context, wishlist *biz.Wishlist) error { return nil }

// --- 心愿单行项目 ---

type wishlistItemRepo struct{ db *gorm.DB }

// NewWishlistItemRepo 创建心愿单行项目仓储
func NewWishlistItemRepo(db *gorm.DB) biz.WishlistItemRepo { return &wishlistItemRepo{db: db} }

func (r *wishlistItemRepo) Create(ctx context.Context, item *biz.WishlistItem) error {
	po := &WishlistItemPO{
		WishlistID:  item.WishlistID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	item.ID = fmtID(po.ID)
	return nil
}

func (r *wishlistItemRepo) Delete(ctx context.Context, wishlistID, productID string) error {
	return r.db.WithContext(ctx).Where("wishlist_id = ? AND product_id = ?", wishlistID, productID).Delete(&WishlistItemPO{}).Error
}

func (r *wishlistItemRepo) ListByWishlistID(ctx context.Context, wishlistID string) ([]biz.WishlistItem, error) {
	var pos []WishlistItemPO
	if err := r.db.WithContext(ctx).Where("wishlist_id = ?", wishlistID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]biz.WishlistItem, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// --- 退换货请求 ---

type returnRequestRepo struct{ db *gorm.DB }

// NewReturnRequestRepo 创建退换货请求仓储
func NewReturnRequestRepo(db *gorm.DB) biz.ReturnRequestRepo {
	return &returnRequestRepo{db: db}
}

func (r *returnRequestRepo) Create(ctx context.Context, req *biz.ReturnRequest) error {
	po := &ReturnRequestPO{
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		Reason:    req.Reason,
		Status:    req.Status,
		RefundAmt: req.RefundAmt,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	req.ID = fmtID(po.ID)
	return nil
}

func (r *returnRequestRepo) GetByID(ctx context.Context, id string) (*biz.ReturnRequest, error) {
	var po ReturnRequestPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

func (r *returnRequestRepo) Update(ctx context.Context, req *biz.ReturnRequest) error {
	return r.db.WithContext(ctx).Model(&ReturnRequestPO{}).Where("id = ?", parseID(req.ID)).Updates(map[string]interface{}{
		"status":        req.Status,
		"refund_amount": req.RefundAmt,
	}).Error
}

func (r *returnRequestRepo) ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*biz.ReturnRequest, int64, error) {
	var pos []ReturnRequestPO
	var total int64
	r.db.WithContext(ctx).Model(&ReturnRequestPO{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.ReturnRequest, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

func (r *returnRequestRepo) ListByOrderID(ctx context.Context, orderID string) ([]*biz.ReturnRequest, error) {
	var pos []ReturnRequestPO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.ReturnRequest, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// fmtID 将 uint ID 转换为 string
func fmtID(id uint) string { return fmt.Sprintf("%d", id) }

// parseID 将 string ID 转换为 uint
func parseID(id string) uint {
	n, _ := strconv.ParseUint(id, 10, 64)
	return uint(n)
}

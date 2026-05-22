// Package data 数据访问层。
//
// 本文件包含订单服务所有持久化对象（PO）定义及仓储实现。
// PO 结构体同时包含 gorm 和 db(sqlx) tag，以支持双 ORM 场景。
package data

import (
	"context"
	"time"

	"nop-go/services/order/internal/biz"

	"gorm.io/gorm"
)

// ============================================================================
// 订单相关 PO
// ============================================================================

// OrderPO 订单持久化对象。
//
// 对应数据库表 orders，存储订单主信息。
// 包含订单基础字段、金额、状态及时间戳。
type OrderPO struct {
	ID            uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	OrderNo       string    `gorm:"uniqueIndex;size:64;not null;column:order_no" db:"order_no" json:"order_no"`               // 订单编号，唯一
	UserID        uint      `gorm:"index;not null;column:user_id" db:"user_id" json:"user_id"`                                 // 下单用户ID
	Status        string    `gorm:"size:32;default:'pending';column:status" db:"status" json:"status"`                         // 订单状态：pending/processing/cancelled/refunded/completed
	SubTotal      float64   `gorm:"type:decimal(12,2);column:sub_total" db:"sub_total" json:"sub_total"`                       // 商品小计
	DiscountTotal float64   `gorm:"type:decimal(12,2);default:0;column:discount_total" db:"discount_total" json:"discount_total"` // 折扣总额
	ShippingTotal float64   `gorm:"type:decimal(12,2);default:0;column:shipping_total" db:"shipping_total" json:"shipping_total"` // 运费总额
	TaxTotal      float64   `gorm:"type:decimal(12,2);default:0;column:tax_total" db:"tax_total" json:"tax_total"`             // 税费总额
	OrderTotal    float64   `gorm:"type:decimal(12,2);column:order_total" db:"order_total" json:"order_total"`                 // 订单总额
	CurrencyCode  string    `gorm:"size:8;default:'CNY';column:currency_code" db:"currency_code" json:"currency_code"`         // 货币代码
	BillingAddrID *uint     `gorm:"column:billing_address_id" db:"billing_address_id" json:"billing_address_id"`               // 账单地址ID（可为空）
	ShippingAddrID *uint    `gorm:"column:shipping_address_id" db:"shipping_address_id" json:"shipping_address_id"`            // 配送地址ID（可为空）
	ShippingMethod string   `gorm:"size:128;column:shipping_method" db:"shipping_method" json:"shipping_method"`               // 配送方式
	PaymentMethod string    `gorm:"size:128;column:payment_method" db:"payment_method" json:"payment_method"`                  // 支付方式
	CouponCode    string    `gorm:"size:128;column:coupon_code" db:"coupon_code" json:"coupon_code"`                           // 使用的折扣码
	GiftCardCode  string    `gorm:"size:128;column:gift_card_code" db:"gift_card_code" json:"gift_card_code"`                  // 使用的礼品卡代码
	CustomerIP    string    `gorm:"size:64;column:customer_ip" db:"customer_ip" json:"customer_ip"`                            // 下单客户IP
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`                        // 创建时间
	UpdatedAt     time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`                        // 更新时间
}

// TableName 表名。
func (OrderPO) TableName() string {
	return "orders"
}

// ToEntity 转换为订单领域实体。
func (po *OrderPO) ToEntity() *biz.Order {
	return &biz.Order{
		ID:            po.ID,
		OrderNo:       po.OrderNo,
		UserID:        po.UserID,
		Status:        po.Status,
		SubTotal:      po.SubTotal,
		DiscountTotal: po.DiscountTotal,
		ShippingTotal: po.ShippingTotal,
		TaxTotal:      po.TaxTotal,
		OrderTotal:    po.OrderTotal,
		CurrencyCode:  po.CurrencyCode,
		BillingAddrID: po.BillingAddrID,
		ShippingAddrID: po.ShippingAddrID,
		ShippingMethod: po.ShippingMethod,
		PaymentMethod: po.PaymentMethod,
		CouponCode:    po.CouponCode,
		GiftCardCode:  po.GiftCardCode,
		CustomerIP:    po.CustomerIP,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}
}

// ============================================================================
// 订单项 PO
// ============================================================================

// OrderItemPO 订单项持久化对象。
//
// 对应数据库表 order_items，存储订单中的商品明细。
type OrderItemPO struct {
	ID         uint    `gorm:"primaryKey;column:id" db:"id" json:"id"`
	OrderID    uint    `gorm:"index;not null;column:order_id" db:"order_id" json:"order_id"`           // 所属订单ID
	ProductID  uint    `gorm:"not null;column:product_id" db:"product_id" json:"product_id"`           // 商品ID
	ProductName string `gorm:"size:256;not null;column:product_name" db:"product_name" json:"product_name"` // 商品名称（快照）
	SKU        string  `gorm:"size:128;column:sku" db:"sku" json:"sku"`                                // 商品SKU
	Quantity   int     `gorm:"not null;column:quantity" db:"quantity" json:"quantity"`                 // 数量
	UnitPrice  float64 `gorm:"type:decimal(12,2);not null;column:unit_price" db:"unit_price" json:"unit_price"` // 单价
	TotalPrice float64 `gorm:"type:decimal(12,2);not null;column:total_price" db:"total_price" json:"total_price"` // 小计金额
}

// TableName 表名。
func (OrderItemPO) TableName() string {
	return "order_items"
}

// ToEntity 转换为订单项领域实体。
func (po *OrderItemPO) ToEntity() *biz.OrderItem {
	return &biz.OrderItem{
		ID:          po.ID,
		OrderID:     po.OrderID,
		ProductID:   po.ProductID,
		ProductName: po.ProductName,
		SKU:         po.SKU,
		Quantity:    po.Quantity,
		UnitPrice:   po.UnitPrice,
		TotalPrice:  po.TotalPrice,
	}
}

// ============================================================================
// 配送 PO
// ============================================================================

// ShipmentPO 配送持久化对象。
//
// 对应数据库表 shipments，存储订单的配送信息。
type ShipmentPO struct {
	ID             uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	OrderID        uint      `gorm:"index;not null;column:order_id" db:"order_id" json:"order_id"`                       // 所属订单ID
	TrackingNumber string    `gorm:"size:128;column:tracking_number" db:"tracking_number" json:"tracking_number"`         // 物流追踪号
	ShippingMethod string    `gorm:"size:128;column:shipping_method" db:"shipping_method" json:"shipping_method"`         // 配送方式
	ShippedAt      *time.Time `gorm:"column:shipped_at" db:"shipped_at" json:"shipped_at"`                               // 发货时间
	DeliveredAt    *time.Time `gorm:"column:delivered_at" db:"delivered_at" json:"delivered_at"`                           // 送达时间
	Status         string    `gorm:"size:32;default:'pending';column:status" db:"status" json:"status"`                   // 配送状态：pending/shipped/delivered
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`                  // 创建时间
	UpdatedAt      time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`                  // 更新时间
}

// TableName 表名。
func (ShipmentPO) TableName() string {
	return "shipments"
}

// ToEntity 转换为配送领域实体。
func (po *ShipmentPO) ToEntity() *biz.Shipment {
	return &biz.Shipment{
		ID:             po.ID,
		OrderID:        po.OrderID,
		TrackingNumber: po.TrackingNumber,
		ShippingMethod: po.ShippingMethod,
		ShippedAt:      po.ShippedAt,
		DeliveredAt:    po.DeliveredAt,
		Status:         po.Status,
		CreatedAt:      po.CreatedAt,
		UpdatedAt:      po.UpdatedAt,
	}
}

// ============================================================================
// 购物车 PO
// ============================================================================

// ShoppingCartItemPO 购物车项持久化对象。
//
// 对应数据库表 shopping_cart_items，存储用户购物车中的商品。
type ShoppingCartItemPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	UserID      uint      `gorm:"index;not null;column:user_id" db:"user_id" json:"user_id"`                       // 用户ID
	ProductID   uint      `gorm:"not null;column:product_id" db:"product_id" json:"product_id"`                     // 商品ID
	ProductName string    `gorm:"size:256;not null;column:product_name" db:"product_name" json:"product_name"`      // 商品名称
	SKU         string    `gorm:"size:128;column:sku" db:"sku" json:"sku"`                                          // 商品SKU
	Quantity    int       `gorm:"not null;default:1;column:quantity" db:"quantity" json:"quantity"`                 // 数量
	UnitPrice   float64   `gorm:"type:decimal(12,2);not null;column:unit_price" db:"unit_price" json:"unit_price"`  // 单价
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`               // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`               // 更新时间
}

// TableName 表名。
func (ShoppingCartItemPO) TableName() string {
	return "shopping_cart_items"
}

// ToEntity 转换为购物车项领域实体。
func (po *ShoppingCartItemPO) ToEntity() *biz.ShoppingCartItem {
	return &biz.ShoppingCartItem{
		ID:          po.ID,
		UserID:      po.UserID,
		ProductID:   po.ProductID,
		ProductName: po.ProductName,
		SKU:         po.SKU,
		Quantity:    po.Quantity,
		UnitPrice:   po.UnitPrice,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// ============================================================================
// 愿望清单 PO
// ============================================================================

// WishlistItemPO 愿望清单项持久化对象。
//
// 对应数据库表 wishlist_items，存储用户愿望清单中的商品。
type WishlistItemPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	UserID      uint      `gorm:"index;not null;column:user_id" db:"user_id" json:"user_id"`                       // 用户ID
	ProductID   uint      `gorm:"not null;column:product_id" db:"product_id" json:"product_id"`                     // 商品ID
	ProductName string    `gorm:"size:256;not null;column:product_name" db:"product_name" json:"product_name"`      // 商品名称
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`               // 创建时间
}

// TableName 表名。
func (WishlistItemPO) TableName() string {
	return "wishlist_items"
}

// ToEntity 转换为愿望清单项领域实体。
func (po *WishlistItemPO) ToEntity() *biz.WishlistItem {
	return &biz.WishlistItem{
		ID:          po.ID,
		UserID:      po.UserID,
		ProductID:   po.ProductID,
		ProductName: po.ProductName,
		CreatedAt:   po.CreatedAt,
	}
}

// ============================================================================
// 退货请求 PO
// ============================================================================

// ReturnRequestPO 退货请求持久化对象。
//
// 对应数据库表 return_requests，存储用户提交的退货申请。
type ReturnRequestPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	OrderID     uint      `gorm:"index;not null;column:order_id" db:"order_id" json:"order_id"`                     // 关联订单ID
	UserID      uint      `gorm:"index;not null;column:user_id" db:"user_id" json:"user_id"`                        // 用户ID
	Reason      string    `gorm:"size:512;column:reason" db:"reason" json:"reason"`                                 // 退货原因
	Status      string    `gorm:"size:32;default:'pending';column:status" db:"status" json:"status"`                 // 退货状态：pending/approved/rejected/completed
	RefundTotal float64   `gorm:"type:decimal(12,2);default:0;column:refund_total" db:"refund_total" json:"refund_total"` // 退款金额
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`                // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`                // 更新时间
}

// TableName 表名。
func (ReturnRequestPO) TableName() string {
	return "return_requests"
}

// ToEntity 转换为退货请求领域实体。
func (po *ReturnRequestPO) ToEntity() *biz.ReturnRequest {
	return &biz.ReturnRequest{
		ID:          po.ID,
		OrderID:     po.OrderID,
		UserID:      po.UserID,
		Reason:      po.Reason,
		Status:      po.Status,
		RefundTotal: po.RefundTotal,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// ============================================================================
// 仓储实现
// ============================================================================

// orderRepo 订单仓储实现。
//
// 基于 gorm.DB 实现 biz.OrderRepository 接口。
type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建订单仓储。
func NewOrderRepo(db *gorm.DB) biz.OrderRepository {
	return &orderRepo{db: db}
}

// Create 创建订单。
func (r *orderRepo) Create(ctx context.Context, order *biz.Order) error {
	po := &OrderPO{
		OrderNo:       order.OrderNo,
		UserID:        order.UserID,
		Status:        order.Status,
		SubTotal:      order.SubTotal,
		DiscountTotal: order.DiscountTotal,
		ShippingTotal: order.ShippingTotal,
		TaxTotal:      order.TaxTotal,
		OrderTotal:    order.OrderTotal,
		CurrencyCode:  order.CurrencyCode,
		BillingAddrID: order.BillingAddrID,
		ShippingAddrID: order.ShippingAddrID,
		ShippingMethod: order.ShippingMethod,
		PaymentMethod: order.PaymentMethod,
		CouponCode:    order.CouponCode,
		GiftCardCode:  order.GiftCardCode,
		CustomerIP:    order.CustomerIP,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取订单。
func (r *orderRepo) GetByID(ctx context.Context, id uint) (*biz.Order, error) {
	var po OrderPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取订单列表。
//
// 支持按用户ID过滤和分页查询。
func (r *orderRepo) List(ctx context.Context, userID uint, page, size int) ([]*biz.Order, int64, error) {
	var pos []OrderPO
	var total int64

	db := r.db.WithContext(ctx).Model(&OrderPO{})
	// 按用户ID过滤（0 表示不过滤）
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	db.Count(&total)

	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*biz.Order, len(pos))
	for i, po := range pos {
		orders[i] = po.ToEntity()
	}
	return orders, total, nil
}

// UpdateStatus 更新订单状态。
func (r *orderRepo) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&OrderPO{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除订单。
func (r *orderRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&OrderPO{}, id).Error
}

// GetOrderItems 获取订单项列表。
func (r *orderRepo) GetOrderItems(ctx context.Context, orderID uint) ([]*biz.OrderItem, error) {
	var pos []OrderItemPO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.OrderItem, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// GetShipment 获取配送信息。
func (r *orderRepo) GetShipment(ctx context.Context, orderID, shipmentID uint) (*biz.Shipment, error) {
	var po ShipmentPO
	if err := r.db.WithContext(ctx).Where("id = ? AND order_id = ?", shipmentID, orderID).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// ============================================================================
// 购物车仓储实现
// ============================================================================

// cartRepo 购物车仓储实现。
type cartRepo struct {
	db *gorm.DB
}

// NewCartRepo 创建购物车仓储。
func NewCartRepo(db *gorm.DB) biz.CartRepository {
	return &cartRepo{db: db}
}

// GetCart 获取用户购物车。
func (r *cartRepo) GetCart(ctx context.Context, userID uint) ([]*biz.ShoppingCartItem, error) {
	var pos []ShoppingCartItemPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.ShoppingCartItem, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// AddCartItem 添加购物车项。
func (r *cartRepo) AddCartItem(ctx context.Context, item *biz.ShoppingCartItem) error {
	po := &ShoppingCartItemPO{
		UserID:      item.UserID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		SKU:         item.SKU,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// UpdateCartItem 更新购物车项数量。
func (r *cartRepo) UpdateCartItem(ctx context.Context, id uint, quantity int) error {
	return r.db.WithContext(ctx).Model(&ShoppingCartItemPO{}).Where("id = ?", id).Update("quantity", quantity).Error
}

// DeleteCartItem 删除购物车项。
func (r *cartRepo) DeleteCartItem(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&ShoppingCartItemPO{}, id).Error
}

// ClearCart 清空用户购物车。
func (r *cartRepo) ClearCart(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&ShoppingCartItemPO{}).Error
}

// ============================================================================
// 愿望清单仓储实现
// ============================================================================

// wishlistRepo 愿望清单仓储实现。
type wishlistRepo struct {
	db *gorm.DB
}

// NewWishlistRepo 创建愿望清单仓储。
func NewWishlistRepo(db *gorm.DB) biz.WishlistRepository {
	return &wishlistRepo{db: db}
}

// GetWishlist 获取用户愿望清单。
func (r *wishlistRepo) GetWishlist(ctx context.Context, userID uint) ([]*biz.WishlistItem, error) {
	var pos []WishlistItemPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.WishlistItem, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, nil
}

// AddWishlistItem 添加愿望清单项。
func (r *wishlistRepo) AddWishlistItem(ctx context.Context, item *biz.WishlistItem) error {
	po := &WishlistItemPO{
		UserID:      item.UserID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// DeleteWishlistItem 删除愿望清单项。
func (r *wishlistRepo) DeleteWishlistItem(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&WishlistItemPO{}, id).Error
}

// ============================================================================
// 退货请求仓储实现
// ============================================================================

// returnRequestRepo 退货请求仓储实现。
type returnRequestRepo struct {
	db *gorm.DB
}

// NewReturnRequestRepo 创建退货请求仓储。
func NewReturnRequestRepo(db *gorm.DB) biz.ReturnRequestRepository {
	return &returnRequestRepo{db: db}
}

// Create 创建退货请求。
func (r *returnRequestRepo) Create(ctx context.Context, rr *biz.ReturnRequest) error {
	po := &ReturnRequestPO{
		OrderID:     rr.OrderID,
		UserID:      rr.UserID,
		Reason:      rr.Reason,
		Status:      rr.Status,
		RefundTotal: rr.RefundTotal,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// List 获取退货请求列表。
func (r *returnRequestRepo) List(ctx context.Context, userID uint, page, size int) ([]*biz.ReturnRequest, int64, error) {
	var pos []ReturnRequestPO
	var total int64

	db := r.db.WithContext(ctx).Model(&ReturnRequestPO{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	db.Count(&total)

	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.ReturnRequest, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// GetByID 根据ID获取退货请求。
func (r *returnRequestRepo) GetByID(ctx context.Context, id uint) (*biz.ReturnRequest, error) {
	var po ReturnRequestPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

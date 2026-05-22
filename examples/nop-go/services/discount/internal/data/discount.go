// Package data 提供 discount 服务的数据访问层
//
// 包含五张表的 PO 定义与仓储实现：
// 1. discounts — 折扣主体
// 2. discount_products — 折扣关联商品
// 3. discount_categories — 折扣关联分类
// 4. discount_manufacturers — 折扣关联制造商
// 5. discount_usage_history — 折扣使用历史
package data

import (
	"context"
	"time"

	"nop-go/services/discount/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================

// DiscountPO 折扣持久化对象
// 对应数据库表 discounts
type DiscountPO struct {
	ID                uint      `gorm:"column:id;primaryKey" db:"id"`                                   // 主键 ID
	Name              string    `gorm:"column:name;size:256" db:"name"`                                 // 折扣名称
	DiscountType      string    `gorm:"column:discount_type;size:32" db:"discount_type"`                // 折扣类型
	DiscountAmount    float64   `gorm:"column:discount_amount;type:decimal(10,2)" db:"discount_amount"` // 折扣金额/百分比
	StartDate         time.Time `gorm:"column:start_date" db:"start_date"`                              // 折扣开始日期
	EndDate           time.Time `gorm:"column:end_date" db:"end_date"`                                  // 折扣结束日期
	RequiresCouponCode bool     `gorm:"column:requires_coupon_code;default:false" db:"requires_coupon_code"` // 是否需要优惠券码
	CouponCode        string    `gorm:"column:coupon_code;size:128" db:"coupon_code"`                   // 优惠券码
	IsCumulative      bool      `gorm:"column:is_cumulative;default:false" db:"is_cumulative"`          // 是否可叠加使用
	DisplayOrder      int       `gorm:"column:display_order" db:"display_order"`                        // 显示排序
	IsActive          bool      `gorm:"column:is_active;default:true" db:"is_active"`                   // 是否启用
	LimitationTimes   int       `gorm:"column:limitation_times;default:0" db:"limitation_times"`        // 使用次数限制
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`               // 创建时间
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`               // 更新时间
}

// TableName 指定折扣表名
func (DiscountPO) TableName() string { return "discounts" }

// ToEntity 转换为折扣领域实体
func (po *DiscountPO) ToEntity() *biz.Discount {
	return &biz.Discount{
		ID: po.ID, Name: po.Name, DiscountType: po.DiscountType,
		DiscountAmount: po.DiscountAmount, StartDate: po.StartDate, EndDate: po.EndDate,
		RequiresCouponCode: po.RequiresCouponCode, CouponCode: po.CouponCode,
		IsCumulative: po.IsCumulative, DisplayOrder: po.DisplayOrder,
		IsActive: po.IsActive, LimitationTimes: po.LimitationTimes,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// DiscountProductPO 折扣关联商品持久化对象
// 对应数据库表 discount_products
type DiscountProductPO struct {
	ID          uint      `gorm:"column:id;primaryKey" db:"id"`                     // 主键 ID
	DiscountID  uint      `gorm:"column:discount_id;index" db:"discount_id"`        // 折扣 ID
	ProductID   uint      `gorm:"column:product_id;index" db:"product_id"`          // 商品 ID
	ProductName string    `gorm:"column:product_name;size:256" db:"product_name"`   // 商品名称（冗余展示）
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"` // 创建时间
}

// TableName 指定折扣关联商品表名
func (DiscountProductPO) TableName() string { return "discount_products" }

// ToEntity 转换为折扣关联商品领域实体
func (po *DiscountProductPO) ToEntity() *biz.DiscountProduct {
	return &biz.DiscountProduct{
		ID: po.ID, DiscountID: po.DiscountID, ProductID: po.ProductID,
		ProductName: po.ProductName, CreatedAt: po.CreatedAt,
	}
}

// DiscountCategoryPO 折扣关联分类持久化对象
// 对应数据库表 discount_categories
type DiscountCategoryPO struct {
	ID           uint      `gorm:"column:id;primaryKey" db:"id"`                      // 主键 ID
	DiscountID   uint      `gorm:"column:discount_id;index" db:"discount_id"`         // 折扣 ID
	CategoryID   uint      `gorm:"column:category_id;index" db:"category_id"`         // 分类 ID
	CategoryName string    `gorm:"column:category_name;size:256" db:"category_name"`  // 分类名称（冗余展示）
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`  // 创建时间
}

// TableName 指定折扣关联分类表名
func (DiscountCategoryPO) TableName() string { return "discount_categories" }

// ToEntity 转换为折扣关联分类领域实体
func (po *DiscountCategoryPO) ToEntity() *biz.DiscountCategory {
	return &biz.DiscountCategory{
		ID: po.ID, DiscountID: po.DiscountID, CategoryID: po.CategoryID,
		CategoryName: po.CategoryName, CreatedAt: po.CreatedAt,
	}
}

// DiscountManufacturerPO 折扣关联制造商持久化对象
// 对应数据库表 discount_manufacturers
type DiscountManufacturerPO struct {
	ID               uint      `gorm:"column:id;primaryKey" db:"id"`                            // 主键 ID
	DiscountID       uint      `gorm:"column:discount_id;index" db:"discount_id"`               // 折扣 ID
	ManufacturerID   uint      `gorm:"column:manufacturer_id;index" db:"manufacturer_id"`       // 制造商 ID
	ManufacturerName string    `gorm:"column:manufacturer_name;size:256" db:"manufacturer_name"` // 制造商名称（冗余展示）
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`        // 创建时间
}

// TableName 指定折扣关联制造商表名
func (DiscountManufacturerPO) TableName() string { return "discount_manufacturers" }

// ToEntity 转换为折扣关联制造商领域实体
func (po *DiscountManufacturerPO) ToEntity() *biz.DiscountManufacturer {
	return &biz.DiscountManufacturer{
		ID: po.ID, DiscountID: po.DiscountID, ManufacturerID: po.ManufacturerID,
		ManufacturerName: po.ManufacturerName, CreatedAt: po.CreatedAt,
	}
}

// DiscountUsageHistoryPO 折扣使用历史持久化对象
// 对应数据库表 discount_usage_history
type DiscountUsageHistoryPO struct {
	ID           uint      `gorm:"column:id;primaryKey" db:"id"`                      // 主键 ID
	DiscountID   uint      `gorm:"column:discount_id;index" db:"discount_id"`         // 折扣 ID
	OrderID      uint      `gorm:"column:order_id;index" db:"order_id"`               // 订单 ID
	CustomerID   uint      `gorm:"column:customer_id;index" db:"customer_id"`         // 客户 ID
	CustomerName string    `gorm:"column:customer_name;size:256" db:"customer_name"`  // 客户名称（冗余展示）
	CouponCode   string    `gorm:"column:coupon_code;size:128" db:"coupon_code"`      // 使用的优惠券码
	UsedOn       time.Time `gorm:"column:used_on" db:"used_on"`                       // 使用日期
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`  // 创建时间
}

// TableName 指定折扣使用历史表名
func (DiscountUsageHistoryPO) TableName() string { return "discount_usage_history" }

// ToEntity 转换为折扣使用历史领域实体
func (po *DiscountUsageHistoryPO) ToEntity() *biz.DiscountUsageHistory {
	return &biz.DiscountUsageHistory{
		ID: po.ID, DiscountID: po.DiscountID, OrderID: po.OrderID,
		CustomerID: po.CustomerID, CustomerName: po.CustomerName,
		CouponCode: po.CouponCode, UsedOn: po.UsedOn, CreatedAt: po.CreatedAt,
	}
}

// ==================== 仓储实现 ====================

// discountRepo 折扣仓储实现
type discountRepo struct {
	db *gorm.DB
}

// NewDiscountRepo 创建折扣仓储
func NewDiscountRepo(db *gorm.DB) biz.DiscountRepository {
	return &discountRepo{db: db}
}

// List 获取折扣列表（分页）
func (r *discountRepo) List(ctx context.Context, page, pageSize int) ([]*biz.Discount, int64, error) {
	var pos []DiscountPO
	var total int64

	r.db.WithContext(ctx).Model(&DiscountPO{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Discount, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// Create 创建折扣
func (r *discountRepo) Create(ctx context.Context, discount *biz.Discount) (*biz.Discount, error) {
	po := &DiscountPO{
		Name: discount.Name, DiscountType: discount.DiscountType,
		DiscountAmount: discount.DiscountAmount, StartDate: discount.StartDate, EndDate: discount.EndDate,
		RequiresCouponCode: discount.RequiresCouponCode, CouponCode: discount.CouponCode,
		IsCumulative: discount.IsCumulative, DisplayOrder: discount.DisplayOrder,
		IsActive: discount.IsActive, LimitationTimes: discount.LimitationTimes,
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新折扣
func (r *discountRepo) Update(ctx context.Context, discount *biz.Discount) (*biz.Discount, error) {
	po := &DiscountPO{
		ID: discount.ID, Name: discount.Name, DiscountType: discount.DiscountType,
		DiscountAmount: discount.DiscountAmount, StartDate: discount.StartDate, EndDate: discount.EndDate,
		RequiresCouponCode: discount.RequiresCouponCode, CouponCode: discount.CouponCode,
		IsCumulative: discount.IsCumulative, DisplayOrder: discount.DisplayOrder,
		IsActive: discount.IsActive, LimitationTimes: discount.LimitationTimes,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Delete 删除折扣
func (r *discountRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&DiscountPO{}, id).Error
}

// discountProductRepo 折扣关联商品仓储实现
type discountProductRepo struct {
	db *gorm.DB
}

// NewDiscountProductRepo 创建折扣关联商品仓储
func NewDiscountProductRepo(db *gorm.DB) biz.DiscountProductRepository {
	return &discountProductRepo{db: db}
}

// ListByDiscountID 根据折扣 ID 获取关联商品列表（分页）
func (r *discountProductRepo) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountProduct, int64, error) {
	var pos []DiscountProductPO
	var total int64

	db := r.db.WithContext(ctx).Model(&DiscountProductPO{}).Where("discount_id = ?", discountID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	if err := db.Order("id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.DiscountProduct, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// discountCategoryRepo 折扣关联分类仓储实现
type discountCategoryRepo struct {
	db *gorm.DB
}

// NewDiscountCategoryRepo 创建折扣关联分类仓储
func NewDiscountCategoryRepo(db *gorm.DB) biz.DiscountCategoryRepository {
	return &discountCategoryRepo{db: db}
}

// ListByDiscountID 根据折扣 ID 获取关联分类列表（分页）
func (r *discountCategoryRepo) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountCategory, int64, error) {
	var pos []DiscountCategoryPO
	var total int64

	db := r.db.WithContext(ctx).Model(&DiscountCategoryPO{}).Where("discount_id = ?", discountID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	if err := db.Order("id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.DiscountCategory, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// discountManufacturerRepo 折扣关联制造商仓储实现
type discountManufacturerRepo struct {
	db *gorm.DB
}

// NewDiscountManufacturerRepo 创建折扣关联制造商仓储
func NewDiscountManufacturerRepo(db *gorm.DB) biz.DiscountManufacturerRepository {
	return &discountManufacturerRepo{db: db}
}

// ListByDiscountID 根据折扣 ID 获取关联制造商列表（分页）
func (r *discountManufacturerRepo) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountManufacturer, int64, error) {
	var pos []DiscountManufacturerPO
	var total int64

	db := r.db.WithContext(ctx).Model(&DiscountManufacturerPO{}).Where("discount_id = ?", discountID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	if err := db.Order("id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.DiscountManufacturer, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// discountUsageHistoryRepo 折扣使用历史仓储实现
type discountUsageHistoryRepo struct {
	db *gorm.DB
}

// NewDiscountUsageHistoryRepo 创建折扣使用历史仓储
func NewDiscountUsageHistoryRepo(db *gorm.DB) biz.DiscountUsageHistoryRepository {
	return &discountUsageHistoryRepo{db: db}
}

// ListByDiscountID 根据折扣 ID 获取使用历史列表（分页）
func (r *discountUsageHistoryRepo) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountUsageHistory, int64, error) {
	var pos []DiscountUsageHistoryPO
	var total int64

	db := r.db.WithContext(ctx).Model(&DiscountUsageHistoryPO{}).Where("discount_id = ?", discountID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	if err := db.Order("used_on DESC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.DiscountUsageHistory, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}
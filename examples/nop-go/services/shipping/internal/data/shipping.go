// Package data 提供 shipping 服务的数据访问层
//
// 包含四张表的 PO 定义与仓储实现：
// 1. shipping_providers — 配送提供者
// 2. shipping_methods — 配送方式
// 3. shipping_delivery_dates — 配送日期
// 4. shipping_warehouses — 仓库
package data

import (
	"context"
	"time"

	"nop-go/services/shipping/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================

// ProviderPO 配送提供者持久化对象
// 对应数据库表 shipping_providers
type ProviderPO struct {
	ID            uint      `gorm:"column:id;primaryKey" db:"id"`                          // 主键 ID
	Name          string    `gorm:"column:name;size:256" db:"name"`                        // 提供者名称
	SystemKeyword string    `gorm:"column:system_keyword;size:128;uniqueIndex" db:"system_keyword"` // 系统关键字标识
	DisplayOrder  int       `gorm:"column:display_order" db:"display_order"`               // 显示排序
	IsActive      bool      `gorm:"column:is_active;default:true" db:"is_active"`          // 是否启用
	LogoURL       string    `gorm:"column:logo_url;size:512" db:"logo_url"`                // Logo 地址
	TrackingURL   string    `gorm:"column:tracking_url;size:512" db:"tracking_url"`        // 物流追踪 URL 模板
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`      // 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`      // 更新时间
}

// TableName 指定配送提供者表名
func (ProviderPO) TableName() string { return "shipping_providers" }

// ToEntity 转换为配送提供者领域实体
func (po *ProviderPO) ToEntity() *biz.Provider {
	return &biz.Provider{
		ID: po.ID, Name: po.Name, SystemKeyword: po.SystemKeyword,
		DisplayOrder: po.DisplayOrder, IsActive: po.IsActive,
		LogoURL: po.LogoURL, TrackingURL: po.TrackingURL,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// MethodPO 配送方式持久化对象
// 对应数据库表 shipping_methods
type MethodPO struct {
	ID             uint      `gorm:"column:id;primaryKey" db:"id"`                          // 主键 ID
	Name           string    `gorm:"column:name;size:256" db:"name"`                        // 配送方式名称
	SystemKeyword  string    `gorm:"column:system_keyword;size:128" db:"system_keyword"`    // 系统关键字标识
	ProviderID     uint      `gorm:"column:provider_id;index" db:"provider_id"`             // 关联的配送提供者 ID
	DisplayOrder   int       `gorm:"column:display_order" db:"display_order"`               // 显示排序
	IsActive       bool      `gorm:"column:is_active;default:true" db:"is_active"`          // 是否启用
	Rate           float64   `gorm:"column:rate;type:decimal(10,2)" db:"rate"`              // 基础运费
	MinOrderAmount float64   `gorm:"column:min_order_amount;type:decimal(10,2)" db:"min_order_amount"` // 免运费最低订单金额
	MaxOrderAmount float64   `gorm:"column:max_order_amount;type:decimal(10,2)" db:"max_order_amount"` // 运费适用最高订单金额
	EstimatedDays  int       `gorm:"column:estimated_days" db:"estimated_days"`             // 预计配送天数
	Description    string    `gorm:"column:description;size:512" db:"description"`          // 配送方式描述
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`      // 创建时间
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`      // 更新时间
}

// TableName 指定配送方式表名
func (MethodPO) TableName() string { return "shipping_methods" }

// ToEntity 转换为配送方式领域实体
func (po *MethodPO) ToEntity() *biz.Method {
	return &biz.Method{
		ID: po.ID, Name: po.Name, SystemKeyword: po.SystemKeyword,
		ProviderID: po.ProviderID, DisplayOrder: po.DisplayOrder,
		IsActive: po.IsActive, Rate: po.Rate,
		MinOrderAmount: po.MinOrderAmount, MaxOrderAmount: po.MaxOrderAmount,
		EstimatedDays: po.EstimatedDays, Description: po.Description,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// DeliveryDatePO 配送日期持久化对象
// 对应数据库表 shipping_delivery_dates
type DeliveryDatePO struct {
	ID               uint      `gorm:"column:id;primaryKey" db:"id"`                          // 主键 ID
	ShippingMethodID uint      `gorm:"column:shipping_method_id;index" db:"shipping_method_id"` // 关联的配送方式 ID
	DeliveryDate     string    `gorm:"column:delivery_date;size:32" db:"delivery_date"`       // 可选配送日期
	IsAvailable      bool      `gorm:"column:is_available;default:true" db:"is_available"`    // 该日期是否可选
	Description      string    `gorm:"column:description;size:256" db:"description"`          // 日期说明
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`      // 创建时间
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`      // 更新时间
}

// TableName 指定配送日期表名
func (DeliveryDatePO) TableName() string { return "shipping_delivery_dates" }

// ToEntity 转换为配送日期领域实体
func (po *DeliveryDatePO) ToEntity() *biz.DeliveryDate {
	return &biz.DeliveryDate{
		ID: po.ID, ShippingMethodID: po.ShippingMethodID,
		DeliveryDate: po.DeliveryDate, IsAvailable: po.IsAvailable,
		Description: po.Description,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// WarehousePO 仓库持久化对象
// 对应数据库表 shipping_warehouses
type WarehousePO struct {
	ID          uint      `gorm:"column:id;primaryKey" db:"id"`                          // 主键 ID
	Name        string    `gorm:"column:name;size:256" db:"name"`                        // 仓库名称
	Code        string    `gorm:"column:code;size:64;uniqueIndex" db:"code"`             // 仓库编码
	Address     string    `gorm:"column:address;size:512" db:"address"`                  // 仓库地址
	City        string    `gorm:"column:city;size:128" db:"city"`                        // 城市
	CountryID   uint      `gorm:"column:country_id" db:"country_id"`                     // 国家 ID
	StateID     uint      `gorm:"column:state_id" db:"state_id"`                         // 省/州 ID
	ZipCode     string    `gorm:"column:zip_code;size:32" db:"zip_code"`                 // 邮编
	PhoneNumber string    `gorm:"column:phone_number;size:32" db:"phone_number"`         // 联系电话
	IsActive    bool      `gorm:"column:is_active;default:true" db:"is_active"`          // 是否启用
	Latitude    float64   `gorm:"column:latitude;type:decimal(10,6)" db:"latitude"`      // 纬度
	Longitude   float64   `gorm:"column:longitude;type:decimal(10,6)" db:"longitude"`    // 经度
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`      // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`      // 更新时间
}

// TableName 指定仓库表名
func (WarehousePO) TableName() string { return "shipping_warehouses" }

// ToEntity 转换为仓库领域实体
func (po *WarehousePO) ToEntity() *biz.Warehouse {
	return &biz.Warehouse{
		ID: po.ID, Name: po.Name, Code: po.Code,
		Address: po.Address, City: po.City,
		CountryID: po.CountryID, StateID: po.StateID,
		ZipCode: po.ZipCode, PhoneNumber: po.PhoneNumber,
		IsActive: po.IsActive, Latitude: po.Latitude, Longitude: po.Longitude,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// providerRepo 配送提供者仓储实现
type providerRepo struct {
	db *gorm.DB
}

// NewProviderRepo 创建配送提供者仓储
func NewProviderRepo(db *gorm.DB) biz.ProviderRepository {
	return &providerRepo{db: db}
}

// List 获取配送提供者列表（分页）
func (r *providerRepo) List(ctx context.Context, page, pageSize int) ([]*biz.Provider, int64, error) {
	var pos []ProviderPO
	var total int64

	r.db.WithContext(ctx).Model(&ProviderPO{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Provider, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// Update 更新配送提供者
func (r *providerRepo) Update(ctx context.Context, provider *biz.Provider) (*biz.Provider, error) {
	po := &ProviderPO{
		ID: provider.ID, Name: provider.Name, SystemKeyword: provider.SystemKeyword,
		DisplayOrder: provider.DisplayOrder, IsActive: provider.IsActive,
		LogoURL: provider.LogoURL, TrackingURL: provider.TrackingURL,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// methodRepo 配送方式仓储实现
type methodRepo struct {
	db *gorm.DB
}

// NewMethodRepo 创建配送方式仓储
func NewMethodRepo(db *gorm.DB) biz.MethodRepository {
	return &methodRepo{db: db}
}

// List 获取配送方式列表（分页）
func (r *methodRepo) List(ctx context.Context, page, pageSize int) ([]*biz.Method, int64, error) {
	var pos []MethodPO
	var total int64

	r.db.WithContext(ctx).Model(&MethodPO{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Method, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// Create 创建配送方式
func (r *methodRepo) Create(ctx context.Context, method *biz.Method) (*biz.Method, error) {
	po := &MethodPO{
		Name: method.Name, SystemKeyword: method.SystemKeyword,
		ProviderID: method.ProviderID, DisplayOrder: method.DisplayOrder,
		IsActive: method.IsActive, Rate: method.Rate,
		MinOrderAmount: method.MinOrderAmount, MaxOrderAmount: method.MaxOrderAmount,
		EstimatedDays: method.EstimatedDays, Description: method.Description,
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新配送方式
func (r *methodRepo) Update(ctx context.Context, method *biz.Method) (*biz.Method, error) {
	po := &MethodPO{
		ID: method.ID, Name: method.Name, SystemKeyword: method.SystemKeyword,
		ProviderID: method.ProviderID, DisplayOrder: method.DisplayOrder,
		IsActive: method.IsActive, Rate: method.Rate,
		MinOrderAmount: method.MinOrderAmount, MaxOrderAmount: method.MaxOrderAmount,
		EstimatedDays: method.EstimatedDays, Description: method.Description,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Delete 删除配送方式
func (r *methodRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&MethodPO{}, id).Error
}

// deliveryDateRepo 配送日期仓储实现
type deliveryDateRepo struct {
	db *gorm.DB
}

// NewDeliveryDateRepo 创建配送日期仓储
func NewDeliveryDateRepo(db *gorm.DB) biz.DeliveryDateRepository {
	return &deliveryDateRepo{db: db}
}

// List 获取配送日期列表（分页）
func (r *deliveryDateRepo) List(ctx context.Context, page, pageSize int) ([]*biz.DeliveryDate, int64, error) {
	var pos []DeliveryDatePO
	var total int64

	r.db.WithContext(ctx).Model(&DeliveryDatePO{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("delivery_date ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.DeliveryDate, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// Create 创建配送日期
func (r *deliveryDateRepo) Create(ctx context.Context, date *biz.DeliveryDate) (*biz.DeliveryDate, error) {
	po := &DeliveryDatePO{
		ShippingMethodID: date.ShippingMethodID, DeliveryDate: date.DeliveryDate,
		IsAvailable: date.IsAvailable, Description: date.Description,
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新配送日期
func (r *deliveryDateRepo) Update(ctx context.Context, date *biz.DeliveryDate) (*biz.DeliveryDate, error) {
	po := &DeliveryDatePO{
		ID: date.ID, ShippingMethodID: date.ShippingMethodID,
		DeliveryDate: date.DeliveryDate, IsAvailable: date.IsAvailable,
		Description: date.Description,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// warehouseRepo 仓库仓储实现
type warehouseRepo struct {
	db *gorm.DB
}

// NewWarehouseRepo 创建仓库仓储
func NewWarehouseRepo(db *gorm.DB) biz.WarehouseRepository {
	return &warehouseRepo{db: db}
}

// List 获取仓库列表（分页）
func (r *warehouseRepo) List(ctx context.Context, page, pageSize int) ([]*biz.Warehouse, int64, error) {
	var pos []WarehousePO
	var total int64

	r.db.WithContext(ctx).Model(&WarehousePO{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Warehouse, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// Create 创建仓库
func (r *warehouseRepo) Create(ctx context.Context, warehouse *biz.Warehouse) (*biz.Warehouse, error) {
	po := &WarehousePO{
		Name: warehouse.Name, Code: warehouse.Code,
		Address: warehouse.Address, City: warehouse.City,
		CountryID: warehouse.CountryID, StateID: warehouse.StateID,
		ZipCode: warehouse.ZipCode, PhoneNumber: warehouse.PhoneNumber,
		IsActive: warehouse.IsActive, Latitude: warehouse.Latitude, Longitude: warehouse.Longitude,
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新仓库
func (r *warehouseRepo) Update(ctx context.Context, warehouse *biz.Warehouse) (*biz.Warehouse, error) {
	po := &WarehousePO{
		ID: warehouse.ID, Name: warehouse.Name, Code: warehouse.Code,
		Address: warehouse.Address, City: warehouse.City,
		CountryID: warehouse.CountryID, StateID: warehouse.StateID,
		ZipCode: warehouse.ZipCode, PhoneNumber: warehouse.PhoneNumber,
		IsActive: warehouse.IsActive, Latitude: warehouse.Latitude, Longitude: warehouse.Longitude,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Package data 提供 tax 服务的数据访问层
//
// 包含两张表的 PO 定义与仓储实现：
// 1. tax_providers — 税务提供者
// 2. tax_categories — 税类别
package data

import (
	"context"
	"time"

	"nop-go/services/tax/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================

// ProviderPO 税务提供者持久化对象
// 对应数据库表 tax_providers
type ProviderPO struct {
	ID            uint      `gorm:"column:id;primaryKey" db:"id"`                          // 主键 ID
	Name          string    `gorm:"column:name;size:256" db:"name"`                        // 提供者名称
	SystemKeyword string    `gorm:"column:system_keyword;size:128;uniqueIndex" db:"system_keyword"` // 系统关键字标识
	DisplayOrder  int       `gorm:"column:display_order" db:"display_order"`               // 显示排序
	IsActive      bool      `gorm:"column:is_active;default:true" db:"is_active"`          // 是否启用
	IsPrimary     bool      `gorm:"column:is_primary;default:false" db:"is_primary"`       // 是否为主要税务提供者
	LogoURL       string    `gorm:"column:logo_url;size:512" db:"logo_url"`                // Logo 地址
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`      // 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`      // 更新时间
}

// TableName 指定税务提供者表名
func (ProviderPO) TableName() string { return "tax_providers" }

// ToEntity 转换为税务提供者领域实体
func (po *ProviderPO) ToEntity() *biz.Provider {
	return &biz.Provider{
		ID: po.ID, Name: po.Name, SystemKeyword: po.SystemKeyword,
		DisplayOrder: po.DisplayOrder, IsActive: po.IsActive,
		IsPrimary: po.IsPrimary, LogoURL: po.LogoURL,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// CategoryPO 税类别持久化对象
// 对应数据库表 tax_categories
type CategoryPO struct {
	ID           uint      `gorm:"column:id;primaryKey" db:"id"`                          // 主键 ID
	Name         string    `gorm:"column:name;size:256" db:"name"`                        // 税类别名称
	Rate         float64   `gorm:"column:rate;type:decimal(10,4)" db:"rate"`              // 税率百分比
	DisplayOrder int       `gorm:"column:display_order" db:"display_order"`               // 显示排序
	IsActive     bool      `gorm:"column:is_active;default:true" db:"is_active"`          // 是否启用
	Description  string    `gorm:"column:description;size:512" db:"description"`          // 税类别描述
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`      // 创建时间
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`      // 更新时间
}

// TableName 指定税类别表名
func (CategoryPO) TableName() string { return "tax_categories" }

// ToEntity 转换为税类别领域实体
func (po *CategoryPO) ToEntity() *biz.Category {
	return &biz.Category{
		ID: po.ID, Name: po.Name, Rate: po.Rate,
		DisplayOrder: po.DisplayOrder, IsActive: po.IsActive,
		Description: po.Description,
		CreatedAt: po.CreatedAt, UpdatedAt: po.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// providerRepo 税务提供者仓储实现
type providerRepo struct {
	db *gorm.DB
}

// NewProviderRepo 创建税务提供者仓储
func NewProviderRepo(db *gorm.DB) biz.ProviderRepository {
	return &providerRepo{db: db}
}

// List 获取税务提供者列表（分页）
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

// Update 更新税务提供者
func (r *providerRepo) Update(ctx context.Context, provider *biz.Provider) (*biz.Provider, error) {
	po := &ProviderPO{
		ID: provider.ID, Name: provider.Name, SystemKeyword: provider.SystemKeyword,
		DisplayOrder: provider.DisplayOrder, IsActive: provider.IsActive,
		IsPrimary: provider.IsPrimary, LogoURL: provider.LogoURL,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// categoryRepo 税类别仓储实现
type categoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepo 创建税类别仓储
func NewCategoryRepo(db *gorm.DB) biz.CategoryRepository {
	return &categoryRepo{db: db}
}

// List 获取税类别列表（分页）
func (r *categoryRepo) List(ctx context.Context, page, pageSize int) ([]*biz.Category, int64, error) {
	var pos []CategoryPO
	var total int64

	r.db.WithContext(ctx).Model(&CategoryPO{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Category, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}
	return items, total, nil
}

// Create 创建税类别
func (r *categoryRepo) Create(ctx context.Context, category *biz.Category) (*biz.Category, error) {
	po := &CategoryPO{
		Name: category.Name, Rate: category.Rate,
		DisplayOrder: category.DisplayOrder, IsActive: category.IsActive,
		Description: category.Description,
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Update 更新税类别
func (r *categoryRepo) Update(ctx context.Context, category *biz.Category) (*biz.Category, error) {
	po := &CategoryPO{
		ID: category.ID, Name: category.Name, Rate: category.Rate,
		DisplayOrder: category.DisplayOrder, IsActive: category.IsActive,
		Description: category.Description,
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Delete 删除税类别
func (r *categoryRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&CategoryPO{}, id).Error
}
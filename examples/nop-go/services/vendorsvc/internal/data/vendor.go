// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/vendorsvc/internal/biz"

	"gorm.io/gorm"
)

// VendorPO 供应商持久化对象。
type VendorPO struct {
	ID           uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name         string    `gorm:"size:256;column:name" db:"name" json:"name"`
	Email        string    `gorm:"size:256;column:email" db:"email" json:"email"`
	Description  string    `gorm:"type:text;column:description" db:"description" json:"description"`
	Active       bool      `gorm:"column:active" db:"active" json:"active"`
	DisplayOrder int       `gorm:"column:display_order" db:"display_order" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (VendorPO) TableName() string {
	return "vendors"
}

// ToEntity 转换为领域实体。
func (po *VendorPO) ToEntity() *biz.Vendor {
	return &biz.Vendor{
		ID:           po.ID,
		Name:         po.Name,
		Email:        po.Email,
		Description:  po.Description,
		Active:       po.Active,
		DisplayOrder: po.DisplayOrder,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// VendorApplyPO 供应商申请持久化对象。
type VendorApplyPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name        string    `gorm:"size:256;column:name" db:"name" json:"name"`
	Email       string    `gorm:"size:256;column:email" db:"email" json:"email"`
	Description string    `gorm:"type:text;column:description" db:"description" json:"description"`
	Status      string    `gorm:"size:32;column:status" db:"status" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
}

// TableName 表名。
func (VendorApplyPO) TableName() string {
	return "vendor_applies"
}

// ToEntity 转换为领域实体。
func (po *VendorApplyPO) ToEntity() *biz.VendorApply {
	return &biz.VendorApply{
		ID:          po.ID,
		Name:        po.Name,
		Email:       po.Email,
		Description: po.Description,
		Status:      po.Status,
		CreatedAt:   po.CreatedAt,
	}
}

// vendorRepo 供应商仓储实现。
type vendorRepo struct {
	db *gorm.DB
}

// NewVendorRepo 创建供应商仓储。
func NewVendorRepo(db *gorm.DB) biz.VendorRepository {
	return &vendorRepo{db: db}
}

// Create 创建供应商。
func (r *vendorRepo) Create(ctx context.Context, vendor *biz.Vendor) error {
	po := &VendorPO{
		Name:         vendor.Name,
		Email:        vendor.Email,
		Description:  vendor.Description,
		Active:       vendor.Active,
		DisplayOrder: vendor.DisplayOrder,
		CreatedAt:    vendor.CreatedAt,
		UpdatedAt:    vendor.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取供应商。
func (r *vendorRepo) GetByID(ctx context.Context, id uint) (*biz.Vendor, error) {
	var po VendorPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取供应商列表。
func (r *vendorRepo) List(ctx context.Context, page, size int) ([]*biz.Vendor, int64, error) {
	var pos []VendorPO
	var total int64

	r.db.WithContext(ctx).Model(&VendorPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	vendors := make([]*biz.Vendor, len(pos))
	for i, po := range pos {
		vendors[i] = po.ToEntity()
	}

	return vendors, total, nil
}

// Update 更新供应商。
func (r *vendorRepo) Update(ctx context.Context, vendor *biz.Vendor) error {
	return r.db.WithContext(ctx).Model(&VendorPO{}).Where("id = ?", vendor.ID).Updates(map[string]interface{}{
		"name":          vendor.Name,
		"email":         vendor.Email,
		"description":   vendor.Description,
		"active":        vendor.Active,
		"display_order": vendor.DisplayOrder,
		"updated_at":    vendor.UpdatedAt,
	}).Error
}

// Delete 删除供应商。
func (r *vendorRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&VendorPO{}, id).Error
}

// vendorApplyRepo 供应商申请仓储实现。
type vendorApplyRepo struct {
	db *gorm.DB
}

// NewVendorApplyRepo 创建供应商申请仓储。
func NewVendorApplyRepo(db *gorm.DB) biz.VendorApplyRepository {
	return &vendorApplyRepo{db: db}
}

// Create 创建供应商申请。
func (r *vendorApplyRepo) Create(ctx context.Context, apply *biz.VendorApply) error {
	po := &VendorApplyPO{
		Name:        apply.Name,
		Email:       apply.Email,
		Description: apply.Description,
		Status:      apply.Status,
		CreatedAt:   apply.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取申请。
func (r *vendorApplyRepo) GetByID(ctx context.Context, id uint) (*biz.VendorApply, error) {
	var po VendorApplyPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}
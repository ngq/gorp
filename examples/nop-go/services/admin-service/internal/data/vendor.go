// Package data 供应商模块数据层 —— 供应商的持久化对象与仓储实现
package data

import (
	"context"

	"nop-go/services/admin-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== 持久化对象（PO） ====================

// VendorPO 供应商持久化对象 —— 映射数据库 vendors 表
type VendorPO struct {
	ID          uint   `gorm:"primaryKey" db:"id"`
	Name        string `gorm:"column:name;type:varchar(100);not null" db:"name"`             // 供应商名称
	Code        string `gorm:"column:code;type:varchar(100);uniqueIndex;not null" db:"code"` // 供应商编码
	Contact     string `gorm:"column:contact;type:varchar(50)" db:"contact"`                    // 联系人
	Phone       string `gorm:"column:phone;type:varchar(20)" db:"phone"`                      // 联系电话
	Email       string `gorm:"column:email;type:varchar(100)" db:"email"`                     // 邮箱
	Address     string `gorm:"column:address;type:varchar(500)" db:"address"`                   // 地址
	Category    string `gorm:"column:category;type:varchar(50);index" db:"category"`             // 分类
	BankName    string `gorm:"column:bank_name;type:varchar(100)" db:"bank_name"`                 // 开户银行
	BankAccount string `gorm:"column:bank_account;type:varchar(50)" db:"bank_account"`               // 银行账号
	Status      int    `gorm:"column:status;type:tinyint;default:1;index" db:"status"`         // 状态
	CreatedAt   string `gorm:"column:created_at" db:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at" db:"updated_at"`
}

// TableName 指定供应商表名
func (VendorPO) TableName() string { return "vendors" }

// ==================== PO ↔ Entity 转换 ====================

// toEntity 将 VendorPO 转换为 biz.Vendor 领域实体
func (v *VendorPO) toEntity() *biz.Vendor {
	return &biz.Vendor{
		ID:          v.ID,
		Name:        v.Name,
		Code:        v.Code,
		Contact:     v.Contact,
		Phone:       v.Phone,
		Email:       v.Email,
		Address:     v.Address,
		Category:    v.Category,
		BankName:    v.BankName,
		BankAccount: v.BankAccount,
		Status:      v.Status,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

// vendorToPO 将 biz.Vendor 领域实体转换为 VendorPO
func vendorToPO(v *biz.Vendor) *VendorPO {
	return &VendorPO{
		ID:          v.ID,
		Name:        v.Name,
		Code:        v.Code,
		Contact:     v.Contact,
		Phone:       v.Phone,
		Email:       v.Email,
		Address:     v.Address,
		Category:    v.Category,
		BankName:    v.BankName,
		BankAccount: v.BankAccount,
		Status:      v.Status,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// vendorRepo 供应商仓储实现
type vendorRepo struct {
	db *gorm.DB
}

// NewVendorRepo 创建供应商仓储
func NewVendorRepo(db *gorm.DB) biz.VendorRepo {
	return &vendorRepo{db: db}
}

func (r *vendorRepo) Create(ctx context.Context, v *biz.Vendor) error {
	po := vendorToPO(v)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *vendorRepo) GetByID(ctx context.Context, id uint) (*biz.Vendor, error) {
	var po VendorPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *vendorRepo) GetByCode(ctx context.Context, code string) (*biz.Vendor, error) {
	var po VendorPO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *vendorRepo) List(ctx context.Context, category string, status int, page, pageSize int) ([]*biz.Vendor, int64, error) {
	var pos []*VendorPO
	var total int64
	q := r.db.WithContext(ctx).Model(&VendorPO{})
	// 分类过滤
	if category != "" {
		q = q.Where("category = ?", category)
	}
	// 状态过滤：status < 0 表示不过滤
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.Vendor, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *vendorRepo) Update(ctx context.Context, v *biz.Vendor) error {
	po := vendorToPO(v)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *vendorRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&VendorPO{}, id).Error
}

// Package data 提供 payment 服务的数据访问层
//
// 包含两张表的 PO 定义与仓储实现：
// 1. payment_methods — 支付方式
// 2. payment_method_restrictions — 支付方式限制
package data

import (
	"context"
	"time"

	"nop-go/services/payment/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================

// PaymentMethodPO 支付方式持久化对象
// 对应数据库表 payment_methods
type PaymentMethodPO struct {
	ID                    uint      `gorm:"column:id;primaryKey" db:"id"`                                  // 主键 ID
	Name                  string    `gorm:"column:name;size:256" db:"name"`                                // 支付方式名称
	SystemKeyword         string    `gorm:"column:system_keyword;size:128;uniqueIndex" db:"system_keyword"` // 系统关键字标识
	DisplayOrder          int       `gorm:"column:display_order" db:"display_order"`                       // 显示排序
	IsActive              bool      `gorm:"column:is_active;default:true" db:"is_active"`                  // 是否启用
	LogoURL               string    `gorm:"column:logo_url;size:512" db:"logo_url"`                        // Logo 地址
	SupportsRefund        bool      `gorm:"column:supports_refund;default:false" db:"supports_refund"`     // 是否支持退款
	SupportsPartialRefund bool      `gorm:"column:supports_partial_refund;default:false" db:"supports_partial_refund"` // 是否支持部分退款
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`              // 创建时间
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`              // 更新时间
}

// TableName 指定支付方式表名
func (PaymentMethodPO) TableName() string {
	return "payment_methods"
}

// ToEntity 转换为支付方式领域实体
func (po *PaymentMethodPO) ToEntity() *biz.PaymentMethod {
	return &biz.PaymentMethod{
		ID:                    po.ID,
		Name:                  po.Name,
		SystemKeyword:         po.SystemKeyword,
		DisplayOrder:          po.DisplayOrder,
		IsActive:              po.IsActive,
		LogoURL:               po.LogoURL,
		SupportsRefund:        po.SupportsRefund,
		SupportsPartialRefund: po.SupportsPartialRefund,
		CreatedAt:             po.CreatedAt,
		UpdatedAt:             po.UpdatedAt,
	}
}

// MethodRestrictionPO 支付方式限制持久化对象
// 对应数据库表 payment_method_restrictions
type MethodRestrictionPO struct {
	ID               uint      `gorm:"column:id;primaryKey" db:"id"`                        // 主键 ID
	PaymentMethodID  uint      `gorm:"column:payment_method_id;index" db:"payment_method_id"` // 关联的支付方式 ID
	MinOrderAmount   float64   `gorm:"column:min_order_amount;type:decimal(10,2)" db:"min_order_amount"` // 最小订单金额
	MaxOrderAmount   float64   `gorm:"column:max_order_amount;type:decimal(10,2)" db:"max_order_amount"` // 最大订单金额
	RestrictionType  string    `gorm:"column:restriction_type;size:64" db:"restriction_type"` // 限制类型
	RestrictionValue string    `gorm:"column:restriction_value;size:512" db:"restriction_value"` // 限制值
	IsActive         bool      `gorm:"column:is_active;default:true" db:"is_active"`         // 是否启用
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`     // 创建时间
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`     // 更新时间
}

// TableName 指定支付方式限制表名
func (MethodRestrictionPO) TableName() string {
	return "payment_method_restrictions"
}

// ToEntity 转换为支付方式限制领域实体
func (po *MethodRestrictionPO) ToEntity() *biz.MethodRestriction {
	return &biz.MethodRestriction{
		ID:               po.ID,
		PaymentMethodID:  po.PaymentMethodID,
		MinOrderAmount:   po.MinOrderAmount,
		MaxOrderAmount:   po.MaxOrderAmount,
		RestrictionType:  po.RestrictionType,
		RestrictionValue: po.RestrictionValue,
		IsActive:         po.IsActive,
		CreatedAt:        po.CreatedAt,
		UpdatedAt:        po.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// paymentMethodRepo 支付方式仓储实现
type paymentMethodRepo struct {
	db *gorm.DB
}

// NewPaymentMethodRepo 创建支付方式仓储
func NewPaymentMethodRepo(db *gorm.DB) biz.PaymentMethodRepository {
	return &paymentMethodRepo{db: db}
}

// List 获取支付方式列表（分页）
func (r *paymentMethodRepo) List(ctx context.Context, page, pageSize int) ([]*biz.PaymentMethod, int64, error) {
	var pos []PaymentMethodPO
	var total int64

	// 统计总数
	r.db.WithContext(ctx).Model(&PaymentMethodPO{}).Count(&total)

	// 分页查询，按显示排序升序
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	// PO 转领域实体
	methods := make([]*biz.PaymentMethod, len(pos))
	for i, po := range pos {
		methods[i] = po.ToEntity()
	}
	return methods, total, nil
}

// Update 更新支付方式
func (r *paymentMethodRepo) Update(ctx context.Context, method *biz.PaymentMethod) (*biz.PaymentMethod, error) {
	po := &PaymentMethodPO{
		ID:                    method.ID,
		Name:                  method.Name,
		SystemKeyword:         method.SystemKeyword,
		DisplayOrder:          method.DisplayOrder,
		IsActive:              method.IsActive,
		LogoURL:               method.LogoURL,
		SupportsRefund:        method.SupportsRefund,
		SupportsPartialRefund: method.SupportsPartialRefund,
	}

	// GORM Save 会更新所有字段（包括零值）
	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// methodRestrictionRepo 支付方式限制仓储实现
type methodRestrictionRepo struct {
	db *gorm.DB
}

// NewMethodRestrictionRepo 创建支付方式限制仓储
func NewMethodRestrictionRepo(db *gorm.DB) biz.MethodRestrictionRepository {
	return &methodRestrictionRepo{db: db}
}

// List 获取支付方式限制列表（分页）
func (r *methodRestrictionRepo) List(ctx context.Context, page, pageSize int) ([]*biz.MethodRestriction, int64, error) {
	var pos []MethodRestrictionPO
	var total int64

	// 统计总数
	r.db.WithContext(ctx).Model(&MethodRestrictionPO{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("id ASC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	// PO 转领域实体
	restrictions := make([]*biz.MethodRestriction, len(pos))
	for i, po := range pos {
		restrictions[i] = po.ToEntity()
	}
	return restrictions, total, nil
}

// Update 更新支付方式限制
func (r *methodRestrictionRepo) Update(ctx context.Context, restriction *biz.MethodRestriction) (*biz.MethodRestriction, error) {
	po := &MethodRestrictionPO{
		ID:               restriction.ID,
		PaymentMethodID:  restriction.PaymentMethodID,
		MinOrderAmount:   restriction.MinOrderAmount,
		MaxOrderAmount:   restriction.MaxOrderAmount,
		RestrictionType:  restriction.RestrictionType,
		RestrictionValue: restriction.RestrictionValue,
		IsActive:         restriction.IsActive,
	}

	// GORM Save 会更新所有字段
	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

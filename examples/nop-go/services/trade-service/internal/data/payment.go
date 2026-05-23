// Package data 包含交易服务的数据访问层
// payment.go 定义支付相关 PO 及仓库实现
package data

import (
	"context"

	"nop-go/services/trade-service/internal/biz"

	"gorm.io/gorm"
)

// ============================================================================
// PO 定义
// ============================================================================

// PaymentPO 支付记录持久化对象
type PaymentPO struct {
	gorm.Model
	OrderID         string  `gorm:"index;not null;size:64;column:order_id" db:"order_id"`
	UserID          string  `gorm:"index;not null;size:64;column:user_id" db:"user_id"`
	Amount          float64 `gorm:"type:decimal(12,2);not null;column:amount" db:"amount"`
	Currency        string  `gorm:"size:8;default:'CNY';column:currency" db:"currency"`
	Method          string  `gorm:"size:64;column:method" db:"method"`
	Status          string  `gorm:"size:32;default:'pending';column:status" db:"status"`
	TransactionID   string  `gorm:"size:128;column:transaction_id" db:"transaction_id"`
	PaymentMethodID string  `gorm:"size:64;column:payment_method_id" db:"payment_method_id"`
}

// TableName 指定支付记录表名
func (PaymentPO) TableName() string { return "payments" }

// ToEntity 转换为支付领域实体
func (po *PaymentPO) ToEntity() *biz.Payment {
	return &biz.Payment{
		ID:              fmtID(po.ID),
		OrderID:         po.OrderID,
		UserID:          po.UserID,
		Amount:          po.Amount,
		Currency:        po.Currency,
		Method:          po.Method,
		Status:          po.Status,
		TransactionID:   po.TransactionID,
		PaymentMethodID: po.PaymentMethodID,
		CreatedAt:       po.CreatedAt,
		UpdatedAt:       po.UpdatedAt,
	}
}

// PaymentMethodPO 支付方式持久化对象
type PaymentMethodPO struct {
	gorm.Model
	UserID    string `gorm:"index;not null;size:64;column:user_id" db:"user_id"`
	Type      string `gorm:"size:64;not null;column:type" db:"type"`
	Provider  string `gorm:"size:64;column:provider" db:"provider"`
	Last4     string `gorm:"size:8;column:last4" db:"last4"`
	IsDefault bool   `gorm:"default:false;column:is_default" db:"is_default"`
}

// TableName 指定支付方式表名
func (PaymentMethodPO) TableName() string { return "payment_methods" }

// ToEntity 转换为支付方式领域实体
func (po *PaymentMethodPO) ToEntity() *biz.PaymentMethod {
	return &biz.PaymentMethod{
		ID:        fmtID(po.ID),
		UserID:    po.UserID,
		Type:      po.Type,
		Provider:  po.Provider,
		Last4:     po.Last4,
		IsDefault: po.IsDefault,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// ============================================================================
// 仓库实现
// ============================================================================

// paymentRepo 支付记录仓储实现
type paymentRepo struct{ db *gorm.DB }

// NewPaymentRepo 创建支付记录仓储
func NewPaymentRepo(db *gorm.DB) biz.PaymentRepo { return &paymentRepo{db: db} }

func (r *paymentRepo) Create(ctx context.Context, payment *biz.Payment) error {
	po := &PaymentPO{
		OrderID:         payment.OrderID,
		UserID:          payment.UserID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		Method:          payment.Method,
		Status:          payment.Status,
		TransactionID:   payment.TransactionID,
		PaymentMethodID: payment.PaymentMethodID,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	payment.ID = fmtID(po.ID)
	return nil
}

func (r *paymentRepo) GetByID(ctx context.Context, id string) (*biz.Payment, error) {
	var po PaymentPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *paymentRepo) Update(ctx context.Context, payment *biz.Payment) error {
	return r.db.WithContext(ctx).Model(&PaymentPO{}).Where("id = ?", parseID(payment.ID)).Updates(map[string]interface{}{
		"status":         payment.Status,
		"transaction_id": payment.TransactionID,
	}).Error
}

func (r *paymentRepo) ListByOrderID(ctx context.Context, orderID string) ([]*biz.Payment, error) {
	var pos []PaymentPO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.Payment, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *paymentRepo) ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*biz.Payment, int64, error) {
	var pos []PaymentPO
	var total int64
	r.db.WithContext(ctx).Model(&PaymentPO{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Payment, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

// paymentMethodRepo 支付方式仓储实现
type paymentMethodRepo struct{ db *gorm.DB }

// NewPaymentMethodRepo 创建支付方式仓储
func NewPaymentMethodRepo(db *gorm.DB) biz.PaymentMethodRepo {
	return &paymentMethodRepo{db: db}
}

func (r *paymentMethodRepo) Create(ctx context.Context, method *biz.PaymentMethod) error {
	po := &PaymentMethodPO{
		UserID:    method.UserID,
		Type:      method.Type,
		Provider:  method.Provider,
		Last4:     method.Last4,
		IsDefault: method.IsDefault,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	method.ID = fmtID(po.ID)
	return nil
}

func (r *paymentMethodRepo) GetByID(ctx context.Context, id string) (*biz.PaymentMethod, error) {
	var po PaymentMethodPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *paymentMethodRepo) Update(ctx context.Context, method *biz.PaymentMethod) error {
	return r.db.WithContext(ctx).Model(&PaymentMethodPO{}).Where("id = ?", parseID(method.ID)).Updates(map[string]interface{}{
		"is_default": method.IsDefault,
		"last4":      method.Last4,
	}).Error
}

func (r *paymentMethodRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&PaymentMethodPO{}, parseID(id)).Error
}

func (r *paymentMethodRepo) ListByUserID(ctx context.Context, userID string) ([]*biz.PaymentMethod, error) {
	var pos []PaymentMethodPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.PaymentMethod, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

// Package data 包含交易服务的数据访问层
// tax.go 定义税务相关 PO 及仓库实现
// 注意：原 tax 服务中的 Provider PO 已重命名为 TaxProviderPO，Category PO 已重命名为 TaxCategoryPO
package data

import (
	"context"
	"time"

	"nop-go/services/trade-service/internal/biz"

	"gorm.io/gorm"
)

// ============================================================================
// PO 定义
// ============================================================================

// TaxProviderPO 税务服务商持久化对象（原 ProviderPO，重命名以避免冲突）
type TaxProviderPO struct {
	gorm.Model
	Name        string `gorm:"size:128;not null;column:name" db:"name"`
	Code        string `gorm:"size:64;uniqueIndex;not null;column:code" db:"code"`
	Description string `gorm:"size:512;column:description" db:"description"`
	IsActive    bool   `gorm:"default:true;column:is_active" db:"is_active"`
}

// TableName 指定税务服务商表名
func (TaxProviderPO) TableName() string { return "tax_providers" }

// ToEntity 转换为税务服务商领域实体
func (po *TaxProviderPO) ToEntity() *biz.TaxProvider {
	return &biz.TaxProvider{
		ID:          fmtID(po.ID),
		Name:        po.Name,
		Code:        po.Code,
		Description: po.Description,
		IsActive:    po.IsActive,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// TaxCategoryPO 税种分类持久化对象（原 CategoryPO，重命名以避免冲突）
type TaxCategoryPO struct {
	gorm.Model
	Name        string  `gorm:"size:128;not null;column:name" db:"name"`
	Code        string  `gorm:"size:64;uniqueIndex;not null;column:code" db:"code"`
	Description string  `gorm:"size:512;column:description" db:"description"`
	Rate        float64 `gorm:"type:decimal(8,4);not null;column:rate" db:"rate"`
	IsActive    bool    `gorm:"default:true;column:is_active" db:"is_active"`
}

// TableName 指定税种分类表名
func (TaxCategoryPO) TableName() string { return "tax_categories" }

// ToEntity 转换为税种分类领域实体
func (po *TaxCategoryPO) ToEntity() *biz.TaxCategory {
	return &biz.TaxCategory{
		ID:          fmtID(po.ID),
		Name:        po.Name,
		Code:        po.Code,
		Description: po.Description,
		Rate:        po.Rate,
		IsActive:    po.IsActive,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// TaxRatePO 税率持久化对象
type TaxRatePO struct {
	gorm.Model
	TaxCategoryID string     `gorm:"index;not null;size:64;column:tax_category_id" db:"tax_category_id"`
	Region        string     `gorm:"size:64;not null;column:region" db:"region"`
	Rate          float64    `gorm:"type:decimal(8,4);not null;column:rate" db:"rate"`
	IsActive      bool       `gorm:"default:true;column:is_active" db:"is_active"`
	EffectiveFrom time.Time  `gorm:"not null;column:effective_from" db:"effective_from"`
	EffectiveTo   *time.Time `gorm:"column:effective_to" db:"effective_to"`
}

// TableName 指定税率表名
func (TaxRatePO) TableName() string { return "tax_rates" }

// ToEntity 转换为税率领域实体
func (po *TaxRatePO) ToEntity() *biz.TaxRate {
	return &biz.TaxRate{
		ID:            fmtID(po.ID),
		TaxCategoryID: po.TaxCategoryID,
		Region:        po.Region,
		Rate:          po.Rate,
		IsActive:      po.IsActive,
		EffectiveFrom: po.EffectiveFrom,
		EffectiveTo:   po.EffectiveTo,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}
}

// TaxTransactionPO 税务交易记录持久化对象
type TaxTransactionPO struct {
	gorm.Model
	OrderID       string  `gorm:"index;not null;size:64;column:order_id" db:"order_id"`
	TaxCategoryID string  `gorm:"index;not null;size:64;column:tax_category_id" db:"tax_category_id"`
	TaxRateID     string  `gorm:"size:64;column:tax_rate_id" db:"tax_rate_id"`
	TaxableAmount float64 `gorm:"type:decimal(12,2);not null;column:taxable_amount" db:"taxable_amount"`
	TaxAmount     float64 `gorm:"type:decimal(12,2);not null;column:tax_amount" db:"tax_amount"`
	Currency      string  `gorm:"size:8;default:'CNY';column:currency" db:"currency"`
}

// TableName 指定税务交易记录表名
func (TaxTransactionPO) TableName() string { return "tax_transactions" }

// ToEntity 转换为税务交易记录领域实体
func (po *TaxTransactionPO) ToEntity() *biz.TaxTransaction {
	return &biz.TaxTransaction{
		ID:            fmtID(po.ID),
		OrderID:       po.OrderID,
		TaxCategoryID: po.TaxCategoryID,
		TaxRateID:     po.TaxRateID,
		TaxableAmount: po.TaxableAmount,
		TaxAmount:     po.TaxAmount,
		Currency:      po.Currency,
		CreatedAt:     po.CreatedAt,
	}
}

// ============================================================================
// 仓库实现
// ============================================================================

// taxProviderRepo 税务服务商仓储实现
type taxProviderRepo struct{ db *gorm.DB }

// NewTaxProviderRepo 创建税务服务商仓储
func NewTaxProviderRepo(db *gorm.DB) biz.TaxProviderRepo { return &taxProviderRepo{db: db} }

func (r *taxProviderRepo) Create(ctx context.Context, provider *biz.TaxProvider) error {
	po := &TaxProviderPO{
		Name:        provider.Name,
		Code:        provider.Code,
		Description: provider.Description,
		IsActive:    provider.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	provider.ID = fmtID(po.ID)
	return nil
}

func (r *taxProviderRepo) GetByID(ctx context.Context, id string) (*biz.TaxProvider, error) {
	var po TaxProviderPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *taxProviderRepo) Update(ctx context.Context, provider *biz.TaxProvider) error {
	return r.db.WithContext(ctx).Model(&TaxProviderPO{}).Where("id = ?", parseID(provider.ID)).Updates(map[string]interface{}{
		"name":        provider.Name,
		"description": provider.Description,
		"is_active":   provider.IsActive,
	}).Error
}

func (r *taxProviderRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TaxProviderPO{}, parseID(id)).Error
}

func (r *taxProviderRepo) List(ctx context.Context) ([]*biz.TaxProvider, error) {
	var pos []TaxProviderPO
	if err := r.db.WithContext(ctx).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.TaxProvider, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *taxProviderRepo) GetByCode(ctx context.Context, code string) (*biz.TaxProvider, error) {
	var po TaxProviderPO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

// taxCategoryRepo 税种分类仓储实现
type taxCategoryRepo struct{ db *gorm.DB }

// NewTaxCategoryRepo 创建税种分类仓储
func NewTaxCategoryRepo(db *gorm.DB) biz.TaxCategoryRepo { return &taxCategoryRepo{db: db} }

func (r *taxCategoryRepo) Create(ctx context.Context, category *biz.TaxCategory) error {
	po := &TaxCategoryPO{
		Name:        category.Name,
		Code:        category.Code,
		Description: category.Description,
		Rate:        category.Rate,
		IsActive:    category.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	category.ID = fmtID(po.ID)
	return nil
}

func (r *taxCategoryRepo) GetByID(ctx context.Context, id string) (*biz.TaxCategory, error) {
	var po TaxCategoryPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *taxCategoryRepo) Update(ctx context.Context, category *biz.TaxCategory) error {
	return r.db.WithContext(ctx).Model(&TaxCategoryPO{}).Where("id = ?", parseID(category.ID)).Updates(map[string]interface{}{
		"name":        category.Name,
		"description": category.Description,
		"rate":        category.Rate,
		"is_active":   category.IsActive,
	}).Error
}

func (r *taxCategoryRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TaxCategoryPO{}, parseID(id)).Error
}

func (r *taxCategoryRepo) List(ctx context.Context) ([]*biz.TaxCategory, error) {
	var pos []TaxCategoryPO
	if err := r.db.WithContext(ctx).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.TaxCategory, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *taxCategoryRepo) GetByCode(ctx context.Context, code string) (*biz.TaxCategory, error) {
	var po TaxCategoryPO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

// taxRateRepo 税率仓储实现
type taxRateRepo struct{ db *gorm.DB }

// NewTaxRateRepo 创建税率仓储
func NewTaxRateRepo(db *gorm.DB) biz.TaxRateRepo { return &taxRateRepo{db: db} }

func (r *taxRateRepo) Create(ctx context.Context, rate *biz.TaxRate) error {
	po := &TaxRatePO{
		TaxCategoryID: rate.TaxCategoryID,
		Region:        rate.Region,
		Rate:          rate.Rate,
		IsActive:      rate.IsActive,
		EffectiveFrom: rate.EffectiveFrom,
		EffectiveTo:   rate.EffectiveTo,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	rate.ID = fmtID(po.ID)
	return nil
}

func (r *taxRateRepo) GetByID(ctx context.Context, id string) (*biz.TaxRate, error) {
	var po TaxRatePO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *taxRateRepo) Update(ctx context.Context, rate *biz.TaxRate) error {
	return r.db.WithContext(ctx).Model(&TaxRatePO{}).Where("id = ?", parseID(rate.ID)).Updates(map[string]interface{}{
		"rate":     rate.Rate,
		"is_active": rate.IsActive,
	}).Error
}

func (r *taxRateRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TaxRatePO{}, parseID(id)).Error
}

func (r *taxRateRepo) ListByCategory(ctx context.Context, categoryID string) ([]*biz.TaxRate, error) {
	var pos []TaxRatePO
	if err := r.db.WithContext(ctx).Where("tax_category_id = ?", categoryID).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.TaxRate, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *taxRateRepo) FindEffectiveRate(ctx context.Context, categoryID, region string) (*biz.TaxRate, error) {
	var po TaxRatePO
	if err := r.db.WithContext(ctx).Where(
		"tax_category_id = ? AND region = ? AND is_active = true",
		categoryID, region,
	).Order("effective_from DESC").First(&po).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

// taxTransactionRepo 税务交易记录仓储实现
type taxTransactionRepo struct{ db *gorm.DB }

// NewTaxTransactionRepo 创建税务交易记录仓储
func NewTaxTransactionRepo(db *gorm.DB) biz.TaxTransactionRepo { return &taxTransactionRepo{db: db} }

func (r *taxTransactionRepo) Create(ctx context.Context, txn *biz.TaxTransaction) error {
	po := &TaxTransactionPO{
		OrderID:       txn.OrderID,
		TaxCategoryID: txn.TaxCategoryID,
		TaxRateID:     txn.TaxRateID,
		TaxableAmount: txn.TaxableAmount,
		TaxAmount:     txn.TaxAmount,
		Currency:      txn.Currency,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	txn.ID = fmtID(po.ID)
	return nil
}

func (r *taxTransactionRepo) GetByID(ctx context.Context, id string) (*biz.TaxTransaction, error) {
	var po TaxTransactionPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *taxTransactionRepo) ListByOrderID(ctx context.Context, orderID string) ([]*biz.TaxTransaction, error) {
	var pos []TaxTransactionPO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.TaxTransaction, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *taxTransactionRepo) ListByCategory(ctx context.Context, categoryID string, page, pageSize int) ([]*biz.TaxTransaction, int64, error) {
	var pos []TaxTransactionPO
	var total int64
	r.db.WithContext(ctx).Model(&TaxTransactionPO{}).Where("tax_category_id = ?", categoryID).Count(&total)
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Where("tax_category_id = ?", categoryID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.TaxTransaction, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, total, nil
}

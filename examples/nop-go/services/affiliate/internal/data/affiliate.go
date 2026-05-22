// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/affiliate/internal/biz"

	"gorm.io/gorm"
)

// AffiliatePO 联盟持久化对象。
type AffiliatePO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name      string    `gorm:"size:256;column:name" db:"name" json:"name"`
	Url       string    `gorm:"size:512;column:url" db:"url" json:"url"`
	Active    bool      `gorm:"column:active" db:"active" json:"active"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (AffiliatePO) TableName() string {
	return "affiliates"
}

// ToEntity 转换为领域实体。
func (po *AffiliatePO) ToEntity() *biz.Affiliate {
	return &biz.Affiliate{
		ID:        po.ID,
		Name:      po.Name,
		Url:       po.Url,
		Active:    po.Active,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// AffiliateOrderPO 联盟订单持久化对象。
type AffiliateOrderPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	AffiliateID uint      `gorm:"index;column:affiliate_id" db:"affiliate_id" json:"affiliate_id"`
	OrderNo     string    `gorm:"size:64;column:order_no" db:"order_no" json:"order_no"`
	CustomerID  uint      `gorm:"column:customer_id" db:"customer_id" json:"customer_id"`
	TotalAmount float64   `gorm:"column:total_amount" db:"total_amount" json:"total_amount"`
	Status      string    `gorm:"size:32;column:status" db:"status" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
}

// TableName 表名。
func (AffiliateOrderPO) TableName() string {
	return "affiliate_orders"
}

// ToEntity 转换为领域实体。
func (po *AffiliateOrderPO) ToEntity() *biz.AffiliateOrder {
	return &biz.AffiliateOrder{
		ID:          po.ID,
		AffiliateID: po.AffiliateID,
		OrderNo:     po.OrderNo,
		CustomerID:  po.CustomerID,
		TotalAmount: po.TotalAmount,
		Status:      po.Status,
		CreatedAt:   po.CreatedAt,
	}
}

// AffiliateCustomerPO 联盟客户持久化对象。
type AffiliateCustomerPO struct {
	ID          uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	AffiliateID uint      `gorm:"index;column:affiliate_id" db:"affiliate_id" json:"affiliate_id"`
	Username    string    `gorm:"size:128;column:username" db:"username" json:"username"`
	Email       string    `gorm:"size:256;column:email" db:"email" json:"email"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
}

// TableName 表名。
func (AffiliateCustomerPO) TableName() string {
	return "affiliate_customers"
}

// ToEntity 转换为领域实体。
func (po *AffiliateCustomerPO) ToEntity() *biz.AffiliateCustomer {
	return &biz.AffiliateCustomer{
		ID:          po.ID,
		AffiliateID: po.AffiliateID,
		Username:    po.Username,
		Email:       po.Email,
		CreatedAt:   po.CreatedAt,
	}
}

// affiliateRepo 联盟仓储实现。
type affiliateRepo struct {
	db *gorm.DB
}

// NewAffiliateRepo 创建联盟仓储。
func NewAffiliateRepo(db *gorm.DB) biz.AffiliateRepository {
	return &affiliateRepo{db: db}
}

// Create 创建联盟。
func (r *affiliateRepo) Create(ctx context.Context, aff *biz.Affiliate) error {
	po := &AffiliatePO{
		Name:      aff.Name,
		Url:       aff.Url,
		Active:    aff.Active,
		CreatedAt: aff.CreatedAt,
		UpdatedAt: aff.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取联盟。
func (r *affiliateRepo) GetByID(ctx context.Context, id uint) (*biz.Affiliate, error) {
	var po AffiliatePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取联盟列表。
func (r *affiliateRepo) List(ctx context.Context, page, size int) ([]*biz.Affiliate, int64, error) {
	var pos []AffiliatePO
	var total int64

	r.db.WithContext(ctx).Model(&AffiliatePO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Affiliate, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}

	return items, total, nil
}

// Update 更新联盟。
func (r *affiliateRepo) Update(ctx context.Context, aff *biz.Affiliate) error {
	return r.db.WithContext(ctx).Model(&AffiliatePO{}).Where("id = ?", aff.ID).Updates(map[string]interface{}{
		"name":       aff.Name,
		"url":        aff.Url,
		"active":     aff.Active,
		"updated_at": aff.UpdatedAt,
	}).Error
}

// Delete 删除联盟。
func (r *affiliateRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&AffiliatePO{}, id).Error
}

// affiliateOrderRepo 联盟订单仓储实现。
type affiliateOrderRepo struct {
	db *gorm.DB
}

// NewAffiliateOrderRepo 创建联盟订单仓储。
func NewAffiliateOrderRepo(db *gorm.DB) biz.AffiliateOrderRepository {
	return &affiliateOrderRepo{db: db}
}

// ListByAffiliateID 根据联盟ID获取关联订单。
func (r *affiliateOrderRepo) ListByAffiliateID(ctx context.Context, affiliateID uint, page, size int) ([]*biz.AffiliateOrder, int64, error) {
	var pos []AffiliateOrderPO
	var total int64

	query := r.db.WithContext(ctx).Model(&AffiliateOrderPO{}).Where("affiliate_id = ?", affiliateID)
	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.AffiliateOrder, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}

	return items, total, nil
}

// affiliateCustomerRepo 联盟客户仓储实现。
type affiliateCustomerRepo struct {
	db *gorm.DB
}

// NewAffiliateCustomerRepo 创建联盟客户仓储。
func NewAffiliateCustomerRepo(db *gorm.DB) biz.AffiliateCustomerRepository {
	return &affiliateCustomerRepo{db: db}
}

// ListByAffiliateID 根据联盟ID获取关联客户。
func (r *affiliateCustomerRepo) ListByAffiliateID(ctx context.Context, affiliateID uint, page, size int) ([]*biz.AffiliateCustomer, int64, error) {
	var pos []AffiliateCustomerPO
	var total int64

	query := r.db.WithContext(ctx).Model(&AffiliateCustomerPO{}).Where("affiliate_id = ?", affiliateID)
	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*biz.AffiliateCustomer, len(pos))
	for i, po := range pos {
		items[i] = po.ToEntity()
	}

	return items, total, nil
}
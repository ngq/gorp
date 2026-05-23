package data

import (
	"context"

	"nop-go/services/content-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== PO（持久化对象）定义 ====================

// AffiliatePO 推广合作方持久化对象
type AffiliatePO struct {
	ID         uint64  `gorm:"primaryKey;autoIncrement" db:"id"`
	Name       string  `gorm:"size:200;not null" db:"name"`
	Code       string  `gorm:"size:50;uniqueIndex;not null" db:"code"` // 编码唯一
	Contact    string  `gorm:"size:200" db:"contact"`
	Website    string  `gorm:"size:500" db:"website"`
	Commission float64 `gorm:"type:decimal(5,4);default:0" db:"commission"` // 佣金比例
	Status     string  `gorm:"size:20;default:active" db:"status"`
	CreatedAt  int64   `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt  int64   `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定推广合作方表名
func (AffiliatePO) TableName() string { return "affiliates" }

// AffiliateOrderPO 推广订单持久化对象
type AffiliateOrderPO struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement" db:"id"`
	AffiliateID uint64  `gorm:"index;not null" db:"affiliate_id"`
	OrderNo     string  `gorm:"size:100;uniqueIndex;not null" db:"order_no"` // 订单编号唯一
	Amount      float64 `gorm:"type:decimal(12,2);default:0" db:"amount"`
	Commission  float64 `gorm:"type:decimal(12,2);default:0" db:"commission"`
	Status      string  `gorm:"size:20;default:pending" db:"status"`
	CreatedAt   int64   `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   int64   `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定推广订单表名
func (AffiliateOrderPO) TableName() string { return "affiliate_orders" }

// AffiliateCustomerPO 推广客户持久化对象
type AffiliateCustomerPO struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement" db:"id"`
	AffiliateID uint64 `gorm:"index;not null" db:"affiliate_id"`
	CustomerID  uint64 `gorm:"not null" db:"customer_id"`
	Source      string `gorm:"size:100" db:"source"`
	FirstVisit  int64  `gorm:"column:first_visit" db:"first_visit"` // 时间戳
	CreatedAt   int64  `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定推广客户表名
func (AffiliateCustomerPO) TableName() string { return "affiliate_customers" }

// ==================== 仓储实现 ====================

// affiliateRepo 推广合作方仓储实现
type affiliateRepo struct {
	db *gorm.DB
}

// NewAffiliateRepo 创建推广合作方仓储
func NewAffiliateRepo(db *gorm.DB) biz.AffiliateRepo {
	return &affiliateRepo{db: db}
}

func (r *affiliateRepo) Create(ctx context.Context, affiliate *biz.Affiliate) error {
	po := r.toPO(affiliate)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *affiliateRepo) GetByID(ctx context.Context, id uint64) (*biz.Affiliate, error) {
	var po AffiliatePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *affiliateRepo) GetByCode(ctx context.Context, code string) (*biz.Affiliate, error) {
	var po AffiliatePO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil {
		return nil, err
	}
	return r.toEntity(&po), nil
}

func (r *affiliateRepo) List(ctx context.Context, offset, limit int) ([]*biz.Affiliate, error) {
	var pos []*AffiliatePO
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.Affiliate, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *affiliateRepo) Update(ctx context.Context, affiliate *biz.Affiliate) error {
	po := r.toPO(affiliate)
	return r.db.WithContext(ctx).Model(&AffiliatePO{}).Where("id = ?", po.ID).Updates(po).Error
}

func (r *affiliateRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&AffiliatePO{}, id).Error
}

func (r *affiliateRepo) toPO(affiliate *biz.Affiliate) *AffiliatePO {
	return &AffiliatePO{
		ID:         affiliate.ID,
		Name:       affiliate.Name,
		Code:       affiliate.Code,
		Contact:    affiliate.Contact,
		Website:    affiliate.Website,
		Commission: affiliate.Commission,
		Status:     affiliate.Status,
		CreatedAt:  affiliate.CreatedAt.Unix(),
		UpdatedAt:  affiliate.UpdatedAt.Unix(),
	}
}

func (r *affiliateRepo) toEntity(po *AffiliatePO) *biz.Affiliate {
	return &biz.Affiliate{
		ID:         po.ID,
		Name:       po.Name,
		Code:       po.Code,
		Contact:    po.Contact,
		Website:    po.Website,
		Commission: po.Commission,
		Status:     po.Status,
		CreatedAt:  unixToTime(po.CreatedAt),
		UpdatedAt:  unixToTime(po.UpdatedAt),
	}
}

// affiliateOrderRepo 推广订单仓储实现
type affiliateOrderRepo struct {
	db *gorm.DB
}

// NewAffiliateOrderRepo 创建推广订单仓储
func NewAffiliateOrderRepo(db *gorm.DB) biz.AffiliateOrderRepo {
	return &affiliateOrderRepo{db: db}
}

func (r *affiliateOrderRepo) ListByAffiliateID(ctx context.Context, affiliateID uint64, offset, limit int) ([]*biz.AffiliateOrder, error) {
	var pos []*AffiliateOrderPO
	if err := r.db.WithContext(ctx).Where("affiliate_id = ?", affiliateID).Offset(offset).Limit(limit).Order("id DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.AffiliateOrder, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *affiliateOrderRepo) toEntity(po *AffiliateOrderPO) *biz.AffiliateOrder {
	return &biz.AffiliateOrder{
		ID:          po.ID,
		AffiliateID: po.AffiliateID,
		OrderNo:     po.OrderNo,
		Amount:      po.Amount,
		Commission:  po.Commission,
		Status:      po.Status,
		CreatedAt:   unixToTime(po.CreatedAt),
		UpdatedAt:   unixToTime(po.UpdatedAt),
	}
}

// affiliateCustomerRepo 推广客户仓储实现
type affiliateCustomerRepo struct {
	db *gorm.DB
}

// NewAffiliateCustomerRepo 创建推广客户仓储
func NewAffiliateCustomerRepo(db *gorm.DB) biz.AffiliateCustomerRepo {
	return &affiliateCustomerRepo{db: db}
}

func (r *affiliateCustomerRepo) ListByAffiliateID(ctx context.Context, affiliateID uint64, offset, limit int) ([]*biz.AffiliateCustomer, error) {
	var pos []*AffiliateCustomerPO
	if err := r.db.WithContext(ctx).Where("affiliate_id = ?", affiliateID).Offset(offset).Limit(limit).Order("id DESC").Find(&pos).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.AffiliateCustomer, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.toEntity(po))
	}
	return result, nil
}

func (r *affiliateCustomerRepo) toEntity(po *AffiliateCustomerPO) *biz.AffiliateCustomer {
	return &biz.AffiliateCustomer{
		ID:          po.ID,
		AffiliateID: po.AffiliateID,
		CustomerID:  po.CustomerID,
		Source:      po.Source,
		FirstVisit:  unixToTime(po.FirstVisit),
		CreatedAt:   unixToTime(po.CreatedAt),
		UpdatedAt:   unixToTime(po.UpdatedAt),
	}
}

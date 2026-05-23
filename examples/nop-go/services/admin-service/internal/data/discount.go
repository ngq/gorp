// Package data 优惠模块数据层 —— 优惠与使用记录的持久化对象与仓储实现
package data

import (
	"context"

	"nop-go/services/admin-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== 持久化对象（PO） ====================

// DiscountPO 优惠持久化对象 —— 映射数据库 discounts 表
type DiscountPO struct {
	ID           uint    `gorm:"primaryKey" db:"id"`
	Name         string  `gorm:"column:name;type:varchar(100);not null" db:"name"`                // 优惠名称
	Code         string  `gorm:"column:code;type:varchar(100);uniqueIndex;not null" db:"code"`     // 优惠编码
	Type         int     `gorm:"column:type;type:tinyint;not null;index" db:"type"`               // 优惠类型
	Value        float64 `gorm:"column:value;type:decimal(10,2);not null" db:"value"`              // 优惠值
	MinAmount    float64 `gorm:"column:min_amount;type:decimal(10,2);default:0" db:"min_amount"`        // 最低消费
	MaxDiscount  float64 `gorm:"column:max_discount;type:decimal(10,2);default:0" db:"max_discount"`      // 最大优惠
	StartTime    string  `gorm:"column:start_time;type:datetime" db:"start_time"`                        // 开始时间
	EndTime      string  `gorm:"column:end_time;type:datetime" db:"end_time"`                          // 结束时间
	TotalQuota   int     `gorm:"column:total_quota;type:int;default:0" db:"total_quota"`                  // 总发放量
	UsedQuota    int     `gorm:"column:used_quota;type:int;default:0" db:"used_quota"`                   // 已使用量
	PerUserLimit int     `gorm:"column:per_user_limit;type:int;default:1" db:"per_user_limit"`               // 每人限领
	Status       int     `gorm:"column:status;type:tinyint;default:1;index" db:"status"`             // 状态
	Description  string  `gorm:"column:description;type:varchar(500)" db:"description"`                   // 描述
	CreatedAt    string  `gorm:"column:created_at" db:"created_at"`
	UpdatedAt    string  `gorm:"column:updated_at" db:"updated_at"`
}

// TableName 指定优惠表名
func (DiscountPO) TableName() string { return "discounts" }

// DiscountUsagePO 优惠使用记录持久化对象 —— 映射数据库 discount_usages 表
type DiscountUsagePO struct {
	ID         uint   `gorm:"primaryKey" db:"id"`
	DiscountID uint   `gorm:"column:discount_id;type:bigint;index;not null" db:"discount_id"` // 关联优惠ID
	UserID     uint   `gorm:"column:user_id;type:bigint;index;not null" db:"user_id"`     // 使用用户ID
	OrderNo    string `gorm:"column:order_no;type:varchar(50);index" db:"order_no"`        // 关联订单号
	UsedAt     string `gorm:"column:used_at;type:datetime" db:"used_at"`                  // 使用时间
	Status     int    `gorm:"column:status;type:tinyint;default:0" db:"status"`          // 状态
	CreatedAt  string `gorm:"column:created_at" db:"created_at"`
}

// TableName 指定使用记录表名
func (DiscountUsagePO) TableName() string { return "discount_usages" }

// ==================== PO ↔ Entity 转换 ====================

// toEntity 将 DiscountPO 转换为 biz.Discount 领域实体
func (d *DiscountPO) toEntity() *biz.Discount {
	return &biz.Discount{
		ID:           d.ID,
		Name:         d.Name,
		Code:         d.Code,
		Type:         d.Type,
		Value:        d.Value,
		MinAmount:    d.MinAmount,
		MaxDiscount:  d.MaxDiscount,
		StartTime:    d.StartTime,
		EndTime:      d.EndTime,
		TotalQuota:   d.TotalQuota,
		UsedQuota:    d.UsedQuota,
		PerUserLimit: d.PerUserLimit,
		Status:       d.Status,
		Description:  d.Description,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

// toEntity 将 DiscountUsagePO 转换为 biz.DiscountUsage 领域实体
func (u *DiscountUsagePO) toEntity() *biz.DiscountUsage {
	return &biz.DiscountUsage{
		ID:         u.ID,
		DiscountID: u.DiscountID,
		UserID:     u.UserID,
		OrderNo:    u.OrderNo,
		UsedAt:     u.UsedAt,
		Status:     u.Status,
		CreatedAt:  u.CreatedAt,
	}
}

// discountToPO 将 biz.Discount 领域实体转换为 DiscountPO
func discountToPO(d *biz.Discount) *DiscountPO {
	return &DiscountPO{
		ID:           d.ID,
		Name:         d.Name,
		Code:         d.Code,
		Type:         d.Type,
		Value:        d.Value,
		MinAmount:    d.MinAmount,
		MaxDiscount:  d.MaxDiscount,
		StartTime:    d.StartTime,
		EndTime:      d.EndTime,
		TotalQuota:   d.TotalQuota,
		UsedQuota:    d.UsedQuota,
		PerUserLimit: d.PerUserLimit,
		Status:       d.Status,
		Description:  d.Description,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

// discountUsageToPO 将 biz.DiscountUsage 领域实体转换为 DiscountUsagePO
func discountUsageToPO(u *biz.DiscountUsage) *DiscountUsagePO {
	return &DiscountUsagePO{
		ID:         u.ID,
		DiscountID: u.DiscountID,
		UserID:     u.UserID,
		OrderNo:    u.OrderNo,
		UsedAt:     u.UsedAt,
		Status:     u.Status,
		CreatedAt:  u.CreatedAt,
	}
}

// ==================== 仓储实现 ====================

// discountRepo 优惠仓储实现
type discountRepo struct {
	db *gorm.DB
}

// NewDiscountRepo 创建优惠仓储
func NewDiscountRepo(db *gorm.DB) biz.DiscountRepo {
	return &discountRepo{db: db}
}

func (r *discountRepo) Create(ctx context.Context, d *biz.Discount) error {
	po := discountToPO(d)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *discountRepo) GetByID(ctx context.Context, id uint) (*biz.Discount, error) {
	var po DiscountPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *discountRepo) GetByCode(ctx context.Context, code string) (*biz.Discount, error) {
	var po DiscountPO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *discountRepo) List(ctx context.Context, discountType, status int, page, pageSize int) ([]*biz.Discount, int64, error) {
	var pos []*DiscountPO
	var total int64
	q := r.db.WithContext(ctx).Model(&DiscountPO{})
	// 类型过滤
	if discountType > 0 {
		q = q.Where("type = ?", discountType)
	}
	// 状态过滤
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.Discount, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *discountRepo) Update(ctx context.Context, d *biz.Discount) error {
	po := discountToPO(d)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *discountRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&DiscountPO{}, id).Error
}

// discountUsageRepo 优惠使用记录仓储实现
type discountUsageRepo struct {
	db *gorm.DB
}

// NewDiscountUsageRepo 创建优惠使用记录仓储
func NewDiscountUsageRepo(db *gorm.DB) biz.DiscountUsageRepo {
	return &discountUsageRepo{db: db}
}

func (r *discountUsageRepo) Create(ctx context.Context, u *biz.DiscountUsage) error {
	po := discountUsageToPO(u)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *discountUsageRepo) GetByID(ctx context.Context, id uint) (*biz.DiscountUsage, error) {
	var po DiscountUsagePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *discountUsageRepo) List(ctx context.Context, discountID, userID uint, page, pageSize int) ([]*biz.DiscountUsage, int64, error) {
	var pos []*DiscountUsagePO
	var total int64
	q := r.db.WithContext(ctx).Model(&DiscountUsagePO{})
	// 优惠ID过滤
	if discountID > 0 {
		q = q.Where("discount_id = ?", discountID)
	}
	// 用户ID过滤
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.DiscountUsage, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *discountUsageRepo) GetUserUsageCount(ctx context.Context, discountID, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&DiscountUsagePO{}).
		Where("discount_id = ? AND user_id = ?", discountID, userID).
		Count(&count).Error
	return count, err
}

func (r *discountUsageRepo) Update(ctx context.Context, u *biz.DiscountUsage) error {
	po := discountUsageToPO(u)
	return r.db.WithContext(ctx).Save(po).Error
}

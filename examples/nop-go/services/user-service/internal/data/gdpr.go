// Package data 定义 GDPR 相关的数据访问层。
// 本文件合并了原 gdpr 服务的数据层：GDPR 请求持久化对象和仓储实现。
package data

import (
	"context"
	"time"

	"nop-go/services/user-service/internal/biz"

	"gorm.io/gorm"
)

// ======================== 持久化对象(PO) ========================

// GdprPO GDPR 请求持久化对象，映射数据库 gdprs 表
type GdprPO struct {
	ID          uint           `gorm:"primaryKey" db:"id" json:"id"`
	UserID      uint           `gorm:"index;not null" db:"user_id" json:"user_id"`
	RequestType string         `gorm:"size:32;not null" db:"request_type" json:"request_type"`  // delete 或 export
	Status      string         `gorm:"size:32;not null;default:pending" db:"status" json:"status"` // pending/processing/completed/rejected
	Reason      string         `gorm:"size:512" db:"reason" json:"reason"`
	ReviewedBy  *uint          `db:"reviewed_by" json:"reviewed_by"`
	ReviewedAt  *time.Time     `db:"reviewed_at" json:"reviewed_at"`
	CompletedAt *time.Time     `db:"completed_at" json:"completed_at"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" db:"deleted_at" json:"-"`
}

// TableName 指定 GDPR 表名
func (GdprPO) TableName() string { return "gdprs" }

// ======================== PO <-> 领域模型转换 ========================

// toGdpr 将 GdprPO 转换为 GDPR 领域模型
func (p *GdprPO) toGdpr() *biz.Gdpr {
	return &biz.Gdpr{
		ID:          p.ID,
		UserID:      p.UserID,
		RequestType: p.RequestType,
		Status:      p.Status,
		Reason:      p.Reason,
		ReviewedBy:  p.ReviewedBy,
		ReviewedAt:  p.ReviewedAt,
		CompletedAt: p.CompletedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// toGdprPO 将 GDPR 领域模型转换为 GdprPO
func toGdprPO(g *biz.Gdpr) *GdprPO {
	return &GdprPO{
		ID:          g.ID,
		UserID:      g.UserID,
		RequestType: g.RequestType,
		Status:      g.Status,
		Reason:      g.Reason,
		ReviewedBy:  g.ReviewedBy,
		ReviewedAt:  g.ReviewedAt,
		CompletedAt: g.CompletedAt,
	}
}

// ======================== 仓储实现 ========================

// gdprRepo GDPR 仓储实现
type gdprRepo struct {
	db *gorm.DB
}

// NewGdprRepo 创建 GDPR 仓储实例
func NewGdprRepo(db *gorm.DB) biz.GdprRepository {
	return &gdprRepo{db: db}
}

func (r *gdprRepo) Create(ctx context.Context, gdpr *biz.Gdpr) (*biz.Gdpr, error) {
	po := toGdprPO(gdpr)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	return po.toGdpr(), nil
}

func (r *gdprRepo) Update(ctx context.Context, gdpr *biz.Gdpr) (*biz.Gdpr, error) {
	po := toGdprPO(gdpr)
	if err := r.db.WithContext(ctx).Model(po).Updates(po).Error; err != nil {
		return nil, err
	}
	return po.toGdpr(), nil
}

func (r *gdprRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&GdprPO{}, id).Error
}

func (r *gdprRepo) GetByID(ctx context.Context, id uint) (*biz.Gdpr, error) {
	var po GdprPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toGdpr(), nil
}

func (r *gdprRepo) ListByUserID(ctx context.Context, userID uint) ([]*biz.Gdpr, error) {
	var pos []*GdprPO
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	gdprs := make([]*biz.Gdpr, 0, len(pos))
	for _, po := range pos {
		gdprs = append(gdprs, po.toGdpr())
	}
	return gdprs, nil
}

func (r *gdprRepo) List(ctx context.Context, offset, limit int) ([]*biz.Gdpr, int64, error) {
	var pos []*GdprPO
	var total int64
	// 先统计总数
	if err := r.db.WithContext(ctx).Model(&GdprPO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 再分页查询
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	gdprs := make([]*biz.Gdpr, 0, len(pos))
	for _, po := range pos {
		gdprs = append(gdprs, po.toGdpr())
	}
	return gdprs, total, nil
}

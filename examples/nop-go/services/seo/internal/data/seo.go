// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/seo/internal/biz"

	"gorm.io/gorm"
)

// SeoPO Seo持久化对象。
type SeoPO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;column:username" db:"username" json:"username"`
	Email     string    `gorm:"size:128;column:email" db:"email" json:"email"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (SeoPO) TableName() string {
	return "seos"
}

// ToEntity 转换为领域实体。
func (po *SeoPO) ToEntity() *biz.Seo {
	return &biz.Seo{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// seoRepo Seo仓储实现。
type seoRepo struct {
	db *gorm.DB
}

// NewSeoRepo 创建Seo仓储。
func NewSeoRepo(db *gorm.DB) biz.SeoRepository {
	return &seoRepo{db: db}
}

// Create 创建Seo。
func (r *seoRepo) Create(ctx context.Context, seo *biz.Seo) error {
	po := &SeoPO{
		Username:  seo.Username,
		Email:     seo.Email,
		CreatedAt: seo.CreatedAt,
		UpdatedAt: seo.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取Seo。
func (r *seoRepo) GetByID(ctx context.Context, id uint) (*biz.Seo, error) {
	var po SeoPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByUsername 根据ID获取Seo。
func (r *seoRepo) GetByUsername(ctx context.Context, username string) (*biz.Seo, error) {
	var po SeoPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取Seo列表。
func (r *seoRepo) List(ctx context.Context, page, size int) ([]*biz.Seo, int64, error) {
	var pos []SeoPO
	var total int64

	r.db.WithContext(ctx).Model(&SeoPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	seos := make([]*biz.Seo, len(pos))
	for i, po := range pos {
		seos[i] = po.ToEntity()
	}

	return seos, total, nil
}

// Delete 删除Seo。
func (r *seoRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&SeoPO{}, id).Error
}
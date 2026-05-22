// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/gdpr/internal/biz"

	"gorm.io/gorm"
)

// GdprPO Gdpr持久化对象。
type GdprPO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;column:username" db:"username" json:"username"`
	Email     string    `gorm:"size:128;column:email" db:"email" json:"email"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (GdprPO) TableName() string {
	return "gdprs"
}

// ToEntity 转换为领域实体。
func (po *GdprPO) ToEntity() *biz.Gdpr {
	return &biz.Gdpr{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// gdprRepo Gdpr仓储实现。
type gdprRepo struct {
	db *gorm.DB
}

// NewGdprRepo 创建Gdpr仓储。
func NewGdprRepo(db *gorm.DB) biz.GdprRepository {
	return &gdprRepo{db: db}
}

// Create 创建Gdpr。
func (r *gdprRepo) Create(ctx context.Context, gdpr *biz.Gdpr) error {
	po := &GdprPO{
		Username:  gdpr.Username,
		Email:     gdpr.Email,
		CreatedAt: gdpr.CreatedAt,
		UpdatedAt: gdpr.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取Gdpr。
func (r *gdprRepo) GetByID(ctx context.Context, id uint) (*biz.Gdpr, error) {
	var po GdprPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByUsername 根据ID获取Gdpr。
func (r *gdprRepo) GetByUsername(ctx context.Context, username string) (*biz.Gdpr, error) {
	var po GdprPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取Gdpr列表。
func (r *gdprRepo) List(ctx context.Context, page, size int) ([]*biz.Gdpr, int64, error) {
	var pos []GdprPO
	var total int64

	r.db.WithContext(ctx).Model(&GdprPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	gdprs := make([]*biz.Gdpr, len(pos))
	for i, po := range pos {
		gdprs[i] = po.ToEntity()
	}

	return gdprs, total, nil
}

// Delete 删除Gdpr。
func (r *gdprRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&GdprPO{}, id).Error
}
// Package data 数据访问层。
// 包含 SEO 元数据的持久化对象（PO）和仓储实现。
package data

import (
	"context"
	"time"

	"nop-go/services/catalog-service/internal/biz"

	"gorm.io/gorm"
)

// SeoPO SEO 元数据持久化对象。
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
func (po *SeoPO) ToEntity() *biz.SeoMeta {
	return &biz.SeoMeta{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// seoMetaRepo SEO 元数据仓储实现。
type seoMetaRepo struct {
	db *gorm.DB
}

// NewSeoMetaRepo 创建 SEO 元数据仓储。
func NewSeoMetaRepo(db *gorm.DB) biz.SeoMetaRepository {
	return &seoMetaRepo{db: db}
}

// Create 创建 SEO 元数据。
func (r *seoMetaRepo) Create(ctx context.Context, seo *biz.SeoMeta) error {
	po := &SeoPO{
		Username:  seo.Username,
		Email:     seo.Email,
		CreatedAt: seo.CreatedAt,
		UpdatedAt: seo.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取 SEO 元数据。
func (r *seoMetaRepo) GetByID(ctx context.Context, id uint) (*biz.SeoMeta, error) {
	var po SeoPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByUsername 根据用户名获取 SEO 元数据。
func (r *seoMetaRepo) GetByUsername(ctx context.Context, username string) (*biz.SeoMeta, error) {
	var po SeoPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取 SEO 元数据列表。
func (r *seoMetaRepo) List(ctx context.Context, page, size int) ([]*biz.SeoMeta, int64, error) {
	var pos []SeoPO
	var total int64

	r.db.WithContext(ctx).Model(&SeoPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	seos := make([]*biz.SeoMeta, len(pos))
	for i, po := range pos {
		seos[i] = po.ToEntity()
	}

	return seos, total, nil
}

// Delete 删除 SEO 元数据。
func (r *seoMetaRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&SeoPO{}, id).Error
}
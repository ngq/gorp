// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/store/internal/biz"

	"gorm.io/gorm"
)

// StorePO 店铺持久化对象。
type StorePO struct {
	ID           uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name         string    `gorm:"size:256;column:name" db:"name" json:"name"`
	Url          string    `gorm:"size:512;column:url" db:"url" json:"url"`
	SslEnabled   bool      `gorm:"column:ssl_enabled" db:"ssl_enabled" json:"ssl_enabled"`
	Hosts        string    `gorm:"size:1024;column:hosts" db:"hosts" json:"hosts"`
	DisplayOrder int       `gorm:"column:display_order" db:"display_order" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (StorePO) TableName() string {
	return "stores"
}

// ToEntity 转换为领域实体。
func (po *StorePO) ToEntity() *biz.Store {
	return &biz.Store{
		ID:           po.ID,
		Name:         po.Name,
		Url:          po.Url,
		SslEnabled:   po.SslEnabled,
		Hosts:        po.Hosts,
		DisplayOrder: po.DisplayOrder,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// storeRepo 店铺仓储实现。
type storeRepo struct {
	db *gorm.DB
}

// NewStoreRepo 创建店铺仓储。
func NewStoreRepo(db *gorm.DB) biz.StoreRepository {
	return &storeRepo{db: db}
}

// Create 创建店铺。
func (r *storeRepo) Create(ctx context.Context, store *biz.Store) error {
	po := &StorePO{
		Name:         store.Name,
		Url:          store.Url,
		SslEnabled:   store.SslEnabled,
		Hosts:        store.Hosts,
		DisplayOrder: store.DisplayOrder,
		CreatedAt:    store.CreatedAt,
		UpdatedAt:    store.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取店铺。
func (r *storeRepo) GetByID(ctx context.Context, id uint) (*biz.Store, error) {
	var po StorePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取店铺列表。
func (r *storeRepo) List(ctx context.Context, page, size int) ([]*biz.Store, int64, error) {
	var pos []StorePO
	var total int64

	r.db.WithContext(ctx).Model(&StorePO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	stores := make([]*biz.Store, len(pos))
	for i, po := range pos {
		stores[i] = po.ToEntity()
	}

	return stores, total, nil
}

// Update 更新店铺。
func (r *storeRepo) Update(ctx context.Context, store *biz.Store) error {
	return r.db.WithContext(ctx).Model(&StorePO{}).Where("id = ?", store.ID).Updates(map[string]interface{}{
		"name":          store.Name,
		"url":           store.Url,
		"ssl_enabled":   store.SslEnabled,
		"hosts":         store.Hosts,
		"display_order": store.DisplayOrder,
		"updated_at":    store.UpdatedAt,
	}).Error
}

// Delete 删除店铺。
func (r *storeRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&StorePO{}, id).Error
}
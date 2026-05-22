// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/plugin/internal/biz"

	"gorm.io/gorm"
)

// PluginPO Plugin持久化对象。
type PluginPO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;column:username" db:"username" json:"username"`
	Email     string    `gorm:"size:128;column:email" db:"email" json:"email"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (PluginPO) TableName() string {
	return "plugins"
}

// ToEntity 转换为领域实体。
func (po *PluginPO) ToEntity() *biz.Plugin {
	return &biz.Plugin{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// pluginRepo Plugin仓储实现。
type pluginRepo struct {
	db *gorm.DB
}

// NewPluginRepo 创建Plugin仓储。
func NewPluginRepo(db *gorm.DB) biz.PluginRepository {
	return &pluginRepo{db: db}
}

// Create 创建Plugin。
func (r *pluginRepo) Create(ctx context.Context, plugin *biz.Plugin) error {
	po := &PluginPO{
		Username:  plugin.Username,
		Email:     plugin.Email,
		CreatedAt: plugin.CreatedAt,
		UpdatedAt: plugin.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取Plugin。
func (r *pluginRepo) GetByID(ctx context.Context, id uint) (*biz.Plugin, error) {
	var po PluginPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByUsername 根据ID获取Plugin。
func (r *pluginRepo) GetByUsername(ctx context.Context, username string) (*biz.Plugin, error) {
	var po PluginPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取Plugin列表。
func (r *pluginRepo) List(ctx context.Context, page, size int) ([]*biz.Plugin, int64, error) {
	var pos []PluginPO
	var total int64

	r.db.WithContext(ctx).Model(&PluginPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	plugins := make([]*biz.Plugin, len(pos))
	for i, po := range pos {
		plugins[i] = po.ToEntity()
	}

	return plugins, total, nil
}

// Delete 删除Plugin。
func (r *pluginRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PluginPO{}, id).Error
}
// Package data 插件模块数据层 —— 插件的持久化对象与仓储实现
package data

import (
	"context"

	"nop-go/services/admin-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== 持久化对象（PO） ====================

// PluginPO 插件持久化对象 —— 映射数据库 plugins 表
type PluginPO struct {
	ID          uint   `gorm:"primaryKey" db:"id"`
	Name        string `gorm:"column:name;type:varchar(100);not null" db:"name"`            // 插件名称
	Code        string `gorm:"column:code;type:varchar(100);uniqueIndex;not null" db:"code"` // 插件编码
	Version     string `gorm:"column:version;type:varchar(20)" db:"version"`                    // 版本号
	Description string `gorm:"column:description;type:varchar(500)" db:"description"`               // 插件描述
	Author      string `gorm:"column:author;type:varchar(100)" db:"author"`                    // 作者
	Config      string `gorm:"column:config;type:text" db:"config"`                            // 插件配置（JSON）
	Status      int    `gorm:"column:status;type:tinyint;default:0;index" db:"status"`         // 状态
	Sort        int    `gorm:"column:sort;type:int;default:0" db:"sort"`                     // 排序权重
	CreatedAt   string `gorm:"column:created_at" db:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at" db:"updated_at"`
}

// TableName 指定插件表名
func (PluginPO) TableName() string { return "plugins" }

// ==================== PO ↔ Entity 转换 ====================

// toEntity 将 PluginPO 转换为 biz.Plugin 领域实体
func (p *PluginPO) toEntity() *biz.Plugin {
	return &biz.Plugin{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Version:     p.Version,
		Description: p.Description,
		Author:      p.Author,
		Config:      p.Config,
		Status:      p.Status,
		Sort:        p.Sort,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// pluginToPO 将 biz.Plugin 领域实体转换为 PluginPO
func pluginToPO(p *biz.Plugin) *PluginPO {
	return &PluginPO{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Version:     p.Version,
		Description: p.Description,
		Author:      p.Author,
		Config:      p.Config,
		Status:      p.Status,
		Sort:        p.Sort,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ==================== 仓储实现 ====================

// pluginRepo 插件仓储实现
type pluginRepo struct {
	db *gorm.DB
}

// NewPluginRepo 创建插件仓储
func NewPluginRepo(db *gorm.DB) biz.PluginRepo {
	return &pluginRepo{db: db}
}

func (r *pluginRepo) Create(ctx context.Context, p *biz.Plugin) error {
	po := pluginToPO(p)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *pluginRepo) GetByID(ctx context.Context, id uint) (*biz.Plugin, error) {
	var po PluginPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *pluginRepo) GetByCode(ctx context.Context, code string) (*biz.Plugin, error) {
	var po PluginPO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *pluginRepo) List(ctx context.Context, status int, page, pageSize int) ([]*biz.Plugin, int64, error) {
	var pos []*PluginPO
	var total int64
	q := r.db.WithContext(ctx).Model(&PluginPO{})
	// 状态过滤：status < 0 表示不过滤
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.Plugin, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

func (r *pluginRepo) Update(ctx context.Context, p *biz.Plugin) error {
	po := pluginToPO(p)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *pluginRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PluginPO{}, id).Error
}

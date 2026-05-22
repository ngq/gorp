// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/logging/internal/biz"

	"gorm.io/gorm"
)

// LoggingPO Logging持久化对象。
type LoggingPO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex;column:username" db:"username" json:"username"`
	Email     string    `gorm:"size:128;column:email" db:"email" json:"email"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (LoggingPO) TableName() string {
	return "loggings"
}

// ToEntity 转换为领域实体。
func (po *LoggingPO) ToEntity() *biz.Logging {
	return &biz.Logging{
		ID:        po.ID,
		Username:  po.Username,
		Email:     po.Email,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// loggingRepo Logging仓储实现。
type loggingRepo struct {
	db *gorm.DB
}

// NewLoggingRepo 创建Logging仓储。
func NewLoggingRepo(db *gorm.DB) biz.LoggingRepository {
	return &loggingRepo{db: db}
}

// Create 创建Logging。
func (r *loggingRepo) Create(ctx context.Context, logging *biz.Logging) error {
	po := &LoggingPO{
		Username:  logging.Username,
		Email:     logging.Email,
		CreatedAt: logging.CreatedAt,
		UpdatedAt: logging.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取Logging。
func (r *loggingRepo) GetByID(ctx context.Context, id uint) (*biz.Logging, error) {
	var po LoggingPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// GetByUsername 根据ID获取Logging。
func (r *loggingRepo) GetByUsername(ctx context.Context, username string) (*biz.Logging, error) {
	var po LoggingPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取Logging列表。
func (r *loggingRepo) List(ctx context.Context, page, size int) ([]*biz.Logging, int64, error) {
	var pos []LoggingPO
	var total int64

	r.db.WithContext(ctx).Model(&LoggingPO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	loggings := make([]*biz.Logging, len(pos))
	for i, po := range pos {
		loggings[i] = po.ToEntity()
	}

	return loggings, total, nil
}

// Delete 删除Logging。
func (r *loggingRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&LoggingPO{}, id).Error
}
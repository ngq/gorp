// Package data 数据访问层。
// 包含媒体实体的持久化对象（PO）和仓储实现。
package data

import (
	"context"
	"time"

	"nop-go/services/catalog-service/internal/biz"

	"gorm.io/gorm"
)

// MediaPO 媒体持久化对象。
type MediaPO struct {
	ID        uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	FileName  string    `gorm:"size:256;column:file_name" db:"file_name" json:"file_name"`
	MimeType  string    `gorm:"size:64;column:mime_type" db:"mime_type" json:"mime_type"`
	FileSize  int64     `gorm:"column:file_size" db:"file_size" json:"file_size"`
	FileURL   string    `gorm:"size:512;column:file_url" db:"file_url" json:"file_url"`
	AltText   string    `gorm:"size:256;column:alt_text" db:"alt_text" json:"alt_text"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (MediaPO) TableName() string {
	return "medias"
}

// ToEntity 转换为领域实体。
func (po *MediaPO) ToEntity() *biz.Media {
	return &biz.Media{
		ID:        po.ID,
		FileName:  po.FileName,
		MimeType:  po.MimeType,
		FileSize:  po.FileSize,
		FileURL:   po.FileURL,
		AltText:   po.AltText,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

// mediaRepo 媒体仓储实现。
type mediaRepo struct {
	db *gorm.DB
}

// NewMediaRepo 创建媒体仓储。
func NewMediaRepo(db *gorm.DB) biz.MediaRepository {
	return &mediaRepo{db: db}
}

// Create 创建媒体记录。
func (r *mediaRepo) Create(ctx context.Context, media *biz.Media) error {
	po := &MediaPO{
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取媒体。
func (r *mediaRepo) GetByID(ctx context.Context, id uint) (*biz.Media, error) {
	var po MediaPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// Delete 删除媒体。
func (r *mediaRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&MediaPO{}, id).Error
}
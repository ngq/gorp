// Package biz 业务逻辑层。
// 定义媒体服务（media）的领域实体、仓储接口和用例。
// 包含媒体文件上传、查询、删除等核心能力。
package biz

import (
	"context"
	"time"
)

// Media 媒体领域实体。
type Media struct {
	ID        uint      // 媒体ID
	FileName  string    // 文件名
	MimeType  string    // MIME类型
	FileSize  int64     // 文件大小（字节）
	FileURL   string    // 文件存储URL
	AltText   string    // 图片替代文本
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// MediaRepository 媒体仓储接口。
type MediaRepository interface {
	// Create 创建媒体记录
	Create(ctx context.Context, media *Media) error
	// GetByID 根据ID获取媒体
	GetByID(ctx context.Context, id uint) (*Media, error)
	// Delete 删除媒体
	Delete(ctx context.Context, id uint) error
}

// MediaUseCase 媒体用例。
type MediaUseCase struct {
	repo MediaRepository
}

// NewMediaUseCase 创建媒体用例。
func NewMediaUseCase(repo MediaRepository) *MediaUseCase {
	return &MediaUseCase{repo: repo}
}

// Upload 异步上传图片，创建媒体记录。
func (uc *MediaUseCase) Upload(ctx context.Context, fileName, mimeType string, fileSize int64, fileURL, altText string) (*Media, error) {
	media := &Media{
		FileName:  fileName,
		MimeType:  mimeType,
		FileSize:  fileSize,
		FileURL:   fileURL,
		AltText:   altText,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, media); err != nil {
		return nil, err
	}
	return media, nil
}

// GetByID 根据ID获取媒体。
func (uc *MediaUseCase) GetByID(ctx context.Context, id uint) (*Media, error) {
	return uc.repo.GetByID(ctx, id)
}

// Delete 删除媒体。
func (uc *MediaUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

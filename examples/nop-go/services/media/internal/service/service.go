package service

import (
	"context"
	"time"

	"nop-go/services/media/internal/biz"
	"nop-go/services/media/internal/data"

	"gorm.io/gorm"
)

// Services 媒体服务集合。
type Services struct {
	Media *MediaService
}

// NewServices 创建媒体服务集合。
func NewServices(db *gorm.DB) *Services {
	mediaRepo := data.NewMediaRepo(db)
	mediaUC := biz.NewMediaUseCase(mediaRepo)
	return &Services{
		Media: &MediaService{uc: mediaUC},
	}
}

// MediaService 媒体服务。
type MediaService struct {
	uc *biz.MediaUseCase
}

// UploadMediaRequest 上传图片请求。
type UploadMediaRequest struct {
	FileName string `json:"file_name" binding:"required"` // 文件名
	MimeType string `json:"mime_type" binding:"required"` // MIME类型
	FileSize int64  `json:"file_size" binding:"required"` // 文件大小
	FileURL  string `json:"file_url" binding:"required"`  // 文件存储URL
	AltText  string `json:"alt_text"`                     // 图片替代文本
}

// MediaResponse 媒体响应。
type MediaResponse struct {
	ID        uint   `json:"id"`         // 媒体ID
	FileName  string `json:"file_name"`  // 文件名
	MimeType  string `json:"mime_type"`  // MIME类型
	FileSize  int64  `json:"file_size"`  // 文件大小
	FileURL   string `json:"file_url"`   // 文件存储URL
	AltText   string `json:"alt_text"`   // 图片替代文本
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// Upload 异步上传图片。
func (s *MediaService) Upload(ctx context.Context, req UploadMediaRequest) (*MediaResponse, error) {
	media, err := s.uc.Upload(ctx, req.FileName, req.MimeType, req.FileSize, req.FileURL, req.AltText)
	if err != nil {
		return nil, err
	}
	return &MediaResponse{
		ID:        media.ID,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
	}, nil
}

// GetByID 根据ID获取媒体。
func (s *MediaService) GetByID(ctx context.Context, id uint) (*MediaResponse, error) {
	media, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &MediaResponse{
		ID:        media.ID,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
	}, nil
}

// Delete 删除媒体。
func (s *MediaService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

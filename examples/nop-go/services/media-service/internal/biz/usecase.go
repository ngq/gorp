// Package biz 媒体服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nop-go/services/media-service/internal/data"
	"nop-go/services/media-service/internal/models"
)

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string
	LocalPath  string
	URLPrefix  string
	S3Bucket   string
	S3Region   string
	S3AccessKey string
	S3SecretKey string
}

// MediaUseCase 媒体用例
type MediaUseCase struct {
	pictureRepo  data.PictureRepository
	productPicRepo data.ProductPictureRepository
	categoryPicRepo data.CategoryPictureRepository
	documentRepo data.DocumentRepository
	storageConfig StorageConfig
}

// NewMediaUseCase 创建媒体用例
func NewMediaUseCase(
	pictureRepo data.PictureRepository,
	productPicRepo data.ProductPictureRepository,
	categoryPicRepo data.CategoryPictureRepository,
	documentRepo data.DocumentRepository,
	storageConfig StorageConfig,
) *MediaUseCase {
	return &MediaUseCase{
		pictureRepo:    pictureRepo,
		productPicRepo: productPicRepo,
		categoryPicRepo: categoryPicRepo,
		documentRepo:   documentRepo,
		storageConfig:  storageConfig,
	}
}

// UploadPicture 上传图片
func (uc *MediaUseCase) UploadPicture(ctx context.Context, file *multipart.FileHeader, req *models.PictureUploadRequest) (*models.UploadResult, error) {
	// 验证文件类型
	if !uc.isImageFile(file.Filename) {
		return nil, data.ErrInvalidFileType
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 生成存储路径
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	relativePath := fmt.Sprintf("%s/%s", time.Now().Format("2006/01/02"), filename)
	fullPath := filepath.Join(uc.storageConfig.LocalPath, relativePath)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	// 获取 MIME 类型
	mimeType := uc.getMimeType(ext)

	// 创建图片记录
	picture := &models.Picture{
		MimeType:       mimeType,
		SeoFilename:    req.SeoFilename,
		AltAttribute:   req.AltAttribute,
		TitleAttribute: req.TitleAttribute,
		IsNew:          true,
		Path:           relativePath,
		Size:           file.Size,
	}

	if err := uc.pictureRepo.Create(ctx, picture); err != nil {
		// 删除已上传的文件
		os.Remove(fullPath)
		return nil, err
	}

	// 如果指定了实体关联
	if req.EntityType != "" && req.EntityID > 0 {
		if err := uc.linkPictureToEntity(ctx, picture.ID, req); err != nil {
			return nil, err
		}
	}

	return &models.UploadResult{
		ID:       picture.ID,
		URL:      fmt.Sprintf("%s/%s", uc.storageConfig.URLPrefix, relativePath),
		MimeType: mimeType,
		Size:     file.Size,
	}, nil
}

// GetPicture 获取图片
func (uc *MediaUseCase) GetPicture(ctx context.Context, id uint) (*models.Picture, error) {
	return uc.pictureRepo.GetByID(ctx, id)
}

// ListPictures 图片列表
func (uc *MediaUseCase) ListPictures(ctx context.Context, page, pageSize int) ([]*models.Picture, int64, error) {
	return uc.pictureRepo.List(ctx, page, pageSize)
}

// DeletePicture 删除图片
func (uc *MediaUseCase) DeletePicture(ctx context.Context, id uint) error {
	picture, err := uc.pictureRepo.GetByID(ctx, id)
	if err != nil {
		return data.ErrPictureNotFound
	}

	// 删除文件
	fullPath := filepath.Join(uc.storageConfig.LocalPath, picture.Path)
	os.Remove(fullPath)

	return uc.pictureRepo.Delete(ctx, id)
}

// LinkPictureToEntity 将图片关联到实体
func (uc *MediaUseCase) linkPictureToEntity(ctx context.Context, pictureID uint, req *models.PictureUploadRequest) error {
	switch req.EntityType {
	case "product":
		return uc.productPicRepo.Create(ctx, &models.ProductPicture{
			ProductID:    req.EntityID,
			PictureID:    pictureID,
			DisplayOrder: req.DisplayOrder,
			IsMain:       req.IsMain,
		})
	case "category":
		return uc.categoryPicRepo.Create(ctx, &models.CategoryPicture{
			CategoryID:   req.EntityID,
			PictureID:    pictureID,
			DisplayOrder: req.DisplayOrder,
		})
	}
	return errors.New("unsupported entity type")
}

// GetProductPictures 获取商品图片
func (uc *MediaUseCase) GetProductPictures(ctx context.Context, productID uint) ([]*models.ProductPicture, error) {
	return uc.productPicRepo.GetByProductID(ctx, productID)
}

// DeleteProductPicture 删除商品图片关联
func (uc *MediaUseCase) DeleteProductPicture(ctx context.Context, productID, pictureID uint) error {
	return uc.productPicRepo.Delete(ctx, productID, pictureID)
}

// UploadDocument 上传文档
func (uc *MediaUseCase) UploadDocument(ctx context.Context, file *multipart.FileHeader, req *models.DocumentUploadRequest) (*models.Document, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 生成存储路径
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	relativePath := fmt.Sprintf("documents/%s/%s", time.Now().Format("2006/01/02"), filename)
	fullPath := filepath.Join(uc.storageConfig.LocalPath, relativePath)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	// 获取 MIME 类型
	mimeType := uc.getMimeType(ext)

	// 创建文档记录
	doc := &models.Document{
		Name:        req.Name,
		Description: req.Description,
		MimeType:    mimeType,
		Path:        relativePath,
		Size:        file.Size,
	}

	if err := uc.documentRepo.Create(ctx, doc); err != nil {
		os.Remove(fullPath)
		return nil, err
	}

	return doc, nil
}

// GetDocument 获取文档
func (uc *MediaUseCase) GetDocument(ctx context.Context, id uint) (*models.Document, error) {
	return uc.documentRepo.GetByID(ctx, id)
}

// ListDocuments 文档列表
func (uc *MediaUseCase) ListDocuments(ctx context.Context, page, pageSize int) ([]*models.Document, int64, error) {
	return uc.documentRepo.List(ctx, page, pageSize)
}

// DeleteDocument 删除文档
func (uc *MediaUseCase) DeleteDocument(ctx context.Context, id uint) error {
	doc, err := uc.documentRepo.GetByID(ctx, id)
	if err != nil {
		return data.ErrDocumentNotFound
	}

	// 删除文件
	fullPath := filepath.Join(uc.storageConfig.LocalPath, doc.Path)
	os.Remove(fullPath)

	return uc.documentRepo.Delete(ctx, id)
}

// 辅助方法
func (uc *MediaUseCase) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".svg": true, ".ico": true,
	}
	return imageExts[ext]
}

func (uc *MediaUseCase) getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".zip":  "application/zip",
		".txt":  "text/plain",
	}
	if mt, ok := mimeTypes[strings.ToLower(ext)]; ok {
		return mt
	}
	return "application/octet-stream"
}
// Package data 媒体服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/media-service/internal/models"

	"gorm.io/gorm"
)

// PictureRepository 图片仓储接口
type PictureRepository interface {
	Create(ctx context.Context, picture *models.Picture) error
	GetByID(ctx context.Context, id uint) (*models.Picture, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Picture, int64, error)
	Update(ctx context.Context, picture *models.Picture) error
	Delete(ctx context.Context, id uint) error
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Picture, error)
}

type pictureRepository struct {
	db *gorm.DB
}

func NewPictureRepository(db *gorm.DB) PictureRepository {
	return &pictureRepository{db: db}
}

func (r *pictureRepository) Create(ctx context.Context, picture *models.Picture) error {
	return r.db.WithContext(ctx).Create(picture).Error
}

func (r *pictureRepository) GetByID(ctx context.Context, id uint) (*models.Picture, error) {
	var picture models.Picture
	err := r.db.WithContext(ctx).First(&picture, id).Error
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func (r *pictureRepository) List(ctx context.Context, page, pageSize int) ([]*models.Picture, int64, error) {
	var pictures []*models.Picture
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Picture{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&pictures).Error; err != nil {
		return nil, 0, err
	}

	return pictures, total, nil
}

func (r *pictureRepository) Update(ctx context.Context, picture *models.Picture) error {
	return r.db.WithContext(ctx).Save(picture).Error
}

func (r *pictureRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Picture{}, id).Error
}

func (r *pictureRepository) GetByIDs(ctx context.Context, ids []uint) ([]*models.Picture, error) {
	var pictures []*models.Picture
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&pictures).Error
	return pictures, err
}

// ProductPictureRepository 商品图片仓储
type ProductPictureRepository interface {
	Create(ctx context.Context, pp *models.ProductPicture) error
	GetByProductID(ctx context.Context, productID uint) ([]*models.ProductPicture, error)
	DeleteByProductID(ctx context.Context, productID uint) error
	Delete(ctx context.Context, productID, pictureID uint) error
}

type productPictureRepository struct {
	db *gorm.DB
}

func NewProductPictureRepository(db *gorm.DB) ProductPictureRepository {
	return &productPictureRepository{db: db}
}

func (r *productPictureRepository) Create(ctx context.Context, pp *models.ProductPicture) error {
	return r.db.WithContext(ctx).Create(pp).Error
}

func (r *productPictureRepository) GetByProductID(ctx context.Context, productID uint) ([]*models.ProductPicture, error) {
	var pps []*models.ProductPicture
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("display_order").Find(&pps).Error
	return pps, err
}

func (r *productPictureRepository) DeleteByProductID(ctx context.Context, productID uint) error {
	return r.db.WithContext(ctx).Where("product_id = ?", productID).Delete(&models.ProductPicture{}).Error
}

func (r *productPictureRepository) Delete(ctx context.Context, productID, pictureID uint) error {
	return r.db.WithContext(ctx).Where("product_id = ? AND picture_id = ?", productID, pictureID).Delete(&models.ProductPicture{}).Error
}

// CategoryPictureRepository 分类图片仓储
type CategoryPictureRepository interface {
	Create(ctx context.Context, cp *models.CategoryPicture) error
	GetByCategoryID(ctx context.Context, categoryID uint) ([]*models.CategoryPicture, error)
	Delete(ctx context.Context, categoryID, pictureID uint) error
}

type categoryPictureRepository struct {
	db *gorm.DB
}

func NewCategoryPictureRepository(db *gorm.DB) CategoryPictureRepository {
	return &categoryPictureRepository{db: db}
}

func (r *categoryPictureRepository) Create(ctx context.Context, cp *models.CategoryPicture) error {
	return r.db.WithContext(ctx).Create(cp).Error
}

func (r *categoryPictureRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]*models.CategoryPicture, error) {
	var cps []*models.CategoryPicture
	err := r.db.WithContext(ctx).Where("category_id = ?", categoryID).Order("display_order").Find(&cps).Error
	return cps, err
}

func (r *categoryPictureRepository) Delete(ctx context.Context, categoryID, pictureID uint) error {
	return r.db.WithContext(ctx).Where("category_id = ? AND picture_id = ?", categoryID, pictureID).Delete(&models.CategoryPicture{}).Error
}

// DocumentRepository 文档仓储
type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id uint) (*models.Document, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Document, int64, error)
	Delete(ctx context.Context, id uint) error
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, doc *models.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *documentRepository) GetByID(ctx context.Context, id uint) (*models.Document, error) {
	var doc models.Document
	err := r.db.WithContext(ctx).First(&doc, id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) List(ctx context.Context, page, pageSize int) ([]*models.Document, int64, error) {
	var docs []*models.Document
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Document{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (r *documentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Document{}, id).Error
}

// 常见错误
var (
	ErrPictureNotFound  = errors.New("picture not found")
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidFileType  = errors.New("invalid file type")
)
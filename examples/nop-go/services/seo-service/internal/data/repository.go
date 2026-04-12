// Package data SEO服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/seo-service/internal/models"

	"gorm.io/gorm"
)

// UrlRecordRepository URL记录仓储接口
type UrlRecordRepository interface {
	Create(ctx context.Context, record *models.UrlRecord) error
	GetByID(ctx context.Context, id uint) (*models.UrlRecord, error)
	GetBySlug(ctx context.Context, slug string) (*models.UrlRecord, error)
	GetByEntity(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.UrlRecord, error)
	List(ctx context.Context, page, pageSize int) ([]*models.UrlRecord, int64, error)
	ListByEntity(ctx context.Context, entityID uint, entityType string) ([]*models.UrlRecord, error)
	Update(ctx context.Context, record *models.UrlRecord) error
	Delete(ctx context.Context, id uint) error
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*models.UrlRecord, int64, error)
}

type urlRecordRepository struct {
	db *gorm.DB
}

func NewUrlRecordRepository(db *gorm.DB) UrlRecordRepository {
	return &urlRecordRepository{db: db}
}

func (r *urlRecordRepository) Create(ctx context.Context, record *models.UrlRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *urlRecordRepository) GetByID(ctx context.Context, id uint) (*models.UrlRecord, error) {
	var record models.UrlRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *urlRecordRepository) GetBySlug(ctx context.Context, slug string) (*models.UrlRecord, error) {
	var record models.UrlRecord
	err := r.db.WithContext(ctx).Where("slug = ? AND is_active = ?", slug, true).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *urlRecordRepository) GetByEntity(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.UrlRecord, error) {
	var record models.UrlRecord
	err := r.db.WithContext(ctx).Where("entity_id = ? AND entity_type = ? AND language_id = ? AND is_active = ?",
		entityID, entityType, languageID, true).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *urlRecordRepository) List(ctx context.Context, page, pageSize int) ([]*models.UrlRecord, int64, error) {
	var records []*models.UrlRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&models.UrlRecord{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *urlRecordRepository) ListByEntity(ctx context.Context, entityID uint, entityType string) ([]*models.UrlRecord, error) {
	var records []*models.UrlRecord
	err := r.db.WithContext(ctx).Where("entity_id = ? AND entity_type = ?", entityID, entityType).Find(&records).Error
	return records, err
}

func (r *urlRecordRepository) Update(ctx context.Context, record *models.UrlRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *urlRecordRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.UrlRecord{}, id).Error
}

func (r *urlRecordRepository) Search(ctx context.Context, keyword string, page, pageSize int) ([]*models.UrlRecord, int64, error) {
	var records []*models.UrlRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&models.UrlRecord{})
	if keyword != "" {
		db = db.Where("slug LIKE ? OR entity_type LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// UrlRedirectRepository URL重定向仓储接口
type UrlRedirectRepository interface {
	Create(ctx context.Context, redirect *models.UrlRedirect) error
	GetByID(ctx context.Context, id uint) (*models.UrlRedirect, error)
	GetByOldSlug(ctx context.Context, oldSlug string) (*models.UrlRedirect, error)
	List(ctx context.Context, page, pageSize int) ([]*models.UrlRedirect, int64, error)
	Update(ctx context.Context, redirect *models.UrlRedirect) error
	Delete(ctx context.Context, id uint) error
	IncrementHitCount(ctx context.Context, id uint) error
}

type urlRedirectRepository struct {
	db *gorm.DB
}

func NewUrlRedirectRepository(db *gorm.DB) UrlRedirectRepository {
	return &urlRedirectRepository{db: db}
}

func (r *urlRedirectRepository) Create(ctx context.Context, redirect *models.UrlRedirect) error {
	return r.db.WithContext(ctx).Create(redirect).Error
}

func (r *urlRedirectRepository) GetByID(ctx context.Context, id uint) (*models.UrlRedirect, error) {
	var redirect models.UrlRedirect
	err := r.db.WithContext(ctx).First(&redirect, id).Error
	if err != nil {
		return nil, err
	}
	return &redirect, nil
}

func (r *urlRedirectRepository) GetByOldSlug(ctx context.Context, oldSlug string) (*models.UrlRedirect, error) {
	var redirect models.UrlRedirect
	err := r.db.WithContext(ctx).Where("old_slug = ? AND is_active = ?", oldSlug, true).First(&redirect).Error
	if err != nil {
		return nil, err
	}
	return &redirect, nil
}

func (r *urlRedirectRepository) List(ctx context.Context, page, pageSize int) ([]*models.UrlRedirect, int64, error) {
	var redirects []*models.UrlRedirect
	var total int64

	db := r.db.WithContext(ctx).Model(&models.UrlRedirect{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&redirects).Error; err != nil {
		return nil, 0, err
	}

	return redirects, total, nil
}

func (r *urlRedirectRepository) Update(ctx context.Context, redirect *models.UrlRedirect) error {
	return r.db.WithContext(ctx).Save(redirect).Error
}

func (r *urlRedirectRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.UrlRedirect{}, id).Error
}

func (r *urlRedirectRepository) IncrementHitCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.UrlRedirect{}).Where("id = ?", id).
		UpdateColumn("hit_count", gorm.Expr("hit_count + 1")).Error
}

// MetaInfoRepository 元信息仓储接口
type MetaInfoRepository interface {
	Create(ctx context.Context, meta *models.MetaInfo) error
	GetByID(ctx context.Context, id uint) (*models.MetaInfo, error)
	GetByEntity(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.MetaInfo, error)
	List(ctx context.Context, page, pageSize int) ([]*models.MetaInfo, int64, error)
	Update(ctx context.Context, meta *models.MetaInfo) error
	Delete(ctx context.Context, id uint) error
}

type metaInfoRepository struct {
	db *gorm.DB
}

func NewMetaInfoRepository(db *gorm.DB) MetaInfoRepository {
	return &metaInfoRepository{db: db}
}

func (r *metaInfoRepository) Create(ctx context.Context, meta *models.MetaInfo) error {
	return r.db.WithContext(ctx).Create(meta).Error
}

func (r *metaInfoRepository) GetByID(ctx context.Context, id uint) (*models.MetaInfo, error) {
	var meta models.MetaInfo
	err := r.db.WithContext(ctx).First(&meta, id).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (r *metaInfoRepository) GetByEntity(ctx context.Context, entityID uint, entityType string, languageID uint) (*models.MetaInfo, error) {
	var meta models.MetaInfo
	err := r.db.WithContext(ctx).Where("entity_id = ? AND entity_type = ? AND language_id = ?",
		entityID, entityType, languageID).First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (r *metaInfoRepository) List(ctx context.Context, page, pageSize int) ([]*models.MetaInfo, int64, error) {
	var metas []*models.MetaInfo
	var total int64

	db := r.db.WithContext(ctx).Model(&models.MetaInfo{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&metas).Error; err != nil {
		return nil, 0, err
	}

	return metas, total, nil
}

func (r *metaInfoRepository) Update(ctx context.Context, meta *models.MetaInfo) error {
	return r.db.WithContext(ctx).Save(meta).Error
}

func (r *metaInfoRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MetaInfo{}, id).Error
}

// SitemapNodeRepository Sitemap节点仓储接口
type SitemapNodeRepository interface {
	Create(ctx context.Context, node *models.SitemapNode) error
	GetByID(ctx context.Context, id uint) (*models.SitemapNode, error)
	List(ctx context.Context, page, pageSize int) ([]*models.SitemapNode, int64, error)
	ListByEntityType(ctx context.Context, entityType string) ([]*models.SitemapNode, error)
	Update(ctx context.Context, node *models.SitemapNode) error
	Delete(ctx context.Context, id uint) error
	DeleteByEntityType(ctx context.Context, entityType string) error
	GetAll(ctx context.Context) ([]*models.SitemapNode, error)
	Count(ctx context.Context) (int64, error)
}

type sitemapNodeRepository struct {
	db *gorm.DB
}

func NewSitemapNodeRepository(db *gorm.DB) SitemapNodeRepository {
	return &sitemapNodeRepository{db: db}
}

func (r *sitemapNodeRepository) Create(ctx context.Context, node *models.SitemapNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *sitemapNodeRepository) GetByID(ctx context.Context, id uint) (*models.SitemapNode, error) {
	var node models.SitemapNode
	err := r.db.WithContext(ctx).First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *sitemapNodeRepository) List(ctx context.Context, page, pageSize int) ([]*models.SitemapNode, int64, error) {
	var nodes []*models.SitemapNode
	var total int64

	db := r.db.WithContext(ctx).Model(&models.SitemapNode{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("priority desc").Offset(offset).Limit(pageSize).Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

func (r *sitemapNodeRepository) ListByEntityType(ctx context.Context, entityType string) ([]*models.SitemapNode, error) {
	var nodes []*models.SitemapNode
	err := r.db.WithContext(ctx).Where("entity_type = ?", entityType).Find(&nodes).Error
	return nodes, err
}

func (r *sitemapNodeRepository) Update(ctx context.Context, node *models.SitemapNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *sitemapNodeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.SitemapNode{}, id).Error
}

func (r *sitemapNodeRepository) DeleteByEntityType(ctx context.Context, entityType string) error {
	return r.db.WithContext(ctx).Where("entity_type = ?", entityType).Delete(&models.SitemapNode{}).Error
}

func (r *sitemapNodeRepository) GetAll(ctx context.Context) ([]*models.SitemapNode, error) {
	var nodes []*models.SitemapNode
	err := r.db.WithContext(ctx).Order("priority desc").Find(&nodes).Error
	return nodes, err
}

func (r *sitemapNodeRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SitemapNode{}).Count(&count).Error
	return count, err
}

// 常见错误
var (
	ErrUrlRecordNotFound   = errors.New("url record not found")
	ErrUrlRedirectNotFound = errors.New("url redirect not found")
	ErrMetaInfoNotFound    = errors.New("meta info not found")
	ErrSitemapNodeNotFound = errors.New("sitemap node not found")
	ErrSlugExists          = errors.New("slug already exists")
	ErrOldSlugExists       = errors.New("old slug already has a redirect")
)
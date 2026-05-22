// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/localization/internal/biz"

	"gorm.io/gorm"
)

// LanguagePO 语言持久化对象。
type LanguagePO struct {
	ID                uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name              string    `gorm:"size:256;column:name" db:"name" json:"name"`
	LanguageCulture   string    `gorm:"size:64;column:language_culture" db:"language_culture" json:"language_culture"`
	UniqueSeoCode     string    `gorm:"size:8;column:unique_seo_code" db:"unique_seo_code" json:"unique_seo_code"`
	FlagImageFileName string    `gorm:"size:256;column:flag_image_file_name" db:"flag_image_file_name" json:"flag_image_file_name"`
	Rtl               bool      `gorm:"column:rtl" db:"rtl" json:"rtl"`
	IsActive          bool      `gorm:"column:is_active" db:"is_active" json:"is_active"`
	DisplayOrder      int       `gorm:"column:display_order" db:"display_order" json:"display_order"`
	CreatedAt         time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (LanguagePO) TableName() string {
	return "languages"
}

// ToEntity 转换为领域实体。
func (po *LanguagePO) ToEntity() *biz.Language {
	return &biz.Language{
		ID:                po.ID,
		Name:              po.Name,
		LanguageCulture:   po.LanguageCulture,
		UniqueSeoCode:     po.UniqueSeoCode,
		FlagImageFileName: po.FlagImageFileName,
		Rtl:               po.Rtl,
		IsActive:          po.IsActive,
		DisplayOrder:      po.DisplayOrder,
		CreatedAt:         po.CreatedAt,
		UpdatedAt:         po.UpdatedAt,
	}
}

// LocaleResourcePO 本地化资源持久化对象。
type LocaleResourcePO struct {
	ID            uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	LanguageID    uint      `gorm:"index;column:language_id" db:"language_id" json:"language_id"`
	ResourceName  string    `gorm:"size:512;column:resource_name" db:"resource_name" json:"resource_name"`
	ResourceValue string    `gorm:"type:text;column:resource_value" db:"resource_value" json:"resource_value"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (LocaleResourcePO) TableName() string {
	return "locale_resources"
}

// ToEntity 转换为领域实体。
func (po *LocaleResourcePO) ToEntity() *biz.LocaleResource {
	return &biz.LocaleResource{
		ID:            po.ID,
		LanguageID:    po.LanguageID,
		ResourceName:  po.ResourceName,
		ResourceValue: po.ResourceValue,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}
}

// languageRepo 语言仓储实现。
type languageRepo struct {
	db *gorm.DB
}

// NewLanguageRepo 创建语言仓储。
func NewLanguageRepo(db *gorm.DB) biz.LanguageRepository {
	return &languageRepo{db: db}
}

// Create 创建语言。
func (r *languageRepo) Create(ctx context.Context, lang *biz.Language) error {
	po := &LanguagePO{
		Name:              lang.Name,
		LanguageCulture:   lang.LanguageCulture,
		UniqueSeoCode:     lang.UniqueSeoCode,
		FlagImageFileName: lang.FlagImageFileName,
		Rtl:               lang.Rtl,
		IsActive:          lang.IsActive,
		DisplayOrder:      lang.DisplayOrder,
		CreatedAt:         lang.CreatedAt,
		UpdatedAt:         lang.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取语言。
func (r *languageRepo) GetByID(ctx context.Context, id uint) (*biz.Language, error) {
	var po LanguagePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取语言列表。
func (r *languageRepo) List(ctx context.Context, page, size int) ([]*biz.Language, int64, error) {
	var pos []LanguagePO
	var total int64

	r.db.WithContext(ctx).Model(&LanguagePO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Order("display_order ASC, id ASC").Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	langs := make([]*biz.Language, len(pos))
	for i, po := range pos {
		langs[i] = po.ToEntity()
	}

	return langs, total, nil
}

// Update 更新语言。
func (r *languageRepo) Update(ctx context.Context, lang *biz.Language) error {
	return r.db.WithContext(ctx).Model(&LanguagePO{}).Where("id = ?", lang.ID).Updates(map[string]interface{}{
		"name":                lang.Name,
		"language_culture":    lang.LanguageCulture,
		"unique_seo_code":    lang.UniqueSeoCode,
		"flag_image_file_name": lang.FlagImageFileName,
		"rtl":                lang.Rtl,
		"is_active":          lang.IsActive,
		"display_order":      lang.DisplayOrder,
		"updated_at":         lang.UpdatedAt,
	}).Error
}

// Delete 删除语言。
func (r *languageRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&LanguagePO{}, id).Error
}

// localeResourceRepo 本地化资源仓储实现。
type localeResourceRepo struct {
	db *gorm.DB
}

// NewLocaleResourceRepo 创建本地化资源仓储。
func NewLocaleResourceRepo(db *gorm.DB) biz.LocaleResourceRepository {
	return &localeResourceRepo{db: db}
}

// Create 创建本地化资源。
func (r *localeResourceRepo) Create(ctx context.Context, res *biz.LocaleResource) error {
	po := &LocaleResourcePO{
		LanguageID:    res.LanguageID,
		ResourceName:  res.ResourceName,
		ResourceValue: res.ResourceValue,
		CreatedAt:     res.CreatedAt,
		UpdatedAt:     res.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取本地化资源。
func (r *localeResourceRepo) GetByID(ctx context.Context, id uint) (*biz.LocaleResource, error) {
	var po LocaleResourcePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// ListByLanguageID 根据语言ID获取资源列表。
func (r *localeResourceRepo) ListByLanguageID(ctx context.Context, languageID uint, page, size int) ([]*biz.LocaleResource, int64, error) {
	var pos []LocaleResourcePO
	var total int64

	query := r.db.WithContext(ctx).Model(&LocaleResourcePO{}).Where("language_id = ?", languageID)
	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	resources := make([]*biz.LocaleResource, len(pos))
	for i, po := range pos {
		resources[i] = po.ToEntity()
	}

	return resources, total, nil
}

// Update 更新本地化资源。
func (r *localeResourceRepo) Update(ctx context.Context, res *biz.LocaleResource) error {
	return r.db.WithContext(ctx).Model(&LocaleResourcePO{}).Where("id = ?", res.ID).Updates(map[string]interface{}{
		"resource_name":  res.ResourceName,
		"resource_value": res.ResourceValue,
		"updated_at":     res.UpdatedAt,
	}).Error
}

// Delete 删除本地化资源。
func (r *localeResourceRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&LocaleResourcePO{}, id).Error
}

// BatchCreate 批量创建本地化资源。
func (r *localeResourceRepo) BatchCreate(ctx context.Context, resources []*biz.LocaleResource) error {
	pos := make([]LocaleResourcePO, len(resources))
	for i, res := range resources {
		pos[i] = LocaleResourcePO{
			LanguageID:    res.LanguageID,
			ResourceName:  res.ResourceName,
			ResourceValue: res.ResourceValue,
			CreatedAt:     res.CreatedAt,
			UpdatedAt:     res.UpdatedAt,
		}
	}
	return r.db.WithContext(ctx).Create(&pos).Error
}

// ListAllByLanguageID 获取某语言下的所有资源（导出用）。
func (r *localeResourceRepo) ListAllByLanguageID(ctx context.Context, languageID uint) ([]*biz.LocaleResource, error) {
	var pos []LocaleResourcePO
	if err := r.db.WithContext(ctx).Where("language_id = ?", languageID).Find(&pos).Error; err != nil {
		return nil, err
	}

	resources := make([]*biz.LocaleResource, len(pos))
	for i, po := range pos {
		resources[i] = po.ToEntity()
	}

	return resources, nil
}
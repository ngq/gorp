// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Language 语言领域实体。
type Language struct {
	ID                uint      // 语言ID
	Name              string    // 语言名称
	LanguageCulture   string    // 语言文化代码，如 zh-CN
	UniqueSeoCode     string    // SEO唯一代码，如 zh
	FlagImageFileName string    // 国旗图片文件名
	Rtl               bool      // 是否从右到左书写
	IsActive          bool      // 是否启用
	DisplayOrder      int       // 显示排序
	CreatedAt         time.Time // 创建时间
	UpdatedAt         time.Time // 更新时间
}

// LocaleResource 本地化资源领域实体。
type LocaleResource struct {
	ID            uint      // 资源ID
	LanguageID    uint      // 所属语言ID
	ResourceName  string    // 资源名称（键）
	ResourceValue string    // 资源值
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// LanguageRepository 语言仓储接口。
type LanguageRepository interface {
	// Create 创建语言
	Create(ctx context.Context, lang *Language) error
	// GetByID 根据ID获取语言
	GetByID(ctx context.Context, id uint) (*Language, error)
	// List 获取语言列表
	List(ctx context.Context, page, size int) ([]*Language, int64, error)
	// Update 更新语言
	Update(ctx context.Context, lang *Language) error
	// Delete 删除语言
	Delete(ctx context.Context, id uint) error
}

// LocaleResourceRepository 本地化资源仓储接口。
type LocaleResourceRepository interface {
	// Create 创建本地化资源
	Create(ctx context.Context, res *LocaleResource) error
	// GetByID 根据ID获取本地化资源
	GetByID(ctx context.Context, id uint) (*LocaleResource, error)
	// ListByLanguageID 根据语言ID获取资源列表
	ListByLanguageID(ctx context.Context, languageID uint, page, size int) ([]*LocaleResource, int64, error)
	// Update 更新本地化资源
	Update(ctx context.Context, res *LocaleResource) error
	// Delete 删除本地化资源
	Delete(ctx context.Context, id uint) error
	// BatchCreate 批量创建本地化资源（导入用）
	BatchCreate(ctx context.Context, resources []*LocaleResource) error
	// ListAllByLanguageID 获取某语言下的所有资源（导出用）
	ListAllByLanguageID(ctx context.Context, languageID uint) ([]*LocaleResource, error)
}

// LocalizationUseCase 本地化用例。
type LocalizationUseCase struct {
	langRepo LanguageRepository
	resRepo  LocaleResourceRepository
}

// NewLocalizationUseCase 创建本地化用例。
func NewLocalizationUseCase(langRepo LanguageRepository, resRepo LocaleResourceRepository) *LocalizationUseCase {
	return &LocalizationUseCase{langRepo: langRepo, resRepo: resRepo}
}

// CreateLanguage 创建语言。
func (uc *LocalizationUseCase) CreateLanguage(ctx context.Context, name, languageCulture, uniqueSeoCode, flagImageFileName string, rtl, isActive bool, displayOrder int) (*Language, error) {
	lang := &Language{
		Name:              name,
		LanguageCulture:   languageCulture,
		UniqueSeoCode:     uniqueSeoCode,
		FlagImageFileName: flagImageFileName,
		Rtl:               rtl,
		IsActive:          isActive,
		DisplayOrder:      displayOrder,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := uc.langRepo.Create(ctx, lang); err != nil {
		return nil, err
	}
	return lang, nil
}

// GetLanguageByID 根据ID获取语言。
func (uc *LocalizationUseCase) GetLanguageByID(ctx context.Context, id uint) (*Language, error) {
	return uc.langRepo.GetByID(ctx, id)
}

// ListLanguages 获取语言列表。
func (uc *LocalizationUseCase) ListLanguages(ctx context.Context, page, size int) ([]*Language, int64, error) {
	return uc.langRepo.List(ctx, page, size)
}

// UpdateLanguage 更新语言。
func (uc *LocalizationUseCase) UpdateLanguage(ctx context.Context, id uint, name, languageCulture, uniqueSeoCode, flagImageFileName string, rtl, isActive bool, displayOrder int) (*Language, error) {
	lang, err := uc.langRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	lang.Name = name
	lang.LanguageCulture = languageCulture
	lang.UniqueSeoCode = uniqueSeoCode
	lang.FlagImageFileName = flagImageFileName
	lang.Rtl = rtl
	lang.IsActive = isActive
	lang.DisplayOrder = displayOrder
	lang.UpdatedAt = time.Now()
	if err := uc.langRepo.Update(ctx, lang); err != nil {
		return nil, err
	}
	return lang, nil
}

// DeleteLanguage 删除语言。
func (uc *LocalizationUseCase) DeleteLanguage(ctx context.Context, id uint) error {
	return uc.langRepo.Delete(ctx, id)
}

// AddResource 添加本地化资源。
func (uc *LocalizationUseCase) AddResource(ctx context.Context, languageID uint, resourceName, resourceValue string) (*LocaleResource, error) {
	res := &LocaleResource{
		LanguageID:    languageID,
		ResourceName:  resourceName,
		ResourceValue: resourceValue,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := uc.resRepo.Create(ctx, res); err != nil {
		return nil, err
	}
	return res, nil
}

// GetResourceByID 根据ID获取本地化资源。
func (uc *LocalizationUseCase) GetResourceByID(ctx context.Context, id uint) (*LocaleResource, error) {
	return uc.resRepo.GetByID(ctx, id)
}

// ListResources 获取本地化资源列表。
func (uc *LocalizationUseCase) ListResources(ctx context.Context, languageID uint, page, size int) ([]*LocaleResource, int64, error) {
	return uc.resRepo.ListByLanguageID(ctx, languageID, page, size)
}

// UpdateResource 更新本地化资源。
func (uc *LocalizationUseCase) UpdateResource(ctx context.Context, id uint, resourceName, resourceValue string) (*LocaleResource, error) {
	res, err := uc.resRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res.ResourceName = resourceName
	res.ResourceValue = resourceValue
	res.UpdatedAt = time.Now()
	if err := uc.resRepo.Update(ctx, res); err != nil {
		return nil, err
	}
	return res, nil
}

// DeleteResource 删除本地化资源。
func (uc *LocalizationUseCase) DeleteResource(ctx context.Context, id uint) error {
	return uc.resRepo.Delete(ctx, id)
}

// ExportResources 导出语言资源，返回该语言下所有资源。
func (uc *LocalizationUseCase) ExportResources(ctx context.Context, languageID uint) ([]*LocaleResource, error) {
	return uc.resRepo.ListAllByLanguageID(ctx, languageID)
}

// ImportResources 导入语言资源，批量创建。
func (uc *LocalizationUseCase) ImportResources(ctx context.Context, languageID uint, resources []struct {
	ResourceName  string
	ResourceValue string
}) error {
	// 构建批量资源实体列表
	items := make([]*LocaleResource, len(resources))
	now := time.Now()
	for i, r := range resources {
		items[i] = &LocaleResource{
			LanguageID:    languageID,
			ResourceName:  r.ResourceName,
			ResourceValue: r.ResourceValue,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
	}
	return uc.resRepo.BatchCreate(ctx, items)
}
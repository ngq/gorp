package biz

import (
	"context"
	"time"
)

// ==================== 实体定义 ====================

// Language 语言实体，表示系统支持的语言
type Language struct {
	ID        uint64    `json:"id"`
	Code      string    `json:"code"`       // 语言代码，如 zh-CN, en-US
	Name      string    `json:"name"`       // 语言名称，如 中文(中国), English(US)
	IsDefault bool      `json:"is_default"` // 是否为默认语言
	SortOrder int       `json:"sort_order"` // 排序权重
	IsActive  bool      `json:"is_active"`  // 是否启用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LocaleResource 本地化资源实体，存储各语言的键值对翻译
type LocaleResource struct {
	ID         uint64    `json:"id"`
	LanguageID uint64    `json:"language_id"` // 关联语言 ID
	Key        string    `json:"key"`         // 翻译键，如 "common.hello"
	Value      string    `json:"value"`       // 翻译值，如 "你好"
	Module     string    `json:"module"`      // 所属模块，用于分组管理
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// LanguageRepo 语言仓储接口
type LanguageRepo interface {
	// Create 创建语言
	Create(ctx context.Context, lang *Language) error
	// GetByID 根据 ID 获取语言
	GetByID(ctx context.Context, id uint64) (*Language, error)
	// GetByCode 根据语言代码获取语言
	GetByCode(ctx context.Context, code string) (*Language, error)
	// List 获取语言列表
	List(ctx context.Context, offset, limit int) ([]*Language, error)
	// Update 更新语言
	Update(ctx context.Context, lang *Language) error
	// Delete 删除语言
	Delete(ctx context.Context, id uint64) error
}

// LocaleResourceRepo 本地化资源仓储接口
type LocaleResourceRepo interface {
	// Create 创建本地化资源
	Create(ctx context.Context, resource *LocaleResource) error
	// GetByID 根据 ID 获取本地化资源
	GetByID(ctx context.Context, id uint64) (*LocaleResource, error)
	// ListByLanguageID 根据语言 ID 获取资源列表
	ListByLanguageID(ctx context.Context, languageID uint64, offset, limit int) ([]*LocaleResource, error)
	// GetByKey 根据语言 ID 和键获取资源
	GetByKey(ctx context.Context, languageID uint64, key string) (*LocaleResource, error)
	// Update 更新本地化资源
	Update(ctx context.Context, resource *LocaleResource) error
	// Delete 删除本地化资源
	Delete(ctx context.Context, id uint64) error
}

// ==================== 用例 ====================

// LocalizationUseCase 本地化业务用例
type LocalizationUseCase struct {
	langRepo     LanguageRepo
	resourceRepo LocaleResourceRepo
}

// NewLocalizationUseCase 创建本地化用例
func NewLocalizationUseCase(langRepo LanguageRepo, resourceRepo LocaleResourceRepo) *LocalizationUseCase {
	return &LocalizationUseCase{
		langRepo:     langRepo,
		resourceRepo: resourceRepo,
	}
}

// CreateLanguage 创建语言
func (uc *LocalizationUseCase) CreateLanguage(ctx context.Context, lang *Language) error {
	return uc.langRepo.Create(ctx, lang)
}

// GetLanguage 获取语言详情
func (uc *LocalizationUseCase) GetLanguage(ctx context.Context, id uint64) (*Language, error) {
	return uc.langRepo.GetByID(ctx, id)
}

// GetLanguageByCode 根据语言代码获取语言
func (uc *LocalizationUseCase) GetLanguageByCode(ctx context.Context, code string) (*Language, error) {
	return uc.langRepo.GetByCode(ctx, code)
}

// ListLanguages 获取语言列表
func (uc *LocalizationUseCase) ListLanguages(ctx context.Context, offset, limit int) ([]*Language, error) {
	return uc.langRepo.List(ctx, offset, limit)
}

// UpdateLanguage 更新语言
func (uc *LocalizationUseCase) UpdateLanguage(ctx context.Context, lang *Language) error {
	return uc.langRepo.Update(ctx, lang)
}

// DeleteLanguage 删除语言
func (uc *LocalizationUseCase) DeleteLanguage(ctx context.Context, id uint64) error {
	return uc.langRepo.Delete(ctx, id)
}

// CreateLocaleResource 创建本地化资源
func (uc *LocalizationUseCase) CreateLocaleResource(ctx context.Context, resource *LocaleResource) error {
	return uc.resourceRepo.Create(ctx, resource)
}

// GetLocaleResource 获取本地化资源详情
func (uc *LocalizationUseCase) GetLocaleResource(ctx context.Context, id uint64) (*LocaleResource, error) {
	return uc.resourceRepo.GetByID(ctx, id)
}

// ListLocaleResources 根据语言 ID 获取本地化资源列表
func (uc *LocalizationUseCase) ListLocaleResources(ctx context.Context, languageID uint64, offset, limit int) ([]*LocaleResource, error) {
	return uc.resourceRepo.ListByLanguageID(ctx, languageID, offset, limit)
}

// GetLocaleResourceByKey 根据语言 ID 和键获取本地化资源
func (uc *LocalizationUseCase) GetLocaleResourceByKey(ctx context.Context, languageID uint64, key string) (*LocaleResource, error) {
	return uc.resourceRepo.GetByKey(ctx, languageID, key)
}

// UpdateLocaleResource 更新本地化资源
func (uc *LocalizationUseCase) UpdateLocaleResource(ctx context.Context, resource *LocaleResource) error {
	return uc.resourceRepo.Update(ctx, resource)
}

// DeleteLocaleResource 删除本地化资源
func (uc *LocalizationUseCase) DeleteLocaleResource(ctx context.Context, id uint64) error {
	return uc.resourceRepo.Delete(ctx, id)
}

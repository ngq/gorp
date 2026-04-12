// Package biz 本地化服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"nop-go/services/localization-service/internal/data"
	"nop-go/services/localization-service/internal/models"
)

// LocalizationUseCase 本地化用例
type LocalizationUseCase struct {
	languageRepo data.LanguageRepository
	resourceRepo data.LocaleStringResourceRepository
	currencyRepo data.CurrencyRepository
}

// NewLocalizationUseCase 创建本地化用例
func NewLocalizationUseCase(
	languageRepo data.LanguageRepository,
	resourceRepo data.LocaleStringResourceRepository,
	currencyRepo data.CurrencyRepository,
) *LocalizationUseCase {
	return &LocalizationUseCase{
		languageRepo: languageRepo,
		resourceRepo: resourceRepo,
		currencyRepo: currencyRepo,
	}
}

// ========== 语言管理 ==========

// CreateLanguage 创建语言
func (uc *LocalizationUseCase) CreateLanguage(ctx context.Context, req *models.LanguageCreateRequest) (*models.Language, error) {
	// 检查文化代码是否已存在
	existing, err := uc.languageRepo.GetByCulture(ctx, req.LanguageCulture)
	if err == nil && existing != nil {
		return nil, data.ErrCultureExists
	}

	language := &models.Language{
		Name:              req.Name,
		LanguageCulture:   req.LanguageCulture,
		UniqueSeoCode:     req.UniqueSeoCode,
		Published:         req.Published,
		DisplayOrder:      req.DisplayOrder,
		Rtl:               req.Rtl,
		FlagImageFileName: req.FlagImageFileName,
		DefaultCurrencyID: req.DefaultCurrencyID,
	}

	if err := uc.languageRepo.Create(ctx, language); err != nil {
		return nil, err
	}

	return language, nil
}

// GetLanguage 获取语言
func (uc *LocalizationUseCase) GetLanguage(ctx context.Context, id uint) (*models.Language, error) {
	return uc.languageRepo.GetByID(ctx, id)
}

// GetLanguageByCulture 通过文化代码获取语言
func (uc *LocalizationUseCase) GetLanguageByCulture(ctx context.Context, culture string) (*models.Language, error) {
	return uc.languageRepo.GetByCulture(ctx, culture)
}

// ListLanguages 语言列表
func (uc *LocalizationUseCase) ListLanguages(ctx context.Context, page, pageSize int) ([]*models.Language, int64, error) {
	return uc.languageRepo.List(ctx, page, pageSize)
}

// ListPublishedLanguages 获取所有已发布语言
func (uc *LocalizationUseCase) ListPublishedLanguages(ctx context.Context) ([]*models.Language, error) {
	return uc.languageRepo.ListPublished(ctx)
}

// UpdateLanguage 更新语言
func (uc *LocalizationUseCase) UpdateLanguage(ctx context.Context, id uint, req *models.LanguageUpdateRequest) (*models.Language, error) {
	language, err := uc.languageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrLanguageNotFound
	}

	// 更新字段
	if req.Name != "" {
		language.Name = req.Name
	}
	if req.LanguageCulture != "" {
		// 检查新文化代码是否已被其他语言使用
		existing, err := uc.languageRepo.GetByCulture(ctx, req.LanguageCulture)
		if err == nil && existing != nil && existing.ID != id {
			return nil, data.ErrCultureExists
		}
		language.LanguageCulture = req.LanguageCulture
	}
	if req.UniqueSeoCode != "" {
		language.UniqueSeoCode = req.UniqueSeoCode
	}
	language.Published = req.Published
	language.DisplayOrder = req.DisplayOrder
	language.Rtl = req.Rtl
	if req.FlagImageFileName != "" {
		language.FlagImageFileName = req.FlagImageFileName
	}
	language.DefaultCurrencyID = req.DefaultCurrencyID

	if err := uc.languageRepo.Update(ctx, language); err != nil {
		return nil, err
	}

	return language, nil
}

// DeleteLanguage 删除语言
func (uc *LocalizationUseCase) DeleteLanguage(ctx context.Context, id uint) error {
	// 删除语言前删除相关资源
	if err := uc.resourceRepo.DeleteByLanguageID(ctx, id); err != nil {
		return err
	}
	return uc.languageRepo.Delete(ctx, id)
}

// ========== 本地化资源管理 ==========

// CreateResource 创建本地化资源
func (uc *LocalizationUseCase) CreateResource(ctx context.Context, req *models.ResourceCreateRequest) (*models.LocaleStringResource, error) {
	// 验证语言存在
	_, err := uc.languageRepo.GetByID(ctx, req.LanguageID)
	if err != nil {
		return nil, data.ErrLanguageNotFound
	}

	// 检查资源名称是否已存在
	existing, err := uc.resourceRepo.GetByLanguageIDAndName(ctx, req.LanguageID, req.ResourceName)
	if err == nil && existing != nil {
		return nil, errors.New("resource name already exists for this language")
	}

	resource := &models.LocaleStringResource{
		LanguageID:    req.LanguageID,
		ResourceName:  req.ResourceName,
		ResourceValue: req.ResourceValue,
		IsTouched:     true,
	}

	if err := uc.resourceRepo.Create(ctx, resource); err != nil {
		return nil, err
	}

	return resource, nil
}

// GetResource 获取本地化资源
func (uc *LocalizationUseCase) GetResource(ctx context.Context, id uint) (*models.LocaleStringResource, error) {
	return uc.resourceRepo.GetByID(ctx, id)
}

// GetResourcesByLanguage 获取语言的所有资源
func (uc *LocalizationUseCase) GetResourcesByLanguage(ctx context.Context, languageID uint) ([]*models.LocaleStringResource, error) {
	return uc.resourceRepo.GetByLanguageID(ctx, languageID)
}

// GetTranslation 获取翻译
// 根据语言ID和资源名称获取翻译值
func (uc *LocalizationUseCase) GetTranslation(ctx context.Context, languageID uint, resourceName string) (string, error) {
	resource, err := uc.resourceRepo.GetByLanguageIDAndName(ctx, languageID, resourceName)
	if err != nil {
		return "", data.ErrResourceNotFound
	}
	return resource.ResourceValue, nil
}

// GetTranslations 获取多个翻译
// 根据语言ID和资源名称列表批量获取翻译
func (uc *LocalizationUseCase) GetTranslations(ctx context.Context, languageID uint, names []string) (map[string]string, error) {
	resources, err := uc.resourceRepo.GetByNames(ctx, languageID, names)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, r := range resources {
		result[r.ResourceName] = r.ResourceValue
	}
	return result, nil
}

// UpdateResource 更新本地化资源
func (uc *LocalizationUseCase) UpdateResource(ctx context.Context, id uint, req *models.ResourceUpdateRequest) (*models.LocaleStringResource, error) {
	resource, err := uc.resourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, data.ErrResourceNotFound
	}

	resource.ResourceValue = req.ResourceValue
	resource.IsTouched = true

	if err := uc.resourceRepo.Update(ctx, resource); err != nil {
		return nil, err
	}

	return resource, nil
}

// BatchUpdateResources 批量更新资源
func (uc *LocalizationUseCase) BatchUpdateResources(ctx context.Context, req *models.ResourceBatchUpdateRequest) error {
	// 验证语言存在
	_, err := uc.languageRepo.GetByID(ctx, req.LanguageID)
	if err != nil {
		return data.ErrLanguageNotFound
	}

	for _, item := range req.Resources {
		// 查找现有资源
		existing, err := uc.resourceRepo.GetByLanguageIDAndName(ctx, req.LanguageID, item.ResourceName)
		if err != nil {
			// 创建新资源
			newResource := &models.LocaleStringResource{
				LanguageID:    req.LanguageID,
				ResourceName:  item.ResourceName,
				ResourceValue: item.ResourceValue,
				IsTouched:     true,
			}
			if err := uc.resourceRepo.Create(ctx, newResource); err != nil {
				return err
			}
		} else {
			// 更新现有资源
			existing.ResourceValue = item.ResourceValue
			existing.IsTouched = true
			if err := uc.resourceRepo.Update(ctx, existing); err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteResource 删除本地化资源
func (uc *LocalizationUseCase) DeleteResource(ctx context.Context, id uint) error {
	return uc.resourceRepo.Delete(ctx, id)
}

// SearchResources 搜索资源
func (uc *LocalizationUseCase) SearchResources(ctx context.Context, languageID uint, keyword string, page, pageSize int) ([]*models.LocaleStringResource, int64, error) {
	return uc.resourceRepo.Search(ctx, languageID, keyword, page, pageSize)
}

// ========== 货币管理 ==========

// CreateCurrency 创建货币
func (uc *LocalizationUseCase) CreateCurrency(ctx context.Context, currency *models.Currency) (*models.Currency, error) {
	// 检查货币代码是否已存在
	existing, err := uc.currencyRepo.GetByCode(ctx, currency.CurrencyCode)
	if err == nil && existing != nil {
		return nil, data.ErrCurrencyCodeExists
	}

	if err := uc.currencyRepo.Create(ctx, currency); err != nil {
		return nil, err
	}

	return currency, nil
}

// GetCurrency 获取货币
func (uc *LocalizationUseCase) GetCurrency(ctx context.Context, id uint) (*models.Currency, error) {
	return uc.currencyRepo.GetByID(ctx, id)
}

// GetCurrencyByCode 通过代码获取货币
func (uc *LocalizationUseCase) GetCurrencyByCode(ctx context.Context, code string) (*models.Currency, error) {
	return uc.currencyRepo.GetByCode(ctx, code)
}

// ListCurrencies 货币列表
func (uc *LocalizationUseCase) ListCurrencies(ctx context.Context, page, pageSize int) ([]*models.Currency, int64, error) {
	return uc.currencyRepo.List(ctx, page, pageSize)
}

// ListPublishedCurrencies 获取所有已发布货币
func (uc *LocalizationUseCase) ListPublishedCurrencies(ctx context.Context) ([]*models.Currency, error) {
	return uc.currencyRepo.ListPublished(ctx)
}

// UpdateCurrency 更新货币
func (uc *LocalizationUseCase) UpdateCurrency(ctx context.Context, currency *models.Currency) error {
	return uc.currencyRepo.Update(ctx, currency)
}

// DeleteCurrency 删除货币
func (uc *LocalizationUseCase) DeleteCurrency(ctx context.Context, id uint) error {
	return uc.currencyRepo.Delete(ctx, id)
}

// ========== 翻译辅助方法 ==========

// Translate 翻译文本
// 如果找不到翻译，返回原始文本
func (uc *LocalizationUseCase) Translate(ctx context.Context, languageID uint, text string) string {
	// 尝试获取翻译
	resource, err := uc.resourceRepo.GetByLanguageIDAndName(ctx, languageID, text)
	if err != nil {
		return text // 未找到翻译，返回原文
	}
	return resource.ResourceValue
}

// TranslateByCulture 通过文化代码翻译
func (uc *LocalizationUseCase) TranslateByCulture(ctx context.Context, culture string, resourceName string) (string, error) {
	// 获取语言
	language, err := uc.languageRepo.GetByCulture(ctx, culture)
	if err != nil {
		return "", data.ErrLanguageNotFound
	}

	// 获取翻译
	return uc.GetTranslation(ctx, language.ID, resourceName)
}

// GetAllTranslations 获取语言的所有翻译作为Map
func (uc *LocalizationUseCase) GetAllTranslations(ctx context.Context, languageID uint) (map[string]string, error) {
	resources, err := uc.resourceRepo.GetByLanguageID(ctx, languageID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, r := range resources {
		result[r.ResourceName] = r.ResourceValue
	}
	return result, nil
}

// GetResourceGroups 获取资源分组列表
// 通过解析资源名称提取分组信息
func (uc *LocalizationUseCase) GetResourceGroups(ctx context.Context, languageID uint) ([]string, error) {
	resources, err := uc.resourceRepo.GetByLanguageID(ctx, languageID)
	if err != nil {
		return nil, err
	}

	groups := make(map[string]bool)
	for _, r := range resources {
		// 资源名称格式: Group.SubGroup.Key
		parts := strings.SplitN(r.ResourceName, ".", 2)
		if len(parts) > 0 {
			groups[parts[0]] = true
		}
	}

	result := make([]string, 0, len(groups))
	for g := range groups {
		result = append(result, g)
	}
	return result, nil
}

// GetResourcesByGroup 获取指定分组的资源
func (uc *LocalizationUseCase) GetResourcesByGroup(ctx context.Context, languageID uint, group string) ([]*models.LocaleStringResource, error) {
	resources, err := uc.resourceRepo.GetByLanguageID(ctx, languageID)
	if err != nil {
		return nil, err
	}

	// 过滤出指定分组的资源
	var result []*models.LocaleStringResource
	prefix := group + "."
	for _, r := range resources {
		if strings.HasPrefix(r.ResourceName, prefix) {
			result = append(result, r)
		}
	}
	return result, nil
}

// ========== 默认语言处理 ==========

// GetDefaultLanguage 获取默认语言
func (uc *LocalizationUseCase) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	return uc.languageRepo.GetDefaultLanguage(ctx)
}

// EnsureDefaultLanguage 确保存在默认语言
// 如果不存在任何语言，创建默认英语
func (uc *LocalizationUseCase) EnsureDefaultLanguage(ctx context.Context) (*models.Language, error) {
	defaultLang, err := uc.languageRepo.GetDefaultLanguage(ctx)
	if err == nil {
		return defaultLang, nil
	}

	// 创建默认英语语言
	defaultLang = &models.Language{
		Name:            "English",
		LanguageCulture: "en-US",
		UniqueSeoCode:   "en",
		Published:       true,
		DisplayOrder:    1,
		Rtl:             false,
	}

	if err := uc.languageRepo.Create(ctx, defaultLang); err != nil {
		return nil, fmt.Errorf("failed to create default language: %w", err)
	}

	return defaultLang, nil
}
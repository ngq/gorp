// Package data 本地化服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/localization-service/internal/models"

	"gorm.io/gorm"
)

// LanguageRepository 语言仓储接口
type LanguageRepository interface {
	Create(ctx context.Context, language *models.Language) error
	GetByID(ctx context.Context, id uint) (*models.Language, error)
	GetByCulture(ctx context.Context, culture string) (*models.Language, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Language, int64, error)
	ListPublished(ctx context.Context) ([]*models.Language, error)
	Update(ctx context.Context, language *models.Language) error
	Delete(ctx context.Context, id uint) error
	GetDefaultLanguage(ctx context.Context) (*models.Language, error)
}

type languageRepository struct {
	db *gorm.DB
}

func NewLanguageRepository(db *gorm.DB) LanguageRepository {
	return &languageRepository{db: db}
}

func (r *languageRepository) Create(ctx context.Context, language *models.Language) error {
	return r.db.WithContext(ctx).Create(language).Error
}

func (r *languageRepository) GetByID(ctx context.Context, id uint) (*models.Language, error) {
	var language models.Language
	err := r.db.WithContext(ctx).First(&language, id).Error
	if err != nil {
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) GetByCulture(ctx context.Context, culture string) (*models.Language, error) {
	var language models.Language
	err := r.db.WithContext(ctx).Where("language_culture = ?", culture).First(&language).Error
	if err != nil {
		return nil, err
	}
	return &language, nil
}

func (r *languageRepository) List(ctx context.Context, page, pageSize int) ([]*models.Language, int64, error) {
	var languages []*models.Language
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Language{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("display_order asc").Offset(offset).Limit(pageSize).Find(&languages).Error; err != nil {
		return nil, 0, err
	}

	return languages, total, nil
}

func (r *languageRepository) ListPublished(ctx context.Context) ([]*models.Language, error) {
	var languages []*models.Language
	err := r.db.WithContext(ctx).Where("published = ?", true).Order("display_order asc").Find(&languages).Error
	return languages, err
}

func (r *languageRepository) Update(ctx context.Context, language *models.Language) error {
	return r.db.WithContext(ctx).Save(language).Error
}

func (r *languageRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Language{}, id).Error
}

func (r *languageRepository) GetDefaultLanguage(ctx context.Context) (*models.Language, error) {
	var language models.Language
	// 查找第一个发布的语言作为默认语言
	err := r.db.WithContext(ctx).Where("published = ?", true).Order("display_order asc").First(&language).Error
	if err != nil {
		return nil, err
	}
	return &language, nil
}

// LocaleStringResourceRepository 本地化资源仓储接口
type LocaleStringResourceRepository interface {
	Create(ctx context.Context, resource *models.LocaleStringResource) error
	GetByID(ctx context.Context, id uint) (*models.LocaleStringResource, error)
	GetByLanguageID(ctx context.Context, languageID uint) ([]*models.LocaleStringResource, error)
	GetByLanguageIDAndName(ctx context.Context, languageID uint, resourceName string) (*models.LocaleStringResource, error)
	GetByNames(ctx context.Context, languageID uint, names []string) ([]*models.LocaleStringResource, error)
	Update(ctx context.Context, resource *models.LocaleStringResource) error
	Delete(ctx context.Context, id uint) error
	DeleteByLanguageID(ctx context.Context, languageID uint) error
	BatchUpdate(ctx context.Context, resources []*models.LocaleStringResource) error
	Search(ctx context.Context, languageID uint, keyword string, page, pageSize int) ([]*models.LocaleStringResource, int64, error)
}

type localeStringResourceRepository struct {
	db *gorm.DB
}

func NewLocaleStringResourceRepository(db *gorm.DB) LocaleStringResourceRepository {
	return &localeStringResourceRepository{db: db}
}

func (r *localeStringResourceRepository) Create(ctx context.Context, resource *models.LocaleStringResource) error {
	return r.db.WithContext(ctx).Create(resource).Error
}

func (r *localeStringResourceRepository) GetByID(ctx context.Context, id uint) (*models.LocaleStringResource, error) {
	var resource models.LocaleStringResource
	err := r.db.WithContext(ctx).First(&resource, id).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *localeStringResourceRepository) GetByLanguageID(ctx context.Context, languageID uint) ([]*models.LocaleStringResource, error) {
	var resources []*models.LocaleStringResource
	err := r.db.WithContext(ctx).Where("language_id = ?", languageID).Order("resource_name asc").Find(&resources).Error
	return resources, err
}

func (r *localeStringResourceRepository) GetByLanguageIDAndName(ctx context.Context, languageID uint, resourceName string) (*models.LocaleStringResource, error) {
	var resource models.LocaleStringResource
	err := r.db.WithContext(ctx).Where("language_id = ? AND resource_name = ?", languageID, resourceName).First(&resource).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *localeStringResourceRepository) GetByNames(ctx context.Context, languageID uint, names []string) ([]*models.LocaleStringResource, error) {
	var resources []*models.LocaleStringResource
	err := r.db.WithContext(ctx).Where("language_id = ? AND resource_name IN ?", languageID, names).Find(&resources).Error
	return resources, err
}

func (r *localeStringResourceRepository) Update(ctx context.Context, resource *models.LocaleStringResource) error {
	return r.db.WithContext(ctx).Save(resource).Error
}

func (r *localeStringResourceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.LocaleStringResource{}, id).Error
}

func (r *localeStringResourceRepository) DeleteByLanguageID(ctx context.Context, languageID uint) error {
	return r.db.WithContext(ctx).Where("language_id = ?", languageID).Delete(&models.LocaleStringResource{}).Error
}

func (r *localeStringResourceRepository) BatchUpdate(ctx context.Context, resources []*models.LocaleStringResource) error {
	return r.db.WithContext(ctx).Save(resources).Error
}

func (r *localeStringResourceRepository) Search(ctx context.Context, languageID uint, keyword string, page, pageSize int) ([]*models.LocaleStringResource, int64, error) {
	var resources []*models.LocaleStringResource
	var total int64

	db := r.db.WithContext(ctx).Model(&models.LocaleStringResource{}).Where("language_id = ?", languageID)
	if keyword != "" {
		db = db.Where("resource_name LIKE ? OR resource_value LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("resource_name asc").Offset(offset).Limit(pageSize).Find(&resources).Error; err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

// CurrencyRepository 货币仓储接口
type CurrencyRepository interface {
	Create(ctx context.Context, currency *models.Currency) error
	GetByID(ctx context.Context, id uint) (*models.Currency, error)
	GetByCode(ctx context.Context, code string) (*models.Currency, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Currency, int64, error)
	ListPublished(ctx context.Context) ([]*models.Currency, error)
	Update(ctx context.Context, currency *models.Currency) error
	Delete(ctx context.Context, id uint) error
}

type currencyRepository struct {
	db *gorm.DB
}

func NewCurrencyRepository(db *gorm.DB) CurrencyRepository {
	return &currencyRepository{db: db}
}

func (r *currencyRepository) Create(ctx context.Context, currency *models.Currency) error {
	return r.db.WithContext(ctx).Create(currency).Error
}

func (r *currencyRepository) GetByID(ctx context.Context, id uint) (*models.Currency, error) {
	var currency models.Currency
	err := r.db.WithContext(ctx).First(&currency, id).Error
	if err != nil {
		return nil, err
	}
	return &currency, nil
}

func (r *currencyRepository) GetByCode(ctx context.Context, code string) (*models.Currency, error) {
	var currency models.Currency
	err := r.db.WithContext(ctx).Where("currency_code = ?", code).First(&currency).Error
	if err != nil {
		return nil, err
	}
	return &currency, nil
}

func (r *currencyRepository) List(ctx context.Context, page, pageSize int) ([]*models.Currency, int64, error) {
	var currencies []*models.Currency
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Currency{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("display_order asc").Offset(offset).Limit(pageSize).Find(&currencies).Error; err != nil {
		return nil, 0, err
	}

	return currencies, total, nil
}

func (r *currencyRepository) ListPublished(ctx context.Context) ([]*models.Currency, error) {
	var currencies []*models.Currency
	err := r.db.WithContext(ctx).Where("published = ?", true).Order("display_order asc").Find(&currencies).Error
	return currencies, err
}

func (r *currencyRepository) Update(ctx context.Context, currency *models.Currency) error {
	return r.db.WithContext(ctx).Save(currency).Error
}

func (r *currencyRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Currency{}, id).Error
}

// 常见错误
var (
	ErrLanguageNotFound  = errors.New("language not found")
	ErrResourceNotFound  = errors.New("locale resource not found")
	ErrCurrencyNotFound  = errors.New("currency not found")
	ErrCultureExists     = errors.New("language culture already exists")
	ErrCurrencyCodeExists = errors.New("currency code already exists")
)
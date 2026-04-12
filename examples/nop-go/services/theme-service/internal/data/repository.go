// Package data 主题服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/theme-service/internal/models"

	"gorm.io/gorm"
)

// ThemeRepository 主题仓储接口
type ThemeRepository interface {
	Create(ctx context.Context, theme *models.Theme) error
	GetByID(ctx context.Context, id uint) (*models.Theme, error)
	GetByName(ctx context.Context, name string) (*models.Theme, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Theme, int64, error)
	ListActive(ctx context.Context) ([]*models.Theme, error)
	GetDefault(ctx context.Context) (*models.Theme, error)
	Update(ctx context.Context, theme *models.Theme) error
	Delete(ctx context.Context, id uint) error
}

type themeRepo struct{ db *gorm.DB }

func NewThemeRepository(db *gorm.DB) ThemeRepository {
	return &themeRepo{db: db}
}

func (r *themeRepo) Create(ctx context.Context, t *models.Theme) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *themeRepo) GetByID(ctx context.Context, id uint) (*models.Theme, error) {
	var t models.Theme
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *themeRepo) GetByName(ctx context.Context, name string) (*models.Theme, error) {
	var t models.Theme
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *themeRepo) List(ctx context.Context, page, pageSize int) ([]*models.Theme, int64, error) {
	var list []*models.Theme
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Theme{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *themeRepo) ListActive(ctx context.Context) ([]*models.Theme, error) {
	var list []*models.Theme
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("is_default desc, created_at desc").Find(&list).Error
	return list, err
}

func (r *themeRepo) GetDefault(ctx context.Context) (*models.Theme, error) {
	var t models.Theme
	err := r.db.WithContext(ctx).Where("is_default = ? AND is_active = ?", true, true).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *themeRepo) Update(ctx context.Context, t *models.Theme) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *themeRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Theme{}, id).Error
}

// ThemeVariableRepository 主题变量仓储接口
type ThemeVariableRepository interface {
	Create(ctx context.Context, v *models.ThemeVariable) error
	CreateBatch(ctx context.Context, vars []*models.ThemeVariable) error
	GetByID(ctx context.Context, id uint) (*models.ThemeVariable, error)
	GetByThemeID(ctx context.Context, themeID uint) ([]*models.ThemeVariable, error)
	GetByThemeIDAndName(ctx context.Context, themeID uint, name string) (*models.ThemeVariable, error)
	Update(ctx context.Context, v *models.ThemeVariable) error
	Delete(ctx context.Context, id uint) error
	DeleteByThemeID(ctx context.Context, themeID uint) error
}

type themeVariableRepo struct{ db *gorm.DB }

func NewThemeVariableRepository(db *gorm.DB) ThemeVariableRepository {
	return &themeVariableRepo{db: db}
}

func (r *themeVariableRepo) Create(ctx context.Context, v *models.ThemeVariable) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *themeVariableRepo) CreateBatch(ctx context.Context, vars []*models.ThemeVariable) error {
	return r.db.WithContext(ctx).Create(vars).Error
}

func (r *themeVariableRepo) GetByID(ctx context.Context, id uint) (*models.ThemeVariable, error) {
	var v models.ThemeVariable
	err := r.db.WithContext(ctx).First(&v, id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *themeVariableRepo) GetByThemeID(ctx context.Context, themeID uint) ([]*models.ThemeVariable, error) {
	var list []*models.ThemeVariable
	err := r.db.WithContext(ctx).Where("theme_id = ?", themeID).Order("category, display_order").Find(&list).Error
	return list, err
}

func (r *themeVariableRepo) GetByThemeIDAndName(ctx context.Context, themeID uint, name string) (*models.ThemeVariable, error) {
	var v models.ThemeVariable
	err := r.db.WithContext(ctx).Where("theme_id = ? AND name = ?", themeID, name).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *themeVariableRepo) Update(ctx context.Context, v *models.ThemeVariable) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *themeVariableRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ThemeVariable{}, id).Error
}

func (r *themeVariableRepo) DeleteByThemeID(ctx context.Context, themeID uint) error {
	return r.db.WithContext(ctx).Where("theme_id = ?", themeID).Delete(&models.ThemeVariable{}).Error
}

// ThemeConfigurationRepository 主题配置仓储接口
type ThemeConfigurationRepository interface {
	Create(ctx context.Context, c *models.ThemeConfiguration) error
	GetByID(ctx context.Context, id uint) (*models.ThemeConfiguration, error)
	GetByThemeAndStore(ctx context.Context, themeID, storeID uint) (*models.ThemeConfiguration, error)
	GetByStore(ctx context.Context, storeID uint) ([]*models.ThemeConfiguration, error)
	Update(ctx context.Context, c *models.ThemeConfiguration) error
	Delete(ctx context.Context, id uint) error
}

type themeConfigurationRepo struct{ db *gorm.DB }

func NewThemeConfigurationRepository(db *gorm.DB) ThemeConfigurationRepository {
	return &themeConfigurationRepo{db: db}
}

func (r *themeConfigurationRepo) Create(ctx context.Context, c *models.ThemeConfiguration) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *themeConfigurationRepo) GetByID(ctx context.Context, id uint) (*models.ThemeConfiguration, error) {
	var c models.ThemeConfiguration
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *themeConfigurationRepo) GetByThemeAndStore(ctx context.Context, themeID, storeID uint) (*models.ThemeConfiguration, error) {
	var c models.ThemeConfiguration
	err := r.db.WithContext(ctx).Where("theme_id = ? AND store_id = ?", themeID, storeID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *themeConfigurationRepo) GetByStore(ctx context.Context, storeID uint) ([]*models.ThemeConfiguration, error) {
	var list []*models.ThemeConfiguration
	err := r.db.WithContext(ctx).Where("store_id = ? AND is_active = ?", storeID, true).Find(&list).Error
	return list, err
}

func (r *themeConfigurationRepo) Update(ctx context.Context, c *models.ThemeConfiguration) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *themeConfigurationRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ThemeConfiguration{}, id).Error
}

// CustomerThemeSettingRepository 客户主题设置仓储接口
type CustomerThemeSettingRepository interface {
	Create(ctx context.Context, s *models.CustomerThemeSetting) error
	GetByCustomerID(ctx context.Context, customerID uint) (*models.CustomerThemeSetting, error)
	Update(ctx context.Context, s *models.CustomerThemeSetting) error
	Delete(ctx context.Context, customerID uint) error
}

type customerThemeSettingRepo struct{ db *gorm.DB }

func NewCustomerThemeSettingRepository(db *gorm.DB) CustomerThemeSettingRepository {
	return &customerThemeSettingRepo{db: db}
}

func (r *customerThemeSettingRepo) Create(ctx context.Context, s *models.CustomerThemeSetting) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *customerThemeSettingRepo) GetByCustomerID(ctx context.Context, customerID uint) (*models.CustomerThemeSetting, error) {
	var s models.CustomerThemeSetting
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *customerThemeSettingRepo) Update(ctx context.Context, s *models.CustomerThemeSetting) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *customerThemeSettingRepo) Delete(ctx context.Context, customerID uint) error {
	return r.db.WithContext(ctx).Where("customer_id = ?", customerID).Delete(&models.CustomerThemeSetting{}).Error
}

// ThemeFileRepository 主题文件仓储接口
type ThemeFileRepository interface {
	Create(ctx context.Context, f *models.ThemeFile) error
	GetByID(ctx context.Context, id uint) (*models.ThemeFile, error)
	GetByThemeID(ctx context.Context, themeID uint) ([]*models.ThemeFile, error)
	GetByThemeAndPath(ctx context.Context, themeID uint, filePath string) (*models.ThemeFile, error)
	Update(ctx context.Context, f *models.ThemeFile) error
	Delete(ctx context.Context, id uint) error
	DeleteByThemeID(ctx context.Context, themeID uint) error
}

type themeFileRepo struct{ db *gorm.DB }

func NewThemeFileRepository(db *gorm.DB) ThemeFileRepository {
	return &themeFileRepo{db: db}
}

func (r *themeFileRepo) Create(ctx context.Context, f *models.ThemeFile) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *themeFileRepo) GetByID(ctx context.Context, id uint) (*models.ThemeFile, error) {
	var f models.ThemeFile
	err := r.db.WithContext(ctx).First(&f, id).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *themeFileRepo) GetByThemeID(ctx context.Context, themeID uint) ([]*models.ThemeFile, error) {
	var list []*models.ThemeFile
	err := r.db.WithContext(ctx).Where("theme_id = ?", themeID).Order("file_type, file_path").Find(&list).Error
	return list, err
}

func (r *themeFileRepo) GetByThemeAndPath(ctx context.Context, themeID uint, filePath string) (*models.ThemeFile, error) {
	var f models.ThemeFile
	err := r.db.WithContext(ctx).Where("theme_id = ? AND file_path = ?", themeID, filePath).First(&f).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *themeFileRepo) Update(ctx context.Context, f *models.ThemeFile) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *themeFileRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ThemeFile{}, id).Error
}

func (r *themeFileRepo) DeleteByThemeID(ctx context.Context, themeID uint) error {
	return r.db.WithContext(ctx).Where("theme_id = ?", themeID).Delete(&models.ThemeFile{}).Error
}

// 错误定义
var (
	ErrThemeNotFound          = errors.New("theme not found")
	ErrThemeVariableNotFound  = errors.New("theme variable not found")
	ErrThemeConfigurationNotFound = errors.New("theme configuration not found")
	ErrCustomerThemeNotFound  = errors.New("customer theme setting not found")
	ErrThemeFileNotFound      = errors.New("theme file not found")
	ErrThemeNameExists        = errors.New("theme name already exists")
	ErrCannotDeleteDefault    = errors.New("cannot delete default theme")
)
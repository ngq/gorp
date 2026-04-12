// Package biz 主题服务业务逻辑层
package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"nop-go/services/theme-service/internal/data"
	"nop-go/services/theme-service/internal/models"

	"gorm.io/gorm"
)

// ImportConfig 导入配置（主题服务配置部分）
type ThemeConfig struct {
	ThemesDir         string // 主题目录
	DefaultTheme      string // 默认主题名称
	AllowUserSelection bool   // 是否允许用户切换主题
}

// ThemeUseCase 主题用例
type ThemeUseCase struct {
	themeRepo   data.ThemeRepository
	variableRepo data.ThemeVariableRepository
	configRepo  data.ThemeConfigurationRepository
	customerRepo data.CustomerThemeSettingRepository
	fileRepo    data.ThemeFileRepository
	config      ThemeConfig
}

// NewThemeUseCase 创建主题用例
func NewThemeUseCase(
	themeRepo data.ThemeRepository,
	variableRepo data.ThemeVariableRepository,
	configRepo data.ThemeConfigurationRepository,
	customerRepo data.CustomerThemeSettingRepository,
	fileRepo data.ThemeFileRepository,
	config ThemeConfig,
) *ThemeUseCase {
	return &ThemeUseCase{
		themeRepo:    themeRepo,
		variableRepo: variableRepo,
		configRepo:   configRepo,
		customerRepo: customerRepo,
		fileRepo:     fileRepo,
		config:       config,
	}
}

// CreateTheme 创建主题
func (uc *ThemeUseCase) CreateTheme(ctx context.Context, req *models.ThemeCreateRequest) (*models.Theme, error) {
	// 检查名称是否已存在
	existing, err := uc.themeRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, data.ErrThemeNameExists
	}

	theme := &models.Theme{
		Name:            req.Name,
		Title:           req.Title,
		Description:     req.Description,
		Author:          req.Author,
		Version:         req.Version,
		PreviewImageURL: req.PreviewImageURL,
		ThemePath:       req.ThemePath,
		SupportRtl:      req.SupportRtl,
		IsDefault:       req.IsDefault,
		IsActive:        true,
	}

	// 如果设置为默认主题，需要取消其他默认主题
	if req.IsDefault {
		if err := uc.clearDefaultTheme(ctx); err != nil {
			return nil, err
		}
	}

	if err := uc.themeRepo.Create(ctx, theme); err != nil {
		return nil, err
	}

	return theme, nil
}

// GetTheme 获取主题详情
func (uc *ThemeUseCase) GetTheme(ctx context.Context, id uint) (*models.Theme, error) {
	theme, err := uc.themeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}
	return theme, nil
}

// GetThemeWithVariables 获取主题及其变量
func (uc *ThemeUseCase) GetThemeWithVariables(ctx context.Context, id uint) (*models.ThemePreview, error) {
	theme, err := uc.themeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}

	variables, err := uc.variableRepo.GetByThemeID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 转换为预览结构
	var vars []models.ThemeVariable
	for _, v := range variables {
		vars = append(vars, *v)
	}

	return &models.ThemePreview{
		ID:              theme.ID,
		Name:            theme.Name,
		Title:           theme.Title,
		Description:     theme.Description,
		PreviewImageURL: theme.PreviewImageURL,
		Author:          theme.Author,
		Version:         theme.Version,
		SupportRtl:      theme.SupportRtl,
		IsDefault:       theme.IsDefault,
		Variables:       vars,
	}, nil
}

// ListThemes 主题列表
func (uc *ThemeUseCase) ListThemes(ctx context.Context, page, pageSize int) ([]*models.Theme, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return uc.themeRepo.List(ctx, page, pageSize)
}

// ListActiveThemes 获取激活的主题列表
func (uc *ThemeUseCase) ListActiveThemes(ctx context.Context) ([]*models.Theme, error) {
	return uc.themeRepo.ListActive(ctx)
}

// UpdateTheme 更新主题
func (uc *ThemeUseCase) UpdateTheme(ctx context.Context, id uint, req *models.ThemeUpdateRequest) (*models.Theme, error) {
	theme, err := uc.themeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}

	// 更新字段
	theme.Title = req.Title
	theme.Description = req.Description
	theme.Author = req.Author
	theme.Version = req.Version
	theme.PreviewImageURL = req.PreviewImageURL
	theme.SupportRtl = req.SupportRtl
	theme.IsActive = req.IsActive

	// 处理默认主题变更
	if req.IsDefault && !theme.IsDefault {
		if err := uc.clearDefaultTheme(ctx); err != nil {
			return nil, err
		}
		theme.IsDefault = true
	} else if !req.IsDefault && theme.IsDefault {
		// 不允许取消默认主题，需要有其他默认主题
		return nil, data.ErrCannotDeleteDefault
	}

	if err := uc.themeRepo.Update(ctx, theme); err != nil {
		return nil, err
	}

	return theme, nil
}

// DeleteTheme 删除主题
func (uc *ThemeUseCase) DeleteTheme(ctx context.Context, id uint) error {
	theme, err := uc.themeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return data.ErrThemeNotFound
		}
		return err
	}

	// 不允许删除默认主题
	if theme.IsDefault {
		return data.ErrCannotDeleteDefault
	}

	// 删除相关变量
	if err := uc.variableRepo.DeleteByThemeID(ctx, id); err != nil {
		return err
	}

	// 删除相关文件
	if err := uc.fileRepo.DeleteByThemeID(ctx, id); err != nil {
		return err
	}

	return uc.themeRepo.Delete(ctx, id)
}

// GetDefaultTheme 获取默认主题
func (uc *ThemeUseCase) GetDefaultTheme(ctx context.Context) (*models.Theme, error) {
	theme, err := uc.themeRepo.GetDefault(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}
	return theme, nil
}

// clearDefaultTheme 清除所有默认主题标记
func (uc *ThemeUseCase) clearDefaultTheme(ctx context.Context) error {
	themes, _, err := uc.themeRepo.List(ctx, 1, 1000)
	if err != nil {
		return err
	}
	for _, t := range themes {
		if t.IsDefault {
			t.IsDefault = false
			if err := uc.themeRepo.Update(ctx, t); err != nil {
				return err
			}
		}
	}
	return nil
}

// ========== 主题变量操作 ==========

// CreateVariable 创建主题变量
func (uc *ThemeUseCase) CreateVariable(ctx context.Context, req *models.ThemeVariableCreateRequest) (*models.ThemeVariable, error) {
	// 验证主题存在
	_, err := uc.themeRepo.GetByID(ctx, req.ThemeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}

	variable := &models.ThemeVariable{
		ThemeID:      req.ThemeID,
		Name:         req.Name,
		Value:        req.Value,
		Type:         req.Type,
		Category:     req.Category,
		DisplayOrder: req.DisplayOrder,
	}

	if err := uc.variableRepo.Create(ctx, variable); err != nil {
		return nil, err
	}

	return variable, nil
}

// GetVariablesByThemeID 获取主题的所有变量
func (uc *ThemeUseCase) GetVariablesByThemeID(ctx context.Context, themeID uint) ([]*models.ThemeVariable, error) {
	return uc.variableRepo.GetByThemeID(ctx, themeID)
}

// UpdateVariable 更新主题变量
func (uc *ThemeUseCase) UpdateVariable(ctx context.Context, id uint, req *models.ThemeVariableUpdateRequest) (*models.ThemeVariable, error) {
	variable, err := uc.variableRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeVariableNotFound
		}
		return nil, err
	}

	variable.Value = req.Value
	variable.Type = req.Type
	variable.Category = req.Category
	variable.DisplayOrder = req.DisplayOrder

	if err := uc.variableRepo.Update(ctx, variable); err != nil {
		return nil, err
	}

	return variable, nil
}

// DeleteVariable 删除主题变量
func (uc *ThemeUseCase) DeleteVariable(ctx context.Context, id uint) error {
	return uc.variableRepo.Delete(ctx, id)
}

// ========== 主题配置操作 ==========

// ThemeConfigurationUseCase 主题配置用例
type ThemeConfigurationUseCase struct {
	configRepo data.ThemeConfigurationRepository
	themeRepo  data.ThemeRepository
}

// NewThemeConfigurationUseCase 创建主题配置用例
func NewThemeConfigurationUseCase(
	configRepo data.ThemeConfigurationRepository,
	themeRepo data.ThemeRepository,
) *ThemeConfigurationUseCase {
	return &ThemeConfigurationUseCase{
		configRepo: configRepo,
		themeRepo:  themeRepo,
	}
}

// GetStoreConfiguration 获取店铺的主题配置
func (uc *ThemeConfigurationUseCase) GetStoreConfiguration(ctx context.Context, storeID uint) ([]*models.ThemeConfiguration, error) {
	return uc.configRepo.GetByStore(ctx, storeID)
}

// GetThemeConfiguration 获取特定主题和店铺的配置
func (uc *ThemeConfigurationUseCase) GetThemeConfiguration(ctx context.Context, themeID, storeID uint) (*models.ThemeConfiguration, error) {
	config, err := uc.configRepo.GetByThemeAndStore(ctx, themeID, storeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeConfigurationNotFound
		}
		return nil, err
	}
	return config, nil
}

// UpdateConfiguration 更新主题配置
func (uc *ThemeConfigurationUseCase) UpdateConfiguration(ctx context.Context, req *models.ThemeConfigurationUpdateRequest) (*models.ThemeConfiguration, error) {
	// 验证主题存在
	_, err := uc.themeRepo.GetByID(ctx, req.ThemeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}

	// 查找现有配置
	config, err := uc.configRepo.GetByThemeAndStore(ctx, req.ThemeID, req.StoreID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 序列化配置
	configJSON, err := serializeConfig(req.Configuration)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}

	if config == nil {
		// 创建新配置
		config = &models.ThemeConfiguration{
			ThemeID:       req.ThemeID,
			StoreID:       req.StoreID,
			Configuration: configJSON,
			IsActive:      true,
		}
		if err := uc.configRepo.Create(ctx, config); err != nil {
			return nil, err
		}
	} else {
		// 更新现有配置
		config.Configuration = configJSON
		if err := uc.configRepo.Update(ctx, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// DeleteConfiguration 删除主题配置
func (uc *ThemeConfigurationUseCase) DeleteConfiguration(ctx context.Context, id uint) error {
	return uc.configRepo.Delete(ctx, id)
}

// serializeConfig 序列化配置为JSON字符串
func serializeConfig(config map[string]interface{}) (string, error) {
	if config == nil {
		return "{}", nil
	}
	// 简单实现，实际项目中应使用 encoding/json
	data, err := serializeJSON(config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// serializeJSON JSON序列化
func serializeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// ========== 客户主题设置操作 ==========

// CustomerThemeUseCase 客户主题用例
type CustomerThemeUseCase struct {
	customerRepo data.CustomerThemeSettingRepository
	themeRepo    data.ThemeRepository
}

// NewCustomerThemeUseCase 创建客户主题用例
func NewCustomerThemeUseCase(
	customerRepo data.CustomerThemeSettingRepository,
	themeRepo data.ThemeRepository,
) *CustomerThemeUseCase {
	return &CustomerThemeUseCase{
		customerRepo: customerRepo,
		themeRepo:    themeRepo,
	}
}

// GetCustomerTheme 获取客户主题设置
func (uc *CustomerThemeUseCase) GetCustomerTheme(ctx context.Context, customerID uint) (*models.CustomerThemeSetting, error) {
	setting, err := uc.customerRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrCustomerThemeNotFound
		}
		return nil, err
	}
	return setting, nil
}

// SetCustomerTheme 设置客户主题
func (uc *CustomerThemeUseCase) SetCustomerTheme(ctx context.Context, req *models.CustomerThemeRequest) (*models.CustomerThemeSetting, error) {
	// 验证主题存在
	theme, err := uc.themeRepo.GetByID(ctx, req.ThemeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeNotFound
		}
		return nil, err
	}

	// 检查主题是否激活
	if !theme.IsActive {
		return nil, errors.New("主题未激活")
	}

	// 序列化设置
	settingsJSON, err := serializeConfig(req.Settings)
	if err != nil {
		return nil, fmt.Errorf("序列化设置失败: %w", err)
	}

	// 查找现有设置
	setting, err := uc.customerRepo.GetByCustomerID(ctx, req.CustomerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if setting == nil {
		// 创建新设置
		setting = &models.CustomerThemeSetting{
			CustomerID: req.CustomerID,
			ThemeID:    req.ThemeID,
			Settings:   settingsJSON,
		}
		if err := uc.customerRepo.Create(ctx, setting); err != nil {
			return nil, err
		}
	} else {
		// 更新现有设置
		setting.ThemeID = req.ThemeID
		setting.Settings = settingsJSON
		if err := uc.customerRepo.Update(ctx, setting); err != nil {
			return nil, err
		}
	}

	return setting, nil
}

// DeleteCustomerTheme 删除客户主题设置
func (uc *CustomerThemeUseCase) DeleteCustomerTheme(ctx context.Context, customerID uint) error {
	return uc.customerRepo.Delete(ctx, customerID)
}

// ========== 主题文件操作 ==========

// ThemeFileUseCase 主题文件用例
type ThemeFileUseCase struct {
	fileRepo  data.ThemeFileRepository
	themeRepo data.ThemeRepository
}

// NewThemeFileUseCase 创建主题文件用例
func NewThemeFileUseCase(
	fileRepo data.ThemeFileRepository,
	themeRepo data.ThemeRepository,
) *ThemeFileUseCase {
	return &ThemeFileUseCase{
		fileRepo:  fileRepo,
		themeRepo: themeRepo,
	}
}

// GetThemeFiles 获取主题的所有文件
func (uc *ThemeFileUseCase) GetThemeFiles(ctx context.Context, themeID uint) ([]*models.ThemeFile, error) {
	return uc.fileRepo.GetByThemeID(ctx, themeID)
}

// GetThemeFile 获取主题文件详情
func (uc *ThemeFileUseCase) GetThemeFile(ctx context.Context, id uint) (*models.ThemeFile, error) {
	file, err := uc.fileRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeFileNotFound
		}
		return nil, err
	}
	return file, nil
}

// GetThemeFileByPath 通过路径获取主题文件
func (uc *ThemeFileUseCase) GetThemeFileByPath(ctx context.Context, themeID uint, filePath string) (*models.ThemeFile, error) {
	file, err := uc.fileRepo.GetByThemeAndPath(ctx, themeID, filePath)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, data.ErrThemeFileNotFound
		}
		return nil, err
	}
	return file, nil
}

// CreateThemeFile 创建主题文件
func (uc *ThemeFileUseCase) CreateThemeFile(ctx context.Context, file *models.ThemeFile) error {
	// 验证主题存在
	_, err := uc.themeRepo.GetByID(ctx, file.ThemeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return data.ErrThemeNotFound
		}
		return err
	}

	return uc.fileRepo.Create(ctx, file)
}

// UpdateThemeFile 更新主题文件
func (uc *ThemeFileUseCase) UpdateThemeFile(ctx context.Context, file *models.ThemeFile) error {
	return uc.fileRepo.Update(ctx, file)
}

// DeleteThemeFile 删除主题文件
func (uc *ThemeFileUseCase) DeleteThemeFile(ctx context.Context, id uint) error {
	return uc.fileRepo.Delete(ctx, id)
}
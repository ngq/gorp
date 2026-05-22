package service

import (
	"context"

	"nop-go/services/localization/internal/biz"
	"nop-go/services/localization/internal/data"

	"gorm.io/gorm"
)

// Services 本地化服务集合。
type Services struct {
	Localization *LocalizationService
}

// NewServices 创建本地化服务集合。
func NewServices(db *gorm.DB) *Services {
	langRepo := data.NewLanguageRepo(db)
	resRepo := data.NewLocaleResourceRepo(db)
	locUC := biz.NewLocalizationUseCase(langRepo, resRepo)
	return &Services{
		Localization: &LocalizationService{uc: locUC},
	}
}

// LocalizationService 本地化服务。
type LocalizationService struct {
	uc *biz.LocalizationUseCase
}

// CreateLanguageRequest 创建语言请求。
type CreateLanguageRequest struct {
	Name              string `json:"name" binding:"required"`                // 语言名称
	LanguageCulture   string `json:"language_culture" binding:"required"`     // 语言文化代码
	UniqueSeoCode     string `json:"unique_seo_code" binding:"required"`     // SEO唯一代码
	FlagImageFileName string `json:"flag_image_file_name"`                   // 国旗图片文件名
	Rtl               bool   `json:"rtl"`                                    // 是否从右到左书写
	IsActive          bool   `json:"is_active"`                              // 是否启用
	DisplayOrder      int    `json:"display_order"`                          // 显示排序
}

// UpdateLanguageRequest 更新语言请求。
type UpdateLanguageRequest struct {
	Name              string `json:"name" binding:"required"`                // 语言名称
	LanguageCulture   string `json:"language_culture" binding:"required"`     // 语言文化代码
	UniqueSeoCode     string `json:"unique_seo_code" binding:"required"`     // SEO唯一代码
	FlagImageFileName string `json:"flag_image_file_name"`                   // 国旗图片文件名
	Rtl               bool   `json:"rtl"`                                    // 是否从右到左书写
	IsActive          bool   `json:"is_active"`                              // 是否启用
	DisplayOrder      int    `json:"display_order"`                          // 显示排序
}

// CreateLocaleResourceRequest 创建本地化资源请求。
type CreateLocaleResourceRequest struct {
	ResourceName  string `json:"resource_name" binding:"required"`  // 资源名称（键）
	ResourceValue string `json:"resource_value" binding:"required"` // 资源值
}

// UpdateLocaleResourceRequest 更新本地化资源请求。
type UpdateLocaleResourceRequest struct {
	ResourceName  string `json:"resource_name" binding:"required"`  // 资源名称（键）
	ResourceValue string `json:"resource_value" binding:"required"` // 资源值
}

// LanguageResponse 语言响应。
type LanguageResponse struct {
	ID                uint   `json:"id"`                  // 语言ID
	Name              string `json:"name"`                // 语言名称
	LanguageCulture   string `json:"language_culture"`    // 语言文化代码
	UniqueSeoCode     string `json:"unique_seo_code"`     // SEO唯一代码
	FlagImageFileName string `json:"flag_image_file_name"` // 国旗图片文件名
	Rtl               bool   `json:"rtl"`                 // 是否从右到左书写
	IsActive          bool   `json:"is_active"`           // 是否启用
	DisplayOrder      int    `json:"display_order"`       // 显示排序
	CreatedAt         string `json:"created_at"`          // 创建时间
	UpdatedAt         string `json:"updated_at"`          // 更新时间
}

// LocaleResourceResponse 本地化资源响应。
type LocaleResourceResponse struct {
	ID            uint   `json:"id"`             // 资源ID
	LanguageID    uint   `json:"language_id"`    // 所属语言ID
	ResourceName  string `json:"resource_name"`  // 资源名称（键）
	ResourceValue string `json:"resource_value"` // 资源值
	CreatedAt     string `json:"created_at"`     // 创建时间
	UpdatedAt     string `json:"updated_at"`     // 更新时间
}

// toLanguageResponse 将语言领域实体转换为响应结构体。
func toLanguageResponse(lang *biz.Language) *LanguageResponse {
	return &LanguageResponse{
		ID:                lang.ID,
		Name:              lang.Name,
		LanguageCulture:   lang.LanguageCulture,
		UniqueSeoCode:     lang.UniqueSeoCode,
		FlagImageFileName: lang.FlagImageFileName,
		Rtl:               lang.Rtl,
		IsActive:          lang.IsActive,
		DisplayOrder:      lang.DisplayOrder,
		CreatedAt:         lang.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         lang.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toLocaleResourceResponse 将资源领域实体转换为响应结构体。
func toLocaleResourceResponse(res *biz.LocaleResource) *LocaleResourceResponse {
	return &LocaleResourceResponse{
		ID:            res.ID,
		LanguageID:    res.LanguageID,
		ResourceName:  res.ResourceName,
		ResourceValue: res.ResourceValue,
		CreatedAt:     res.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     res.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListLanguages 获取语言列表。
func (s *LocalizationService) ListLanguages(ctx context.Context, page, size int) ([]LanguageResponse, int64, error) {
	langs, total, err := s.uc.ListLanguages(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]LanguageResponse, len(langs))
	for i, lang := range langs {
		items[i] = *toLanguageResponse(lang)
	}
	return items, total, nil
}

// CreateLanguage 创建语言。
func (s *LocalizationService) CreateLanguage(ctx context.Context, req CreateLanguageRequest) (*LanguageResponse, error) {
	lang, err := s.uc.CreateLanguage(ctx, req.Name, req.LanguageCulture, req.UniqueSeoCode, req.FlagImageFileName, req.Rtl, req.IsActive, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toLanguageResponse(lang), nil
}

// UpdateLanguage 更新语言。
func (s *LocalizationService) UpdateLanguage(ctx context.Context, id uint, req UpdateLanguageRequest) (*LanguageResponse, error) {
	lang, err := s.uc.UpdateLanguage(ctx, id, req.Name, req.LanguageCulture, req.UniqueSeoCode, req.FlagImageFileName, req.Rtl, req.IsActive, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toLanguageResponse(lang), nil
}

// DeleteLanguage 删除语言。
func (s *LocalizationService) DeleteLanguage(ctx context.Context, id uint) error {
	return s.uc.DeleteLanguage(ctx, id)
}

// ListResources 获取本地化资源列表。
func (s *LocalizationService) ListResources(ctx context.Context, languageID uint, page, size int) ([]LocaleResourceResponse, int64, error) {
	resources, total, err := s.uc.ListResources(ctx, languageID, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]LocaleResourceResponse, len(resources))
	for i, res := range resources {
		items[i] = *toLocaleResourceResponse(res)
	}
	return items, total, nil
}

// AddResource 添加本地化资源。
func (s *LocalizationService) AddResource(ctx context.Context, languageID uint, req CreateLocaleResourceRequest) (*LocaleResourceResponse, error) {
	res, err := s.uc.AddResource(ctx, languageID, req.ResourceName, req.ResourceValue)
	if err != nil {
		return nil, err
	}
	return toLocaleResourceResponse(res), nil
}

// UpdateResource 更新本地化资源。
func (s *LocalizationService) UpdateResource(ctx context.Context, id uint, req UpdateLocaleResourceRequest) (*LocaleResourceResponse, error) {
	res, err := s.uc.UpdateResource(ctx, id, req.ResourceName, req.ResourceValue)
	if err != nil {
		return nil, err
	}
	return toLocaleResourceResponse(res), nil
}

// DeleteResource 删除本地化资源。
func (s *LocalizationService) DeleteResource(ctx context.Context, id uint) error {
	return s.uc.DeleteResource(ctx, id)
}

// ExportResources 导出语言资源。
func (s *LocalizationService) ExportResources(ctx context.Context, languageID uint) ([]LocaleResourceResponse, error) {
	resources, err := s.uc.ExportResources(ctx, languageID)
	if err != nil {
		return nil, err
	}

	items := make([]LocaleResourceResponse, len(resources))
	for i, res := range resources {
		items[i] = *toLocaleResourceResponse(res)
	}
	return items, nil
}

// ImportResources 导入语言资源。
func (s *LocalizationService) ImportResources(ctx context.Context, languageID uint, resources []CreateLocaleResourceRequest) error {
	// 将请求结构体转换为用例所需的格式
	items := make([]struct {
		ResourceName  string
		ResourceValue string
	}, len(resources))
	for i, r := range resources {
		items[i].ResourceName = r.ResourceName
		items[i].ResourceValue = r.ResourceValue
	}
	return s.uc.ImportResources(ctx, languageID, items)
}
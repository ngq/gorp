// Package models 本地化服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// Language 语言
type Language struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"size:100;not null" json:"name"`                  // 语言名称如 "English"
	LanguageCulture    string         `gorm:"size:20;not null;uniqueIndex" json:"language_culture"` // 文化代码如 "en-US"
	UniqueSeoCode      string         `gorm:"size:10;not null" json:"unique_seo_code"`        // SEO代码如 "en"
	Published          bool           `gorm:"default:true" json:"published"`                  // 是否发布
	DisplayOrder       int            `gorm:"default:0" json:"display_order"`                 // 显示顺序
	Rtl                bool           `gorm:"default:false" json:"rtl"`                       // 是否从右到左语言
	FlagImageFileName  string         `gorm:"size:255" json:"flag_image_file_name"`          // 旗帜图片文件名
	DefaultCurrencyID  uint           `json:"default_currency_id"`                            // 默认货币ID
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Language) TableName() string {
	return "languages"
}

// LocaleStringResource 本地化字符串资源
type LocaleStringResource struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	LanguageID   uint      `gorm:"not null;index" json:"language_id"`       // 语言ID
	ResourceName string    `gorm:"size:200;not null;index" json:"resource_name"` // 资源名称(键)
	ResourceValue string   `gorm:"type:text;not null" json:"resource_value"` // 资源值(翻译文本)
	IsTouched     bool      `gorm:"default:false" json:"is_touched"`        // 是否被修改过
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (LocaleStringResource) TableName() string {
	return "locale_string_resources"
}

// LocaleStringResourceTranslationKey 翻译键分组
// 用于管理资源名称的分组和分类
type LocaleStringResourceTranslationKey struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	KeyName     string         `gorm:"size:200;not null;uniqueIndex" json:"key_name"` // 键名称
	KeyGroup    string         `gorm:"size:100;not null;index" json:"key_group"`     // 键所属分组
	Description string         `gorm:"type:text" json:"description"`                // 键描述
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LocaleStringResourceTranslationKey) TableName() string {
	return "locale_string_resource_translation_keys"
}

// Currency 货币
type Currency struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	Name                  string         `gorm:"size:50;not null" json:"name"`                     // 货币名称
	CurrencyCode          string         `gorm:"size:10;not null;uniqueIndex" json:"currency_code"` // 货币代码如 USD
	Rate                  float64        `gorm:"type:decimal(18,8);default:1" json:"rate"`        // 汇率
	DisplayLocale         string         `gorm:"size:20" json:"display_locale"`                    // 显示格式区域
	CustomFormatting      string         `gorm:"size:50" json:"custom_formatting"`                 // 自定义格式
	Published             bool           `gorm:"default:true" json:"published"`                    // 是否发布
	DisplayOrder          int            `gorm:"default:0" json:"display_order"`                   // 显示顺序
	CreatedOnUtc          time.Time      `json:"created_on_utc"`
	UpdatedOnUtc          time.Time      `json:"updated_on_utc"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Currency) TableName() string {
	return "currencies"
}

// LanguageCreateRequest 语言创建请求
type LanguageCreateRequest struct {
	Name              string `json:"name" binding:"required"`
	LanguageCulture   string `json:"language_culture" binding:"required"`
	UniqueSeoCode     string `json:"unique_seo_code" binding:"required"`
	Published         bool   `json:"published"`
	DisplayOrder      int    `json:"display_order"`
	Rtl               bool   `json:"rtl"`
	FlagImageFileName string `json:"flag_image_file_name"`
	DefaultCurrencyID uint   `json:"default_currency_id"`
}

// LanguageUpdateRequest 语言更新请求
type LanguageUpdateRequest struct {
	Name              string `json:"name"`
	LanguageCulture   string `json:"language_culture"`
	UniqueSeoCode     string `json:"unique_seo_code"`
	Published         bool   `json:"published"`
	DisplayOrder      int    `json:"display_order"`
	Rtl               bool   `json:"rtl"`
	FlagImageFileName string `json:"flag_image_file_name"`
	DefaultCurrencyID uint   `json:"default_currency_id"`
}

// ResourceCreateRequest 资源创建请求
type ResourceCreateRequest struct {
	LanguageID    uint   `json:"language_id" binding:"required"`
	ResourceName  string `json:"resource_name" binding:"required"`
	ResourceValue string `json:"resource_value" binding:"required"`
	KeyGroup      string `json:"key_group"` // 资源分组(可选)
}

// ResourceUpdateRequest 资源更新请求
type ResourceUpdateRequest struct {
	ResourceValue string `json:"resource_value" binding:"required"`
}

// ResourceBatchUpdateRequest 批量更新资源请求
type ResourceBatchUpdateRequest struct {
	LanguageID uint                      `json:"language_id" binding:"required"`
	Resources  []ResourceUpdateItem      `json:"resources" binding:"required"`
}

// ResourceUpdateItem 单个资源更新项
type ResourceUpdateItem struct {
	ResourceName  string `json:"resource_name" binding:"required"`
	ResourceValue string `json:"resource_value" binding:"required"`
}

// TranslationResult 翻译结果
type TranslationResult struct {
	ResourceName  string `json:"resource_name"`
	ResourceValue string `json:"resource_value"`
	LanguageID    uint   `json:"language_id"`
	LanguageCode  string `json:"language_code"`
}
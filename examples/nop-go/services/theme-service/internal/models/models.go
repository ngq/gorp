// Package models 主题服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// Theme 主题
type Theme struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:128;not null;uniqueIndex" json:"name"`        // 主题名称(系统标识)
	Title           string         `gorm:"size:256;not null" json:"title"`                    // 显示标题
	Description     string         `gorm:"type:text" json:"description"`                      // 描述
	Author          string         `gorm:"size:128" json:"author"`                            // 作者
	Version         string         `gorm:"size:32" json:"version"`                            // 版本
	PreviewImageURL string         `gorm:"size:512" json:"preview_image_url"`                // 预览图片
	ThemePath       string         `gorm:"size:256;not null" json:"theme_path"`              // 主题路径
	SupportRtl      bool           `gorm:"default:false" json:"support_rtl"`                 // 是否支持RTL
	IsDefault       bool           `gorm:"default:false" json:"is_default"`                   // 是否默认主题
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Theme) TableName() string {
	return "themes"
}

// ThemeVariable 主题变量
type ThemeVariable struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ThemeID     uint      `gorm:"not null;index" json:"theme_id"`               // 主题ID
	Name        string    `gorm:"size:128;not null" json:"name"`                // 变量名
	Value       string    `gorm:"type:text" json:"value"`                       // 变量值
	Type        string    `gorm:"size:32;not null" json:"type"`                 // 类型: color/image/font/text
	Category    string    `gorm:"size:64" json:"category"`                      // 分类: general/header/footer等
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ThemeVariable) TableName() string {
	return "theme_variables"
}

// ThemeConfiguration 主题配置（店铺级别）
type ThemeConfiguration struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ThemeID          uint      `gorm:"not null;index" json:"theme_id"`           // 主题ID
	StoreID          uint      `gorm:"not null;index" json:"store_id"`           // 店铺ID
	Configuration    string    `gorm:"type:json" json:"configuration"`           // 配置JSON
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ThemeConfiguration) TableName() string {
	return "theme_configurations"
}

// CustomerThemeSetting 客户主题设置
type CustomerThemeSetting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CustomerID  uint      `gorm:"not null;uniqueIndex" json:"customer_id"`   // 客户ID
	ThemeID     uint      `gorm:"not null" json:"theme_id"`                  // 主题ID
	Settings    string    `gorm:"type:json" json:"settings"`                 // 用户自定义设置
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CustomerThemeSetting) TableName() string {
	return "customer_theme_settings"
}

// ThemeFile 主题文件
type ThemeFile struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ThemeID      uint      `gorm:"not null;index" json:"theme_id"`
	FilePath     string    `gorm:"size:512;not null" json:"file_path"`       // 文件路径
	FileName     string    `gorm:"size:256;not null" json:"file_name"`       // 文件名
	FileType     string    `gorm:"size:32" json:"file_type"`                 // 类型: template/style/script
	IsEditable   bool      `gorm:"default:true" json:"is_editable"`          // 是否可编辑
	Content      string    `gorm:"type:longtext" json:"content"`             // 文件内容
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ThemeFile) TableName() string {
	return "theme_files"
}

// ========== DTO ==========

// ThemeCreateRequest 主题创建请求
type ThemeCreateRequest struct {
	Name            string `json:"name" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	Author          string `json:"author"`
	Version         string `json:"version"`
	PreviewImageURL string `json:"preview_image_url"`
	ThemePath       string `json:"theme_path" binding:"required"`
	SupportRtl      bool   `json:"support_rtl"`
	IsDefault       bool   `json:"is_default"`
}

// ThemeUpdateRequest 主题更新请求
type ThemeUpdateRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Author          string `json:"author"`
	Version         string `json:"version"`
	PreviewImageURL string `json:"preview_image_url"`
	SupportRtl      bool   `json:"support_rtl"`
	IsDefault       bool   `json:"is_default"`
	IsActive        bool   `json:"is_active"`
}

// ThemeVariableCreateRequest 主题变量创建请求
type ThemeVariableCreateRequest struct {
	ThemeID       uint   `json:"theme_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Value         string `json:"value"`
	Type          string `json:"type" binding:"required"`
	Category      string `json:"category"`
	DisplayOrder  int    `json:"display_order"`
}

// ThemeVariableUpdateRequest 主题变量更新请求
type ThemeVariableUpdateRequest struct {
	Value        string `json:"value"`
	Type         string `json:"type"`
	Category     string `json:"category"`
	DisplayOrder int    `json:"display_order"`
}

// ThemeConfigurationUpdateRequest 主题配置更新请求
type ThemeConfigurationUpdateRequest struct {
	ThemeID       uint                   `json:"theme_id" binding:"required"`
	StoreID       uint                   `json:"store_id"`
	Configuration map[string]interface{} `json:"configuration"`
}

// CustomerThemeRequest 客户主题设置请求
type CustomerThemeRequest struct {
	CustomerID uint                   `json:"customer_id" binding:"required"`
	ThemeID    uint                   `json:"theme_id" binding:"required"`
	Settings   map[string]interface{} `json:"settings"`
}

// ThemePreview 主题预览信息
type ThemePreview struct {
	ID              uint            `json:"id"`
	Name            string          `json:"name"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	PreviewImageURL string          `json:"preview_image_url"`
	Author          string          `json:"author"`
	Version         string          `json:"version"`
	SupportRtl      bool            `json:"support_rtl"`
	IsDefault       bool            `json:"is_default"`
	Variables       []ThemeVariable `json:"variables"`
}
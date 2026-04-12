// Package models SEO服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// UrlRecord URL记录
type UrlRecord struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Slug        string         `gorm:"size:200;not null;uniqueIndex" json:"slug"`    // URL slug
	EntityID    uint           `gorm:"not null;index" json:"entity_id"`               // 实体ID
	EntityType  string         `gorm:"size:50;not null;index" json:"entity_type"`     // 实体类型(product/category/brand等)
	LanguageID  uint           `gorm:"default:0;index" json:"language_id"`           // 语言ID(0表示默认)
	IsActive    bool           `gorm:"default:true" json:"is_active"`                 // 是否激活
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (UrlRecord) TableName() string {
	return "url_records"
}

// UrlRedirect URL重定向记录
type UrlRedirect struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	OldSlug        string         `gorm:"size:200;not null;uniqueIndex" json:"old_slug"`   // 旧URL
	NewSlug        string         `gorm:"size:200;not null" json:"new_slug"`               // 新URL
	RedirectType   int            `gorm:"default:301" json:"redirect_type"`               // 重定向类型(301/302)
	IsActive       bool           `gorm:"default:true" json:"is_active"`                  // 是否激活
	HitCount       int            `gorm:"default:0" json:"hit_count"`                     // 访问次数
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (UrlRedirect) TableName() string {
	return "url_redirects"
}

// MetaInfo 元信息
type MetaInfo struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	EntityID      uint           `gorm:"not null;index" json:"entity_id"`                 // 实体ID
	EntityType    string         `gorm:"size:50;not null;index" json:"entity_type"`       // 实体类型
	LanguageID    uint           `gorm:"default:0;index" json:"language_id"`              // 语言ID
	MetaTitle     string         `gorm:"size:400" json:"meta_title"`                      // SEO标题
	MetaKeywords  string         `gorm:"size:400" json:"meta_keywords"`                   // SEO关键词
	MetaDescription string       `gorm:"type:text" json:"meta_description"`               // SEO描述
	PageTitle     string         `gorm:"size:400" json:"page_title"`                      // 页面标题
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MetaInfo) TableName() string {
	return "meta_infos"
}

// SitemapNode Sitemap节点
type SitemapNode struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	EntityType      string    `gorm:"size:50;not null;index" json:"entity_type"`     // 实体类型
	EntityID        uint      `gorm:"not null;index" json:"entity_id"`               // 实体ID
	URL             string    `gorm:"size:500;not null" json:"url"`                  // URL
	LastModified    time.Time `json:"last_modified"`                                 // 最后修改时间
	ChangeFrequency string    `gorm:"size:20" json:"change_frequency"`               // 更新频率
	Priority        float64   `gorm:"type:decimal(3,1)" json:"priority"`             // 优先级(0-1)
	LanguageID      uint      `gorm:"default:0" json:"language_id"`                  // 语言ID
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SitemapNode) TableName() string {
	return "sitemap_nodes"
}

// UrlRecordCreateRequest URL记录创建请求
type UrlRecordCreateRequest struct {
	Slug       string `json:"slug" binding:"required"`
	EntityID   uint   `json:"entity_id" binding:"required"`
	EntityType string `json:"entity_type" binding:"required"`
	LanguageID uint   `json:"language_id"`
	IsActive   bool   `json:"is_active"`
}

// UrlRecordUpdateRequest URL记录更新请求
type UrlRecordUpdateRequest struct {
	Slug       string `json:"slug"`
	IsActive   bool   `json:"is_active"`
}

// UrlRedirectCreateRequest URL重定向创建请求
type UrlRedirectCreateRequest struct {
	OldSlug      string `json:"old_slug" binding:"required"`
	NewSlug      string `json:"new_slug" binding:"required"`
	RedirectType int    `json:"redirect_type"` // 301 or 302
	IsActive     bool   `json:"is_active"`
}

// UrlRedirectUpdateRequest URL重定向更新请求
type UrlRedirectUpdateRequest struct {
	NewSlug      string `json:"new_slug"`
	RedirectType int    `json:"redirect_type"`
	IsActive     bool   `json:"is_active"`
}

// MetaInfoCreateRequest 元信息创建请求
type MetaInfoCreateRequest struct {
	EntityID        uint   `json:"entity_id" binding:"required"`
	EntityType      string `json:"entity_type" binding:"required"`
	LanguageID      uint   `json:"language_id"`
	MetaTitle       string `json:"meta_title"`
	MetaKeywords    string `json:"meta_keywords"`
	MetaDescription string `json:"meta_description"`
	PageTitle       string `json:"page_title"`
}

// MetaInfoUpdateRequest 元信息更新请求
type MetaInfoUpdateRequest struct {
	MetaTitle       string `json:"meta_title"`
	MetaKeywords    string `json:"meta_keywords"`
	MetaDescription string `json:"meta_description"`
	PageTitle       string `json:"page_title"`
}

// SitemapConfig Sitemap配置
type SitemapConfig struct {
	Enabled           bool    `json:"enabled"`
	IncludeProducts   bool    `json:"include_products"`
	IncludeCategories bool    `json:"include_categories"`
	IncludeBrands     bool    `json:"include_brands"`
	ChangeFrequency   string  `json:"change_frequency"`
	Priority          float64 `json:"priority"`
}

// SitemapGenerationResult Sitemap生成结果
type SitemapGenerationResult struct {
	TotalNodes   int64  `json:"total_nodes"`
	GeneratedAt  string `json:"generated_at"`
	SitemapURL   string `json:"sitemap_url"`
	FileSize     int64  `json:"file_size"`
}

// SEOAnalysisResult SEO分析结果
type SEOAnalysisResult struct {
	EntityID         uint     `json:"entity_id"`
	EntityType       string   `json:"entity_type"`
	Slug             string   `json:"slug"`
	MetaTitle        string   `json:"meta_title"`
	MetaKeywords     string   `json:"meta_keywords"`
	MetaDescription  string   `json:"meta_description"`
	TitleLength      int      `json:"title_length"`
	DescriptionLength int     `json:"description_length"`
	Issues           []string `json:"issues"` // SEO问题列表
	Score            int      `json:"score"` // SEO评分
}
package response

import "time"

// Language 语言响应结构体。
type Language struct {
	ID                uint      `json:"id"`                  // 语言ID
	Name              string    `json:"name"`                // 语言名称
	LanguageCulture   string    `json:"language_culture"`    // 语言文化代码
	UniqueSeoCode     string    `json:"unique_seo_code"`    // SEO唯一代码
	FlagImageFileName string    `json:"flag_image_file_name"` // 国旗图片文件名
	Rtl               bool      `json:"rtl"`                // 是否从右到左书写
	IsActive          bool      `json:"is_active"`          // 是否启用
	DisplayOrder      int       `json:"display_order"`      // 显示排序
	CreatedAt         time.Time `json:"created_at"`         // 创建时间
	UpdatedAt         time.Time `json:"updated_at"`         // 更新时间
}

// LanguageList 语言列表响应。
type LanguageList struct {
	Items []Language `json:"items"` // 语言列表
	Total int64     `json:"total"` // 总数
	Page  int       `json:"page"`  // 当前页
	Size  int       `json:"size"`  // 每页大小
}

// LocaleResource 本地化资源响应结构体。
type LocaleResource struct {
	ID            uint      `json:"id"`             // 资源ID
	LanguageID    uint      `json:"language_id"`    // 所属语言ID
	ResourceName  string    `json:"resource_name"`  // 资源名称（键）
	ResourceValue string    `json:"resource_value"` // 资源值
	CreatedAt     time.Time `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time `json:"updated_at"`     // 更新时间
}

// LocaleResourceList 本地化资源列表响应。
type LocaleResourceList struct {
	Items []LocaleResource `json:"items"` // 资源列表
	Total int64            `json:"total"` // 总数
	Page  int              `json:"page"`  // 当前页
	Size  int              `json:"size"`  // 每页大小
}
package request

// CreateLanguage 创建语言请求。
type CreateLanguage struct {
	Name              string `json:"name" binding:"required"`                // 语言名称
	LanguageCulture   string `json:"language_culture" binding:"required"`     // 语言文化代码，如 zh-CN
	UniqueSeoCode     string `json:"unique_seo_code" binding:"required"`     // SEO唯一代码，如 zh
	FlagImageFileName string `json:"flag_image_file_name"`                   // 国旗图片文件名
	Rtl               bool   `json:"rtl"`                                    // 是否从右到左书写
	IsActive          bool   `json:"is_active"`                              // 是否启用
	DisplayOrder      int    `json:"display_order"`                          // 显示排序
}

// UpdateLanguage 更新语言请求。
type UpdateLanguage struct {
	Name              string `json:"name" binding:"required"`                // 语言名称
	LanguageCulture   string `json:"language_culture" binding:"required"`     // 语言文化代码
	UniqueSeoCode     string `json:"unique_seo_code" binding:"required"`     // SEO唯一代码
	FlagImageFileName string `json:"flag_image_file_name"`                   // 国旗图片文件名
	Rtl               bool   `json:"rtl"`                                    // 是否从右到左书写
	IsActive          bool   `json:"is_active"`                              // 是否启用
	DisplayOrder      int    `json:"display_order"`                          // 显示排序
}

// CreateLocaleResource 创建本地化资源请求。
type CreateLocaleResource struct {
	ResourceName  string `json:"resource_name" binding:"required"`  // 资源名称（键）
	ResourceValue string `json:"resource_value" binding:"required"` // 资源值
}

// UpdateLocaleResource 更新本地化资源请求。
type UpdateLocaleResource struct {
	ResourceName  string `json:"resource_name" binding:"required"`  // 资源名称（键）
	ResourceValue string `json:"resource_value" binding:"required"` // 资源值
}

// ImportResources 导入资源请求。
type ImportResources struct {
	Resources []CreateLocaleResource `json:"resources" binding:"required"` // 待导入的资源列表
}
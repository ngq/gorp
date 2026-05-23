package request

// ==================== 语言请求 ====================

// CreateLanguage 创建语言请求
type CreateLanguage struct {
	Code      string `json:"code" binding:"required"`  // 语言代码（必填），如 zh-CN
	Name      string `json:"name" binding:"required"`  // 语言名称（必填），如 中文(中国)
	IsDefault bool   `json:"is_default"`               // 是否为默认语言
	SortOrder int    `json:"sort_order"`                // 排序权重
	IsActive  bool   `json:"is_active"`                  // 是否启用
}

// UpdateLanguage 更新语言请求
type UpdateLanguage struct {
	Code      string `json:"code" binding:"required"`
	Name      string `json:"name" binding:"required"`
	IsDefault bool   `json:"is_default"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

// ==================== 本地化资源请求 ====================

// CreateLocaleResource 创建本地化资源请求
type CreateLocaleResource struct {
	Key    string `json:"key" binding:"required"`  // 翻译键（必填）
	Value  string `json:"value" binding:"required"` // 翻译值（必填）
	Module string `json:"module"`                   // 所属模块
}

// UpdateLocaleResource 更新本地化资源请求
type UpdateLocaleResource struct {
	Key    string `json:"key" binding:"required"`
	Value  string `json:"value" binding:"required"`
	Module string `json:"module"`
}
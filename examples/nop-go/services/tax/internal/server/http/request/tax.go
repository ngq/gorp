// Package request 提供 tax 服务的 HTTP 请求结构体定义
package request

// ==================== 税务提供者 ====================

// UpdateProviderRequest 更新税务提供者请求
// 对应路由: PUT /api/v1/tax/providers/:id
type UpdateProviderRequest struct {
	ID             uint   `json:"id"`               // 提供者 ID（路径参数传入）
	Name           string `json:"name"`             // 提供者名称（如"固定税率计算器"）
	SystemKeyword  string `json:"system_keyword"`   // 系统关键字标识
	DisplayOrder   int    `json:"display_order"`    // 显示排序
	IsActive       bool   `json:"is_active"`        // 是否启用
	IsPrimary      bool   `json:"is_primary"`       // 是否为主要税务提供者
	LogoURL        string `json:"logo_url"`         // 提供者 Logo 地址
}

// ==================== 税类别 ====================

// CreateCategoryRequest 创建税类别请求
// 对应路由: POST /api/v1/tax/categories
type CreateCategoryRequest struct {
	Name         string  `json:"name"`          // 税类别名称（如"标准税率"、"减免税率"）
	Rate         float64 `json:"rate"`          // 税率百分比（如 8.5 表示 8.5%）
	DisplayOrder int     `json:"display_order"` // 显示排序
	IsActive     bool    `json:"is_active"`     // 是否启用
	Description  string  `json:"description"`   // 税类别描述
}

// UpdateCategoryRequest 更新税类别请求
// 对应路由: PUT /api/v1/tax/categories/:id
type UpdateCategoryRequest struct {
	ID           uint    `json:"id"`            // 税类别 ID（路径参数传入）
	Name         string  `json:"name"`          // 税类别名称
	Rate         float64 `json:"rate"`          // 税率百分比
	DisplayOrder int     `json:"display_order"` // 显示排序
	IsActive     bool    `json:"is_active"`     // 是否启用
	Description  string  `json:"description"`   // 税类别描述
}
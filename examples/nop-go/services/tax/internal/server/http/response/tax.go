// Package response 提供 tax 服务的 HTTP 响应结构体定义
package response

// ==================== 税务提供者 ====================

// ProviderResponse 税务提供者响应
// 对应 nopCommerce Admin/Tax/Providers 返回数据
type ProviderResponse struct {
	ID            uint   `json:"id"`              // 提供者 ID
	Name          string `json:"name"`            // 提供者名称
	SystemKeyword string `json:"system_keyword"`  // 系统关键字标识
	DisplayOrder  int    `json:"display_order"`   // 显示排序
	IsActive      bool   `json:"is_active"`       // 是否启用
	IsPrimary     bool   `json:"is_primary"`      // 是否为主要税务提供者
	LogoURL       string `json:"logo_url"`        // Logo 地址
	CreatedAt     int64  `json:"created_at"`      // 创建时间（Unix 时间戳）
	UpdatedAt     int64  `json:"updated_at"`      // 更新时间（Unix 时间戳）
}

// ListProvidersResponse 税务提供者列表响应
type ListProvidersResponse struct {
	Total int64              `json:"total"` // 总数
	Items []*ProviderResponse `json:"items"` // 提供者列表
}

// ==================== 税类别 ====================

// CategoryResponse 税类别响应
// 对应 nopCommerce Admin/Tax/Categories 返回数据
type CategoryResponse struct {
	ID           uint    `json:"id"`            // 税类别 ID
	Name         string  `json:"name"`          // 税类别名称
	Rate         float64 `json:"rate"`          // 税率百分比
	DisplayOrder int     `json:"display_order"` // 显示排序
	IsActive     bool    `json:"is_active"`     // 是否启用
	Description  string  `json:"description"`   // 税类别描述
	CreatedAt    int64   `json:"created_at"`    // 创建时间（Unix 时间戳）
	UpdatedAt    int64   `json:"updated_at"`    // 更新时间（Unix 时间戳）
}

// ListCategoriesResponse 税类别列表响应
type ListCategoriesResponse struct {
	Total int64              `json:"total"` // 总数
	Items []*CategoryResponse `json:"items"` // 税类别列表
}
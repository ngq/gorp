// Package request 定义 HTTP 请求结构体（目录服务相关：国家、省/州、货币）。
package request

// ==================== 国家请求 ====================

// ListCountryRequest 国家列表查询请求。
type ListCountryRequest struct {
	Page int `form:"page"`                                        // 页码（默认1）
	Size int `form:"size"`                                        // 每页数量（默认10）
}

// CreateCountry 创建国家请求。
type CreateCountry struct {
	Name             string `json:"name" binding:"required"`              // 国家名称
	IsoCode2         string `json:"iso_code2" binding:"required"`         // ISO 2字母代码
	IsoCode3         string `json:"iso_code3"`                            // ISO 3字母代码
	AddressFormat    string `json:"address_format"`                       // 地址格式
	PostcodeRequired bool   `json:"postcode_required"`                    // 是否需要邮编
}

// UpdateCountry 更新国家请求。
type UpdateCountry struct {
	Name             string `json:"name" binding:"required"`              // 国家名称
	IsoCode2         string `json:"iso_code2" binding:"required"`         // ISO 2字母代码
	IsoCode3         string `json:"iso_code3"`                            // ISO 3字母代码
	AddressFormat    string `json:"address_format"`                       // 地址格式
	PostcodeRequired bool   `json:"postcode_required"`                    // 是否需要邮编
}

// ==================== 省/州请求 ====================

// ListStateRequest 省/州列表查询请求。
type ListStateRequest struct {
	Page int `form:"page"`                                        // 页码（默认1）
	Size int `form:"size"`                                        // 每页数量（默认10）
}

// CreateState 创建省/州请求。
type CreateState struct {
	Name    string `json:"name" binding:"required"`                  // 省/州名称
	IsoCode string `json:"iso_code"`                                 // ISO 代码
}

// UpdateState 更新省/州请求。
type UpdateState struct {
	Name    string `json:"name" binding:"required"`                  // 省/州名称
	IsoCode string `json:"iso_code"`                                 // ISO 代码
}

// ==================== 货币请求 ====================

// ListCurrencyRequest 货币列表查询请求。
type ListCurrencyRequest struct {
	Page int `form:"page"`                                        // 页码（默认1）
	Size int `form:"size"`                                        // 每页数量（默认10）
}

// CreateCurrency 创建货币请求。
type CreateCurrency struct {
	Name     string  `json:"name" binding:"required"`               // 货币名称
	Code     string  `json:"code" binding:"required"`               // 货币代码（如 CNY、USD）
	Symbol   string  `json:"symbol"`                                // 货币符号（如 ¥、$）
	Rate     float64 `json:"rate"`                                  // 汇率（相对于基础货币）
	IsActive bool    `json:"is_active"`                             // 是否启用
}

// UpdateCurrency 更新货币请求。
type UpdateCurrency struct {
	Name     string  `json:"name" binding:"required"`               // 货币名称
	Code     string  `json:"code" binding:"required"`               // 货币代码
	Symbol   string  `json:"symbol"`                                // 货币符号
	Rate     float64 `json:"rate"`                                  // 汇率
	IsActive bool    `json:"is_active"`                             // 是否启用
}

// ApplyRatesRequest 应用汇率请求。
type ApplyRatesRequest struct {
	Rates []CurrencyRateRequest `json:"rates" binding:"required"` // 汇率列表
}

// CurrencyRateRequest 单条汇率项。
type CurrencyRateRequest struct {
	CurrencyID uint    `json:"currency_id" binding:"required"` // 货币ID
	Rate       float64 `json:"rate" binding:"required"`       // 汇率
}

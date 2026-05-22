package request

// ==================== 国家请求 ====================

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

// CreateState 创建省/州请求。
type CreateState struct {
	Name    string `json:"name" binding:"required"`    // 省/州名称
	IsoCode string `json:"iso_code"`                   // ISO 代码
}

// UpdateState 更新省/州请求。
type UpdateState struct {
	Name    string `json:"name" binding:"required"`    // 省/州名称
	IsoCode string `json:"iso_code"`                   // ISO 代码
}

// ==================== 货币请求 ====================

// CreateCurrency 创建货币请求。
type CreateCurrency struct {
	Name     string  `json:"name" binding:"required"`   // 货币名称
	Code     string  `json:"code" binding:"required"`   // 货币代码（如 CNY、USD）
	Symbol   string  `json:"symbol"`                    // 货币符号（如 ¥、$）
	Rate     float64 `json:"rate"`                      // 汇率（相对于基础货币）
	IsActive bool    `json:"is_active"`                 // 是否启用
}

// UpdateCurrency 更新货币请求。
type UpdateCurrency struct {
	Name     string  `json:"name" binding:"required"`   // 货币名称
	Code     string  `json:"code" binding:"required"`   // 货币代码
	Symbol   string  `json:"symbol"`                    // 货币符号
	Rate     float64 `json:"rate"`                      // 汇率
	IsActive bool    `json:"is_active"`                 // 是否启用
}

// ApplyRates 应用汇率请求。
type ApplyRates struct {
	Rates []CurrencyRate `json:"rates" binding:"required"` // 汇率列表
}

// CurrencyRate 单条汇率项。
type CurrencyRate struct {
	CurrencyID uint    `json:"currency_id" binding:"required"` // 货币ID
	Rate       float64 `json:"rate" binding:"required"`       // 汇率
}

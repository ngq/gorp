package response

// ==================== 国家响应 ====================

// Country 国家响应结构体。
type Country struct {
	ID               uint   `json:"id"`                // 国家ID
	Name             string `json:"name"`              // 国家名称
	IsoCode2         string `json:"iso_code2"`         // ISO 2字母代码
	IsoCode3         string `json:"iso_code3"`         // ISO 3字母代码
	AddressFormat    string `json:"address_format"`    // 地址格式
	PostcodeRequired bool   `json:"postcode_required"` // 是否需要邮编
	CreatedAt        string `json:"created_at"`        // 创建时间
	UpdatedAt        string `json:"updated_at"`        // 更新时间
}

// CountryList 国家列表响应。
type CountryList struct {
	Items []Country `json:"items"` // 国家列表
	Total int64     `json:"total"` // 总数
	Page  int       `json:"page"`  // 当前页
	Size  int       `json:"size"`  // 每页大小
}

// ==================== 省/州响应 ====================

// State 省/州响应结构体。
type State struct {
	ID        uint   `json:"id"`         // 省/州ID
	CountryID uint   `json:"country_id"` // 所属国家ID
	Name      string `json:"name"`       // 省/州名称
	IsoCode   string `json:"iso_code"`   // ISO 代码
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// StateList 省/州列表响应。
type StateList struct {
	Items []State `json:"items"` // 省/州列表
	Total int64   `json:"total"` // 总数
	Page  int     `json:"page"`  // 当前页
	Size  int     `json:"size"`  // 每页大小
}

// ==================== 货币响应 ====================

// Currency 货币响应结构体。
type Currency struct {
	ID        uint    `json:"id"`         // 货币ID
	Name      string  `json:"name"`       // 货币名称
	Code      string  `json:"code"`       // 货币代码
	Symbol    string  `json:"symbol"`     // 货币符号
	Rate      float64 `json:"rate"`       // 汇率
	IsActive  bool    `json:"is_active"`  // 是否启用
	CreatedAt string  `json:"created_at"` // 创建时间
	UpdatedAt string  `json:"updated_at"` // 更新时间
}

// CurrencyList 货币列表响应。
type CurrencyList struct {
	Items []Currency `json:"items"` // 货币列表
	Total int64      `json:"total"` // 总数
	Page  int        `json:"page"`  // 当前页
	Size  int        `json:"size"`  // 每页大小
}

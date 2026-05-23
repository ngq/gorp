// Package response 定义 HTTP 响应结构体（目录服务相关：国家、省/州、货币）。
package response

// ==================== 国家响应 ====================

// Country 国家响应结构体。
type Country struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	IsoCode2         string `json:"iso_code2"`
	IsoCode3         string `json:"iso_code3"`
	AddressFormat    string `json:"address_format"`
	PostcodeRequired bool   `json:"postcode_required"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CountryList 国家列表响应。
type CountryList struct {
	Items []Country `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

// ==================== 省/州响应 ====================

// State 省/州响应结构体。
type State struct {
	ID        uint   `json:"id"`
	CountryID uint   `json:"country_id"`
	Name      string `json:"name"`
	IsoCode   string `json:"iso_code"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// StateList 省/州列表响应。
type StateList struct {
	Items []State `json:"items"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
}

// ==================== 货币响应 ====================

// Currency 货币响应结构体。
type Currency struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	Symbol    string  `json:"symbol"`
	Rate      float64 `json:"rate"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CurrencyList 货币列表响应。
type CurrencyList struct {
	Items []Currency `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}
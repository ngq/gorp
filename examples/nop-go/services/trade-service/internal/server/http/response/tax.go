// Package response 包含交易服务 HTTP 响应结构体定义
// tax.go 定义税务相关响应结构体
// 注意：原 Provider/Category 响应类型已重命名为 TaxProvider/TaxCategory 类型
package response

import "time"

// TaxProviderResponse 税务服务商响应（原 ProviderResponse，重命名避免冲突）
type TaxProviderResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TaxCategoryResponse 税种分类响应（原 CategoryResponse，重命名避免冲突）
type TaxCategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Rate        float64   `json:"rate"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TaxRateResponse 税率响应
type TaxRateResponse struct {
	ID            string     `json:"id"`
	TaxCategoryID string     `json:"taxCategoryId"`
	Region        string     `json:"region"`
	Rate          float64    `json:"rate"`
	IsActive      bool       `json:"isActive"`
	EffectiveFrom time.Time  `json:"effectiveFrom"`
	EffectiveTo   *time.Time `json:"effectiveTo"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// TaxTransactionResponse 税务交易记录响应
type TaxTransactionResponse struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"orderId"`
	TaxCategoryID string    `json:"taxCategoryId"`
	TaxRateID     string    `json:"taxRateId"`
	TaxableAmount float64   `json:"taxableAmount"`
	TaxAmount     float64   `json:"taxAmount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
}

// TaxTransactionListResponse 税务交易记录列表响应
type TaxTransactionListResponse struct {
	Items []TaxTransactionResponse `json:"items"`
	Total int64                    `json:"total"`
}

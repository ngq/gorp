// Package request 包含交易服务 HTTP 请求结构体定义
// tax.go 定义税务相关请求结构体
// 注意：原 Provider/Category 请求类型已重命名为 TaxProvider/TaxCategory 类型
package request

// CreateTaxProviderRequest 创建税务服务商请求
type CreateTaxProviderRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// UpdateTaxProviderRequest 更新税务服务商请求
type UpdateTaxProviderRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// CreateTaxCategoryRequest 创建税种分类请求
type CreateTaxCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate" binding:"required"`
	IsActive    bool    `json:"isActive"`
}

// UpdateTaxCategoryRequest 更新税种分类请求
type UpdateTaxCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate"`
	IsActive    bool    `json:"isActive"`
}

// CreateTaxRateRequest 创建税率请求
type CreateTaxRateRequest struct {
	TaxCategoryID string  `json:"taxCategoryId" binding:"required"`
	Region        string  `json:"region" binding:"required"`
	Rate          float64 `json:"rate" binding:"required"`
	IsActive      bool    `json:"isActive"`
}

// UpdateTaxRateRequest 更新税率请求
type UpdateTaxRateRequest struct {
	Rate     float64 `json:"rate"`
	IsActive bool    `json:"isActive"`
}
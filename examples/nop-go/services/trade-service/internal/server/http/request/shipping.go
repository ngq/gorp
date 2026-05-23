// Package request 包含交易服务 HTTP 请求结构体定义
// shipping.go 定义物流相关请求结构体
// 注意：原 Provider 请求类型已重命名为 ShippingProvider 类型
package request

// CreateShippingProviderRequest 创建物流服务商请求
type CreateShippingProviderRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// UpdateShippingProviderRequest 更新物流服务商请求
type UpdateShippingProviderRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// CreateShippingOrderRequest 创建物流订单请求
type CreateShippingOrderRequest struct {
	OrderID            string `json:"orderId" binding:"required"`
	ShippingProviderID string `json:"shippingProviderId" binding:"required"`
	ShippingAddress    string `json:"shippingAddress"`
}

// UpdateShippingOrderRequest 更新物流订单请求
type UpdateShippingOrderRequest struct {
	Status         string `json:"status"`
	TrackingNumber string `json:"trackingNumber"`
}

// CreateShippingEventRequest 创建物流事件请求
type CreateShippingEventRequest struct {
	ShippingOrderID string `json:"shippingOrderId" binding:"required"`
	Status          string `json:"status" binding:"required"`
	Location        string `json:"location"`
	Description     string `json:"description"`
}

// CreateShippingRateRequest 创建运费率请求
type CreateShippingRateRequest struct {
	ShippingProviderID string  `json:"shippingProviderId" binding:"required"`
	OriginZone         string  `json:"originZone" binding:"required"`
	DestinationZone    string  `json:"destinationZone" binding:"required"`
	WeightMin          float64 `json:"weightMin"`
	WeightMax          float64 `json:"weightMax"`
	Rate               float64 `json:"rate" binding:"required"`
	Currency           string  `json:"currency"`
	EstimatedDays      int     `json:"estimatedDays"`
}

// UpdateShippingRateRequest 更新运费率请求
type UpdateShippingRateRequest struct {
	Rate          float64 `json:"rate"`
	EstimatedDays int     `json:"estimatedDays"`
}
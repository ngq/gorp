// Package response 包含交易服务 HTTP 响应结构体定义
// shipping.go 定义物流相关响应结构体
// 注意：原 Provider 响应类型已重命名为 ShippingProvider 类型
package response

import "time"

// ShippingProviderResponse 物流服务商响应（原 ProviderResponse，重命名避免冲突）
type ShippingProviderResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ShippingOrderResponse 物流订单响应
type ShippingOrderResponse struct {
	ID                 string     `json:"id"`
	OrderID            string     `json:"orderId"`
	ShippingProviderID string     `json:"shippingProviderId"`
	TrackingNumber     string     `json:"trackingNumber"`
	Status             string     `json:"status"`
	ShippingAddress    string     `json:"shippingAddress"`
	EstimatedDelivery  *time.Time `json:"estimatedDelivery"`
	ActualDelivery     *time.Time `json:"actualDelivery"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// ShippingEventResponse 物流事件响应
type ShippingEventResponse struct {
	ID              string    `json:"id"`
	ShippingOrderID string    `json:"shippingOrderId"`
	Status          string    `json:"status"`
	Location        string    `json:"location"`
	Description     string    `json:"description"`
	EventTime       time.Time `json:"eventTime"`
	CreatedAt       time.Time `json:"createdAt"`
}

// ShippingRateResponse 运费率响应
type ShippingRateResponse struct {
	ID                 string    `json:"id"`
	ShippingProviderID string    `json:"shippingProviderId"`
	OriginZone         string    `json:"originZone"`
	DestinationZone    string    `json:"destinationZone"`
	WeightMin          float64   `json:"weightMin"`
	WeightMax          float64   `json:"weightMax"`
	Rate               float64   `json:"rate"`
	Currency           string    `json:"currency"`
	EstimatedDays      int       `json:"estimatedDays"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

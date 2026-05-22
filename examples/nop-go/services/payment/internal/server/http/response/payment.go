// Package response 提供 payment 服务的 HTTP 响应结构体定义
package response

// PaymentMethodResponse 支付方式响应
// 对应 nopCommerce Admin/Payment/Methods 返回数据
type PaymentMethodResponse struct {
	ID                    uint   `json:"id"`                        // 支付方式 ID
	Name                  string `json:"name"`                      // 支付方式名称
	SystemKeyword         string `json:"system_keyword"`            // 系统关键字标识
	DisplayOrder          int    `json:"display_order"`             // 显示排序
	IsActive              bool   `json:"is_active"`                 // 是否启用
	LogoURL               string `json:"logo_url"`                  // Logo 地址
	SupportsRefund        bool   `json:"supports_refund"`           // 是否支持退款
	SupportsPartialRefund bool   `json:"supports_partial_refund"`   // 是否支持部分退款
	CreatedAt             int64  `json:"created_at"`                // 创建时间（Unix 时间戳）
	UpdatedAt             int64  `json:"updated_at"`                // 更新时间（Unix 时间戳）
}

// MethodRestrictionResponse 支付方式限制响应
// 对应 nopCommerce Admin/Payment/MethodRestrictions 返回数据
type MethodRestrictionResponse struct {
	ID               uint    `json:"id"`                // 限制规则 ID
	PaymentMethodID  uint    `json:"payment_method_id"` // 关联的支付方式 ID
	MinOrderAmount   float64 `json:"min_order_amount"`  // 最小订单金额
	MaxOrderAmount   float64 `json:"max_order_amount"`  // 最大订单金额
	RestrictionType  string  `json:"restriction_type"`  // 限制类型
	RestrictionValue string  `json:"restriction_value"` // 限制值
	IsActive         bool    `json:"is_active"`         // 是否启用
	CreatedAt        int64   `json:"created_at"`        // 创建时间（Unix 时间戳）
	UpdatedAt        int64   `json:"updated_at"`        // 更新时间（Unix 时间戳）
}

// ListPaymentMethodsResponse 支付方式列表响应
type ListPaymentMethodsResponse struct {
	Total int64                   `json:"total"` // 总数
	Items []*PaymentMethodResponse `json:"items"` // 支付方式列表
}

// ListMethodRestrictionsResponse 支付方式限制列表响应
type ListMethodRestrictionsResponse struct {
	Total int64                      `json:"total"` // 总数
	Items []*MethodRestrictionResponse `json:"items"` // 限制列表
}

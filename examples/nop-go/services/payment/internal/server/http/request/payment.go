// Package request 提供 payment 服务的 HTTP 请求结构体定义
package request

// UpdatePaymentMethodRequest 更新支付方式请求
// 对应路由: PUT /api/v1/payment/methods/:id
type UpdatePaymentMethodRequest struct {
	ID                    uint   `json:"id"`                        // 支付方式 ID（路径参数传入）
	Name                  string `json:"name"`                      // 支付方式名称
	SystemKeyword         string `json:"system_keyword"`            // 系统关键字标识（如 "Payments.Manual"）
	DisplayOrder          int    `json:"display_order"`             // 显示排序
	IsActive              bool   `json:"is_active"`                 // 是否启用
	LogoURL               string `json:"logo_url"`                  // 支付方式 Logo 地址
	SupportsRefund        bool   `json:"supports_refund"`           // 是否支持退款
	SupportsPartialRefund bool   `json:"supports_partial_refund"`   // 是否支持部分退款
}

// UpdateMethodRestrictionsRequest 更新支付方式限制请求
// 对应路由: PUT /api/v1/payment/method-restrictions
type UpdateMethodRestrictionsRequest struct {
	ID               uint    `json:"id"`                  // 限制规则 ID
	PaymentMethodID  uint    `json:"payment_method_id"`   // 关联的支付方式 ID
	MinOrderAmount   float64 `json:"min_order_amount"`    // 最小订单金额
	MaxOrderAmount   float64 `json:"max_order_amount"`    // 最大订单金额
	RestrictionType  string  `json:"restriction_type"`    // 限制类型（amount_range/country/currency 等）
	RestrictionValue string  `json:"restriction_value"`   // 限制值
	IsActive         bool    `json:"is_active"`           // 是否启用
}

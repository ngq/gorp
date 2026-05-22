// Package request 提供 discount 服务的 HTTP 请求结构体定义
package request

// ==================== 折扣 CRUD ====================

// CreateDiscountRequest 创建折扣请求
// 对应路由: POST /api/v1/discounts
type CreateDiscountRequest struct {
	Name           string  `json:"name"`             // 折扣名称（如"夏季大促"）
	DiscountType   string  `json:"discount_type"`    // 折扣类型（percentage/fixed/free_shipping）
	DiscountAmount float64 `json:"discount_amount"`  // 折扣金额/百分比
	StartDate      string  `json:"start_date"`       // 折扣开始日期（YYYY-MM-DD）
	EndDate        string  `json:"end_date"`         // 折扣结束日期（YYYY-MM-DD）
	RequiresCouponCode bool `json:"requires_coupon_code"` // 是否需要优惠券码
	CouponCode     string  `json:"coupon_code"`      // 优惠券码
	IsCumulative   bool    `json:"is_cumulative"`    // 是否可叠加使用
	DisplayOrder   int     `json:"display_order"`    // 显示排序
	IsActive       bool    `json:"is_active"`        // 是否启用
	LimitationTimes int    `json:"limitation_times"` // 使用次数限制（0 表示不限）
}

// UpdateDiscountRequest 更新折扣请求
// 对应路由: PUT /api/v1/discounts/:id
type UpdateDiscountRequest struct {
	ID                uint    `json:"id"`                     // 折扣 ID（路径参数传入）
	Name              string  `json:"name"`                   // 折扣名称
	DiscountType      string  `json:"discount_type"`          // 折扣类型
	DiscountAmount    float64 `json:"discount_amount"`        // 折扣金额/百分比
	StartDate         string  `json:"start_date"`             // 折扣开始日期
	EndDate           string  `json:"end_date"`               // 折扣结束日期
	RequiresCouponCode bool   `json:"requires_coupon_code"`   // 是否需要优惠券码
	CouponCode        string  `json:"coupon_code"`            // 优惠券码
	IsCumulative      bool    `json:"is_cumulative"`          // 是否可叠加使用
	DisplayOrder      int     `json:"display_order"`          // 显示排序
	IsActive          bool    `json:"is_active"`              // 是否启用
	LimitationTimes   int     `json:"limitation_times"`       // 使用次数限制
}
// Package request 优惠模块请求结构 —— 优惠与使用记录的HTTP请求DTO
package request

// CreateDiscountRequest 创建优惠请求
type CreateDiscountRequest struct {
	Name         string  `json:"name" binding:"required"`        // 优惠名称
	Code         string  `json:"code" binding:"required"`        // 优惠编码
	Type         int     `json:"type" binding:"required"`        // 优惠类型：1=满减 2=折扣 3=固定金额
	Value        float64 `json:"value" binding:"required"`       // 优惠值
	MinAmount    float64 `json:"min_amount"`                      // 最低消费金额
	MaxDiscount  float64 `json:"max_discount"`                    // 最大优惠金额
	StartTime    string  `json:"start_time"`                      // 开始时间
	EndTime      string  `json:"end_time"`                        // 结束时间
	TotalQuota   int     `json:"total_quota"`                     // 总发放量
	PerUserLimit int     `json:"per_user_limit"`                  // 每人限领数量
	Status       int     `json:"status"`                          // 状态
	Description  string  `json:"description"`                     // 描述
}

// UpdateDiscountRequest 更新优惠请求
type UpdateDiscountRequest struct {
	Name         string  `json:"name"`                            // 优惠名称
	Type         int     `json:"type"`                            // 优惠类型
	Value        float64 `json:"value"`                           // 优惠值
	MinAmount    float64 `json:"min_amount"`                      // 最低消费金额
	MaxDiscount  float64 `json:"max_discount"`                    // 最大优惠金额
	StartTime    string  `json:"start_time"`                      // 开始时间
	EndTime      string  `json:"end_time"`                        // 结束时间
	TotalQuota   int     `json:"total_quota"`                     // 总发放量
	PerUserLimit int     `json:"per_user_limit"`                  // 每人限领数量
	Status       int     `json:"status"`                          // 状态
	Description  string  `json:"description"`                     // 描述
}

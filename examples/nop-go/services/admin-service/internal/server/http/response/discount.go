// Package response 优惠模块响应结构 —— 优惠与使用记录的HTTP响应DTO
//
// 重要：本文件负责将 biz.Discount / biz.DiscountUsage 领域实体转换为响应 DTO。
// 这是从原 discount 独立服务合并而来，原 biz 层直接返回 response DTO（已修正）。
package response

import (
	"nop-go/services/admin-service/internal/biz"
)

// DiscountResponse 优惠响应
type DiscountResponse struct {
	ID           uint    `json:"id"`            // 优惠ID
	Name         string  `json:"name"`          // 优惠名称
	Code         string  `json:"code"`          // 优惠编码
	Type         int     `json:"type"`          // 优惠类型
	Value        float64 `json:"value"`         // 优惠值
	MinAmount    float64 `json:"min_amount"`    // 最低消费
	MaxDiscount  float64 `json:"max_discount"`  // 最大优惠
	StartTime    string  `json:"start_time"`    // 开始时间
	EndTime      string  `json:"end_time"`      // 结束时间
	TotalQuota   int     `json:"total_quota"`   // 总发放量
	UsedQuota    int     `json:"used_quota"`    // 已使用量
	PerUserLimit int     `json:"per_user_limit"`// 每人限领
	Status       int     `json:"status"`        // 状态
	Description  string  `json:"description"`   // 描述
	CreatedAt    string  `json:"created_at"`    // 创建时间
	UpdatedAt    string  `json:"updated_at"`    // 更新时间
}

// NewDiscountResponse 将领域实体转换为响应 DTO
func NewDiscountResponse(d *biz.Discount) *DiscountResponse {
	return &DiscountResponse{
		ID:           d.ID,
		Name:         d.Name,
		Code:         d.Code,
		Type:         d.Type,
		Value:        d.Value,
		MinAmount:    d.MinAmount,
		MaxDiscount:  d.MaxDiscount,
		StartTime:    d.StartTime,
		EndTime:      d.EndTime,
		TotalQuota:   d.TotalQuota,
		UsedQuota:    d.UsedQuota,
		PerUserLimit: d.PerUserLimit,
		Status:       d.Status,
		Description:  d.Description,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

// NewDiscountResponseList 将领域实体列表转换为响应 DTO 列表
func NewDiscountResponseList(discounts []*biz.Discount) []*DiscountResponse {
	items := make([]*DiscountResponse, 0, len(discounts))
	for _, d := range discounts {
		items = append(items, NewDiscountResponse(d))
	}
	return items
}

// DiscountUsageResponse 优惠使用记录响应
type DiscountUsageResponse struct {
	ID         uint   `json:"id"`          // 使用记录ID
	DiscountID uint   `json:"discount_id"` // 关联优惠ID
	UserID     uint   `json:"user_id"`     // 使用用户ID
	OrderNo    string `json:"order_no"`    // 关联订单号
	UsedAt     string `json:"used_at"`     // 使用时间
	Status     int    `json:"status"`      // 状态
	CreatedAt  string `json:"created_at"`  // 创建时间
}

// NewDiscountUsageResponse 将领域实体转换为响应 DTO
func NewDiscountUsageResponse(u *biz.DiscountUsage) *DiscountUsageResponse {
	return &DiscountUsageResponse{
		ID:         u.ID,
		DiscountID: u.DiscountID,
		UserID:     u.UserID,
		OrderNo:    u.OrderNo,
		UsedAt:     u.UsedAt,
		Status:     u.Status,
		CreatedAt:  u.CreatedAt,
	}
}

// NewDiscountUsageResponseList 将领域实体列表转换为响应 DTO 列表
func NewDiscountUsageResponseList(usages []*biz.DiscountUsage) []*DiscountUsageResponse {
	items := make([]*DiscountUsageResponse, 0, len(usages))
	for _, u := range usages {
		items = append(items, NewDiscountUsageResponse(u))
	}
	return items
}
package response

import "nop-go/services/content-service/internal/service"

// ==================== 推广合作方响应 ====================

// AffiliateList 推广合作方列表响应
type AffiliateList struct {
	Items []*service.AffiliateResponse `json:"items"`
	Total int64                        `json:"total"`
	Page  int                          `json:"page"`
	Size  int                          `json:"size"`
}

// ==================== 推广订单响应 ====================

// AffiliateOrderList 推广订单列表响应
type AffiliateOrderList struct {
	Items []*service.AffiliateOrderResponse `json:"items"`
	Total int64                             `json:"total"`
	Page  int                               `json:"page"`
	Size  int                               `json:"size"`
}

// ==================== 推广客户响应 ====================

// AffiliateCustomerList 推广客户列表响应
type AffiliateCustomerList struct {
	Items []*service.AffiliateCustomerResponse `json:"items"`
	Total int64                                `json:"total"`
	Page  int                                  `json:"page"`
	Size  int                                  `json:"size"`
}
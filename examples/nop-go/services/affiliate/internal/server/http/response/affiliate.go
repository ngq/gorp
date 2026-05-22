package response

import "nop-go/services/affiliate/internal/service"

// AffiliateList 联盟列表响应。
// 直接使用 service 层的 AffiliateResponse，避免重复定义与类型转换。
type AffiliateList struct {
	Items []service.AffiliateResponse `json:"items"` // 联盟列表
	Total int64                       `json:"total"` // 总数
	Page  int                         `json:"page"`  // 当前页
	Size  int                         `json:"size"`  // 每页大小
}

// AffiliateOrderList 联盟关联订单列表响应。
// 直接使用 service 层的 AffiliateOrderResponse，避免重复定义与类型转换。
type AffiliateOrderList struct {
	Items []service.AffiliateOrderResponse `json:"items"` // 订单列表
	Total int64                            `json:"total"` // 总数
	Page  int                              `json:"page"`  // 当前页
	Size  int                              `json:"size"`  // 每页大小
}

// AffiliateCustomerList 联盟关联客户列表响应。
// 直接使用 service 层的 AffiliateCustomerResponse，避免重复定义与类型转换。
type AffiliateCustomerList struct {
	Items []service.AffiliateCustomerResponse `json:"items"` // 客户列表
	Total int64                               `json:"total"` // 总数
	Page  int                                 `json:"page"`  // 当前页
	Size  int                                 `json:"size"`  // 每页大小
}
package response

import "nop-go/services/content/internal/service"

// BlogList 博客列表响应。
// 直接使用 service 层的 BlogResponse，避免重复定义与类型转换。
type BlogList struct {
	Items []service.BlogResponse `json:"items"` // 博客列表
	Total int64                  `json:"total"` // 总数
	Page  int                    `json:"page"`  // 当前页
	Size  int                    `json:"size"`  // 每页大小
}

// NewsList 新闻列表响应。
// 直接使用 service 层的 NewsResponse，避免重复定义与类型转换。
type NewsList struct {
	Items []service.NewsResponse `json:"items"` // 新闻列表
	Total int64                  `json:"total"` // 总数
	Page  int                    `json:"page"`  // 当前页
	Size  int                    `json:"size"`  // 每页大小
}

// TopicList 页面列表响应。
// 直接使用 service 层的 TopicResponse，避免重复定义与类型转换。
type TopicList struct {
	Items []service.TopicResponse `json:"items"` // 页面列表
	Total int64                   `json:"total"` // 总数
	Page  int                     `json:"page"`  // 当前页
	Size  int                     `json:"size"`  // 每页大小
}

// PollList 投票列表响应。
// 直接使用 service 层的 PollResponse，避免重复定义与类型转换。
type PollList struct {
	Items []service.PollResponse `json:"items"` // 投票列表
	Total int64                  `json:"total"` // 总数
	Page  int                    `json:"page"`  // 当前页
	Size  int                    `json:"size"`  // 每页大小
}

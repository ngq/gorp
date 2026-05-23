package response

import "nop-go/services/content-service/internal/service"

// ==================== 博客响应 ====================

// BlogList 博客列表响应
type BlogList struct {
	Items []*service.BlogResponse `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

// ==================== 新闻响应 ====================

// NewsList 新闻列表响应
type NewsList struct {
	Items []*service.NewsResponse `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

// ==================== 话题响应 ====================

// TopicList 话题列表响应
type TopicList struct {
	Items []*service.TopicResponse `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}

// ==================== 投票响应 ====================

// PollList 投票列表响应
type PollList struct {
	Items []*service.PollResponse `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}
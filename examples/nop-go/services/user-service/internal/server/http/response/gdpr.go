// Package response 定义 GDPR 相关的 HTTP 响应结构体。
// 用于统一响应格式。
package response

import "time"

// GdprResponse GDPR 请求响应
type GdprResponse struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	RequestType string     `json:"request_type"`
	Status      string     `json:"status"`
	Reason      string     `json:"reason"`
	ReviewedBy  *uint      `json:"reviewed_by"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GdprListResponse GDPR 请求列表响应（带分页信息）
type GdprListResponse struct {
	Data   []*GdprResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}
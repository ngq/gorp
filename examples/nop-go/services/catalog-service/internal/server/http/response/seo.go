// Package response 定义 HTTP 响应结构体（SEO 相关）。
package response

// Seo SEO 元数据响应结构体。
type Seo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// SeoList SEO 元数据列表响应。
type SeoList struct {
	Items []Seo  `json:"items"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}
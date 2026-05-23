// Package request 定义 HTTP 请求结构体（SEO 相关）。
package request

// ListSeoRequest SEO 元数据列表查询请求。
type ListSeoRequest struct {
	Page int `form:"page"`                                            // 页码（默认1）
	Size int `form:"size"`                                            // 每页数量（默认10）
}

// CreateSeo 创建 SEO 元数据请求。
type CreateSeo struct {
	Username string `json:"username" binding:"required"`             // 用户名
	Email    string `json:"email" binding:"required,email"`          // 邮箱
}

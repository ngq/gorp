// Package request HTTP 请求结构体定义。
//
// 定义消息服务各接口的请求参数结构体，
// 包含 binding 标签用于 gin 框架的参数校验。
package request

// CreateMessageTemplate 创建消息模板请求。
type CreateMessageTemplate struct {
	Name         string `json:"name" binding:"required"`          // 模板名称
	Subject      string `json:"subject" binding:"required"`       // 邮件主题
	Body         string `json:"body" binding:"required"`          // 邮件正文
	EmailAccount string `json:"email_account" binding:"required"` // 发件邮箱账号
	IsActive     bool   `json:"is_active"`                        // 是否启用
}

// UpdateMessageTemplate 更新消息模板请求。
type UpdateMessageTemplate struct {
	Name         string `json:"name" binding:"required"`          // 模板名称
	Subject      string `json:"subject" binding:"required"`       // 邮件主题
	Body         string `json:"body" binding:"required"`          // 邮件正文
	EmailAccount string `json:"email_account" binding:"required"` // 发件邮箱账号
	IsActive     bool   `json:"is_active"`                        // 是否启用
}

// TestMessageTemplate 测试消息模板请求。
type TestMessageTemplate struct {
	ToEmail string `json:"to_email" binding:"required,email"` // 收件邮箱
}

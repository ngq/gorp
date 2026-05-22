package response

import "time"

// MessageTemplate 消息模板响应结构体。
type MessageTemplate struct {
	ID           uint      `json:"id"`            // 模板ID
	Name         string    `json:"name"`          // 模板名称
	Subject      string    `json:"subject"`       // 邮件主题
	Body         string    `json:"body"`          // 邮件正文
	EmailAccount string    `json:"email_account"` // 发件邮箱账号
	IsActive     bool      `json:"is_active"`     // 是否启用
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// MessageTemplateList 消息模板列表响应。
type MessageTemplateList struct {
	Items []MessageTemplate `json:"items"` // 模板列表
	Total int64             `json:"total"` // 总数
	Page  int               `json:"page"`  // 当前页
	Size  int               `json:"size"`  // 每页大小
}

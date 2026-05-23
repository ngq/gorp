// Package response HTTP 响应结构体定义。
//
// 定义消息服务各接口的响应数据结构，
// 与 service 层的 DTO 分离，handler 层负责两者之间的转换。
package response

import "time"

// MessageTemplate 消息模板响应结构体。
//
// 时间字段使用 time.Time 类型，JSON 序列化时由框架统一处理。
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
//
// 包含分页信息，供前端分页展示使用。
type MessageTemplateList struct {
	Items []MessageTemplate `json:"items"` // 模板列表
	Total int64             `json:"total"` // 总数
	Page  int               `json:"page"`  // 当前页
	Size  int               `json:"size"`  // 每页大小
}

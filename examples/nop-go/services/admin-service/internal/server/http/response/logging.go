// Package response 日志模块响应结构 —— 活动日志与系统日志的HTTP响应DTO
package response

// ActivityLogResponse 活动日志响应
type ActivityLogResponse struct {
	ID        uint   `json:"id"`          // 日志ID
	UserID    uint   `json:"user_id"`    // 操作用户ID
	UserName  string `json:"user_name"`  // 用户名
	Action    string `json:"action"`     // 操作类型
	Resource  string `json:"resource"`   // 操作资源
	IP        string `json:"ip"`          // IP地址
	UserAgent string `json:"user_agent"`  // 用户代理
	Detail    string `json:"detail"`     // 操作详情
	Status    int    `json:"status"`     // 操作结果
	CreatedAt string `json:"created_at"` // 创建时间
}

// SystemLogResponse 系统日志响应
type SystemLogResponse struct {
	ID        uint   `json:"id"`          // 日志ID
	Level     string `json:"level"`      // 日志级别
	Module    string `json:"module"`     // 所属模块
	Message   string `json:"message"`    // 日志消息
	Stack     string `json:"stack"`      // 堆栈信息
	Hostname  string `json:"hostname"`   // 主机名
	CreatedAt string `json:"created_at"` // 创建时间
}
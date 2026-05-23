// Package request 日志模块请求结构 —— 活动日志与系统日志的HTTP请求DTO
package request

// CreateActivityLogRequest 创建活动日志请求
type CreateActivityLogRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`     // 操作用户ID
	UserName  string `json:"user_name"`                       // 用户名
	Action    string `json:"action" binding:"required"`       // 操作类型
	Resource  string `json:"resource"`                        // 操作资源
	IP        string `json:"ip"`                              // IP地址
	UserAgent string `json:"user_agent"`                      // 用户代理
	Detail    string `json:"detail"`                          // 操作详情
	Status    int    `json:"status"`                          // 操作结果
}

// CreateSystemLogRequest 创建系统日志请求
type CreateSystemLogRequest struct {
	Level    string `json:"level" binding:"required"`        // 日志级别
	Module   string `json:"module" binding:"required"`       // 所属模块
	Message  string `json:"message" binding:"required"`      // 日志消息
	Stack    string `json:"stack"`                            // 堆栈信息
	Hostname string `json:"hostname"`                         // 主机名
}
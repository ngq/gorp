// Package response 安全模块响应结构 —— 权限与ACL的HTTP响应DTO
package response

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          uint   `json:"id"`           // 权限ID
	Name        string `json:"name"`         // 权限名称
	Code        string `json:"code"`         // 权限编码
	Description string `json:"description"`  // 权限描述
	Module      string `json:"module"`       // 所属模块
	Action      string `json:"action"`       // 动作类型
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
}

// ACLResponse ACL响应
type ACLResponse struct {
	ID        uint   `json:"id"`          // ACL规则ID
	RoleID    uint   `json:"role_id"`     // 角色ID
	Resource  string `json:"resource"`   // 资源标识
	Action    string `json:"action"`      // 允许的动作
	Effect    string `json:"effect"`      // 效果：allow/deny
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
}
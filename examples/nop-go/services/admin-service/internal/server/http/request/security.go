// Package request 安全模块请求结构 —— 权限与ACL的HTTP请求DTO
package request

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`        // 权限名称
	Code        string `json:"code" binding:"required"`        // 权限编码
	Description string `json:"description"`                     // 权限描述
	Module      string `json:"module" binding:"required"`       // 所属模块
	Action      string `json:"action" binding:"required"`       // 动作类型
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name"`                            // 权限名称
	Code        string `json:"code"`                            // 权限编码
	Description string `json:"description"`                     // 权限描述
	Module      string `json:"module"`                          // 所属模块
	Action      string `json:"action"`                          // 动作类型
}

// CreateACLRequest 创建ACL规则请求
type CreateACLRequest struct {
	RoleID   uint   `json:"role_id" binding:"required"`      // 角色ID
	Resource string `json:"resource" binding:"required"`     // 资源标识
	Action   string `json:"action" binding:"required"`       // 允许的动作
	Effect   string `json:"effect" binding:"required"`       // 效果：allow/deny
}

// UpdateACLRequest 更新ACL规则请求
type UpdateACLRequest struct {
	RoleID   uint   `json:"role_id"`                          // 角色ID
	Resource string `json:"resource"`                         // 资源标识
	Action   string `json:"action"`                           // 允许的动作
	Effect   string `json:"effect"`                           // 效果
}
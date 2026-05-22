package request

// CreatePermission 创建权限请求。
type CreatePermission struct {
	Name         string `json:"name" binding:"required"`           // 权限名称
	SystemName   string `json:"system_name" binding:"required"`    // 权限系统名称（唯一标识）
	Category     string `json:"category" binding:"required"`      // 权限分类
	DisplayOrder int    `json:"display_order"`                    // 显示排序
}

// UpdatePermission 更新权限请求。
type UpdatePermission struct {
	Name         string `json:"name" binding:"required"`           // 权限名称
	SystemName   string `json:"system_name" binding:"required"`    // 权限系统名称
	Category     string `json:"category" binding:"required"`       // 权限分类
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// CreateACLRecord 创建ACL记录请求。
type CreateACLRecord struct {
	UserID       uint `json:"user_id" binding:"required"`        // 用户ID
	PermissionID uint `json:"permission_id" binding:"required"`  // 权限ID
}
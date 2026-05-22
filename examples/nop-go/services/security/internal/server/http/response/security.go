package response

import "time"

// Permission 权限响应结构体。
type Permission struct {
	ID           uint      `json:"id"`            // 权限ID
	Name         string    `json:"name"`          // 权限名称
	SystemName   string    `json:"system_name"`   // 权限系统名称
	Category     string    `json:"category"`      // 权限分类
	DisplayOrder int       `json:"display_order"` // 显示排序
	CreatedAt    time.Time `json:"created_at"`   // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`   // 更新时间
}

// PermissionList 权限列表响应。
type PermissionList struct {
	Items []Permission `json:"items"` // 权限列表
	Total int64       `json:"total"` // 总数
	Page  int         `json:"page"`  // 当前页
	Size  int         `json:"size"`  // 每页大小
}

// ACLRecord ACL记录响应结构体。
type ACLRecord struct {
	ID             uint   `json:"id"`              // ACL记录ID
	UserID         uint   `json:"user_id"`         // 用户ID
	PermissionID   uint   `json:"permission_id"`   // 权限ID
	PermissionName string `json:"permission_name"` // 权限名称
}

// ACLRecordList ACL记录列表响应。
type ACLRecordList struct {
	Items []ACLRecord `json:"items"` // ACL记录列表
	Total int64      `json:"total"` // 总数
	Page  int        `json:"page"`  // 当前页
	Size  int        `json:"size"`  // 每页大小
}
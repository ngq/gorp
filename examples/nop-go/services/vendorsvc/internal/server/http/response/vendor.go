package response

import "time"

// Vendor 供应商响应结构体。
type Vendor struct {
	ID           uint      `json:"id"`            // 供应商ID
	Name         string    `json:"name"`          // 供应商名称
	Email        string    `json:"email"`         // 供应商邮箱
	Active       bool      `json:"active"`        // 是否启用
	DisplayOrder int       `json:"display_order"` // 显示排序
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// VendorList 供应商列表响应。
type VendorList struct {
	Items []Vendor `json:"items"` // 供应商列表
	Total int64   `json:"total"` // 总数
	Page  int     `json:"page"`  // 当前页
	Size  int     `json:"size"`  // 每页大小
}

// VendorApply 供应商申请响应结构体。
type VendorApply struct {
	ID          uint   `json:"id"`          // 申请ID
	Name        string `json:"name"`        // 供应商名称
	Email       string `json:"email"`       // 供应商邮箱
	Description string `json:"description"` // 申请描述
	Status      string `json:"status"`      // 申请状态：pending/approved/rejected
	CreatedAt   string `json:"created_at"`  // 创建时间
}
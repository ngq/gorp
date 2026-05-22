package response

import "time"

// Store 店铺响应结构体。
type Store struct {
	ID           uint      `json:"id"`            // 店铺ID
	Name         string    `json:"name"`          // 店铺名称
	Url          string    `json:"url"`           // 店铺URL
	SslEnabled   bool      `json:"ssl_enabled"`   // 是否启用SSL
	Hosts        string    `json:"hosts"`         // 绑定主机列表
	DisplayOrder int       `json:"display_order"` // 显示排序
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// StoreList 店铺列表响应。
type StoreList struct {
	Items []Store `json:"items"` // 店铺列表
	Total int64   `json:"total"` // 总数
	Page  int     `json:"page"`  // 当前页
	Size  int     `json:"size"`  // 每页大小
}
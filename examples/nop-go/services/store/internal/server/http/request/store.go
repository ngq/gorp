package request

// CreateStore 创建店铺请求。
type CreateStore struct {
	Name         string `json:"name" binding:"required"`           // 店铺名称
	Url          string `json:"url" binding:"required"`            // 店铺URL
	SslEnabled   bool   `json:"ssl_enabled"`                       // 是否启用SSL
	Hosts        string `json:"hosts"`                             // 绑定主机列表（逗号分隔）
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// UpdateStore 更新店铺请求。
type UpdateStore struct {
	Name         string `json:"name" binding:"required"`           // 店铺名称
	Url          string `json:"url" binding:"required"`            // 店铺URL
	SslEnabled   bool   `json:"ssl_enabled"`                       // 是否启用SSL
	Hosts        string `json:"hosts"`                             // 绑定主机列表（逗号分隔）
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}
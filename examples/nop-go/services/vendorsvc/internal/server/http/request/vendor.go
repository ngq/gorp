package request

// CreateVendor 创建供应商请求。
type CreateVendor struct {
	Name         string `json:"name" binding:"required"`           // 供应商名称
	Email        string `json:"email" binding:"required,email"`    // 供应商邮箱
	Description  string `json:"description"`                       // 供应商描述
	Active       bool   `json:"active"`                            // 是否启用
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// UpdateVendor 更新供应商请求。
type UpdateVendor struct {
	Name         string `json:"name" binding:"required"`           // 供应商名称
	Email        string `json:"email" binding:"required,email"`    // 供应商邮箱
	Description  string `json:"description"`                       // 供应商描述
	Active       bool   `json:"active"`                            // 是否启用
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// VendorApply 供应商申请请求。
type VendorApply struct {
	Name        string `json:"name" binding:"required"`           // 供应商名称
	Email       string `json:"email" binding:"required,email"`    // 供应商邮箱
	Description string `json:"description"`                       // 申请描述
}
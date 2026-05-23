// Package request 定义用户相关的 HTTP 请求结构体。
// 用于参数绑定和校验。
package request

import "time"

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`  // 用户名，3-64 位
	Password string `json:"password" binding:"required,min=6,max=128"` // 密码，6-128 位
	Email    string `json:"email" binding:"omitempty,email"`           // 邮箱，可选
	Phone    string `json:"phone" binding:"omitempty"`                 // 手机号，可选
	Nickname string `json:"nickname" binding:"omitempty,max=64"`       // 昵称，可选
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// CreateUserRequest 创建用户请求（管理端）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`  // 用户名
	Password string `json:"password" binding:"required,min=6,max=128"` // 密码
	Email    string `json:"email" binding:"omitempty,email"`           // 邮箱
	Phone    string `json:"phone" binding:"omitempty"`                 // 手机号
	Nickname string `json:"nickname" binding:"omitempty,max=64"`       // 昵称
	Avatar   string `json:"avatar" binding:"omitempty,url"`            // 头像 URL
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`      // 状态：0-禁用 1-启用
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`       // 邮箱
	Phone    string `json:"phone" binding:"omitempty"`             // 手机号
	Nickname string `json:"nickname" binding:"omitempty,max=64"`   // 昵称
	Avatar   string `json:"avatar" binding:"omitempty,url"`        // 头像 URL
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`  // 状态：0-禁用 1-启用
}

// CreateAddressRequest 创建地址请求
type CreateAddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"required,max=64"` // 收件人姓名
	Phone         string `json:"phone" binding:"required,max=20"`          // 收件人手机号
	Province      string `json:"province" binding:"required,max=32"`       // 省份
	City          string `json:"city" binding:"required,max=32"`           // 城市
	District      string `json:"district" binding:"required,max=32"`       // 区/县
	Detail        string `json:"detail" binding:"required,max=256"`        // 详细地址
	IsDefault     bool   `json:"is_default"`                                // 是否默认地址
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"omitempty,max=64"` // 收件人姓名
	Phone         string `json:"phone" binding:"omitempty,max=20"`          // 收件人手机号
	Province      string `json:"province" binding:"omitempty,max=32"`       // 省份
	City          string `json:"city" binding:"omitempty,max=32"`           // 城市
	District      string `json:"district" binding:"omitempty,max=32"`       // 区/县
	Detail        string `json:"detail" binding:"omitempty,max=256"`        // 详细地址
	IsDefault     bool   `json:"is_default"`                                // 是否默认地址
}

// CreateExternalAssociationRequest 创建外部关联请求
type CreateExternalAssociationRequest struct {
	Platform     string `json:"platform" binding:"required,max=32"`    // 第三方平台标识
	ExternalID   string `json:"external_id" binding:"required,max=128"` // 第三方平台用户标识
	ExternalData string `json:"external_data" binding:"omitempty,max=1024"` // 附加数据
}

// CreateDownloadableProductRequest 创建可下载产品请求
type CreateDownloadableProductRequest struct {
	ProductID   string     `json:"product_id" binding:"required,max=64"`  // 产品标识
	ProductName string     `json:"product_name" binding:"required,max=128"` // 产品名称
	DownloadURL string     `json:"download_url" binding:"omitempty,url"`  // 下载地址
	ExpireAt    *time.Time `json:"expire_at"`                             // 过期时间，nil 表示永不过期
}
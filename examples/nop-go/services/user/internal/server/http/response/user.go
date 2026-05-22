// Package response 定义 HTTP 响应结构体。
// 每个响应结构体对应一个 API 端点的出参，
// 不包含敏感字段（如密码哈希、令牌等）。
package response

import "time"

// ============================================================
// 认证相关响应
// ============================================================

// LoginResponse 登录成功响应。
// 对应 POST /api/v1/auth/login
type LoginResponse struct {
	Token     string    `json:"token"`      // JWT 令牌
	ExpiresAt time.Time `json:"expires_at"` // 令牌过期时间
	UserInfo  UserInfo  `json:"user_info"`  // 用户基本信息
}

// RegisterResponse 注册成功响应。
// 对应 POST /api/v1/auth/register
type RegisterResponse struct {
	ID       uint   `json:"id"`       // 用户 ID
	Username string `json:"username"` // 用户名
	Email    string `json:"email"`    // 邮箱
}

// PasswordRecoveryResponse 密码恢复响应。
// 对应 POST /api/v1/auth/password-recovery
type PasswordRecoveryResponse struct {
	Message string `json:"message"` // 提示信息
}

// ConfirmPasswordRecoveryResponse 确认密码恢复响应。
// 对应 POST /api/v1/auth/password-recovery/confirm
type ConfirmPasswordRecoveryResponse struct {
	Message string `json:"message"` // 提示信息
}

// MultiFactorVerificationResponse 多因素验证响应。
// 对应 GET /api/v1/auth/multi-factor-verification
type MultiFactorVerificationResponse struct {
	Verified bool `json:"verified"` // 是否验证通过
}

// LogoutResponse 登出响应。
// 对应 POST /api/v1/auth/logout
type LogoutResponse struct {
	Message string `json:"message"` // 提示信息
}

// ============================================================
// 用户信息相关响应
// ============================================================

// UserInfo 用户基本信息。
// 多个响应中复用的用户信息结构体。
type UserInfo struct {
	ID             uint       `json:"id"`              // 用户 ID
	Username       string     `json:"username"`        // 用户名
	Email          string     `json:"email"`           // 邮箱
	Phone          string     `json:"phone"`           // 手机号
	Active         bool       `json:"active"`          // 是否激活
	AvatarURL      string     `json:"avatar_url"`      // 头像 URL
	MFAEnabled     bool       `json:"mfa_enabled"`     // 是否启用 MFA
	LastLoginAt    *time.Time `json:"last_login_at"`   // 最后登录时间
	LastActivityAt *time.Time `json:"last_activity_at"` // 最后活动时间
	CreatedAt      time.Time  `json:"created_at"`      // 注册时间
}

// UpdateUserInfoResponse 更新用户信息响应。
// 对应 PUT /api/v1/users/info
type UpdateUserInfoResponse struct {
	UserInfo UserInfo `json:"user_info"` // 更新后的用户信息
}

// ChangePasswordResponse 修改密码响应。
// 对应 PUT /api/v1/users/password
type ChangePasswordResponse struct {
	Message string `json:"message"` // 提示信息
}

// ============================================================
// 地址相关响应
// ============================================================

// Address 地址信息。
// 多个响应中复用的地址结构体。
type Address struct {
	ID              uint      `json:"id"`               // 地址 ID
	FirstName       string    `json:"first_name"`       // 名
	LastName        string    `json:"last_name"`        // 姓
	Email           string    `json:"email"`            // 邮箱
	Phone           string    `json:"phone"`            // 电话
	Fax             string    `json:"fax"`              // 传真
	Company         string    `json:"company"`          // 公司
	CountryID       uint      `json:"country_id"`       // 国家 ID
	StateProvinceID uint      `json:"state_province_id"` // 省/州 ID
	City            string    `json:"city"`             // 城市
	Address1        string    `json:"address1"`         // 地址行1
	Address2        string    `json:"address2"`         // 地址行2
	ZipPostalCode   string    `json:"zip_postal_code"`  // 邮编
	IsDefault       bool      `json:"is_default"`       // 是否默认地址
	CreatedAt       time.Time `json:"created_at"`       // 创建时间
}

// AddressListResponse 地址列表响应。
// 对应 GET /api/v1/users/addresses
type AddressListResponse struct {
	Items []Address `json:"items"` // 地址列表
}

// AddAddressResponse 添加地址响应。
// 对应 POST /api/v1/users/addresses
type AddAddressResponse struct {
	ID uint `json:"id"` // 新创建的地址 ID
}

// ============================================================
// 头像相关响应
// ============================================================

// AvatarResponse 头像信息响应。
// 对应 GET /api/v1/users/avatar
type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"` // 头像 URL（空表示无头像）
}

// UploadAvatarResponse 上传头像响应。
// 对应 POST /api/v1/users/avatar/upload
type UploadAvatarResponse struct {
	AvatarURL string `json:"avatar_url"` // 新头像 URL
}

// ============================================================
// 其他响应
// ============================================================

// CheckUsernameResponse 检查用户名可用性响应。
// 对应 POST /api/v1/users/check-username
type CheckUsernameResponse struct {
	Available bool `json:"available"` // 是否可用
}

// DownloadableProductResponse 可下载产品响应。
// 对应 GET /api/v1/users/downloadable-products
type DownloadableProductResponse struct {
	ID            uint       `json:"id"`             // 记录 ID
	ProductID     uint       `json:"product_id"`     // 产品 ID
	DownloadCount int        `json:"download_count"` // 已下载次数
	MaxDownloads  int        `json:"max_downloads"`  // 最大下载次数
	IsActivated   bool       `json:"is_activated"`   // 是否已激活
	ExpiresAt     *time.Time `json:"expires_at"`     // 过期时间
}

// DownloadableProductListResponse 可下载产品列表响应。
type DownloadableProductListResponse struct {
	Items []DownloadableProductResponse `json:"items"` // 产品列表
}

// RemoveExternalAssociationResponse 移除外部关联响应。
// 对应 POST /api/v1/users/external-association/remove
type RemoveExternalAssociationResponse struct {
	Message string `json:"message"` // 提示信息
}

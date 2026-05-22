// Package request 定义 HTTP 请求结构体。
// 每个请求结构体对应一个 API 端点的入参，
// 包含 binding tag 用于 Gin 自动校验。
package request

// ============================================================
// 认证相关请求
// ============================================================

// LoginRequest 登录请求。
// 对应 POST /api/v1/auth/login
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`  // 用户名或邮箱
	Password string `json:"password" binding:"required,min=6" example:"123456"` // 密码
}

// RegisterRequest 注册请求。
// 对应 POST /api/v1/auth/register
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64" example:"newuser"` // 用户名
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`   // 邮箱
	Password string `json:"password" binding:"required,min=6" example:"123456"`         // 密码
}

// PasswordRecoveryRequest 密码恢复请求。
// 对应 POST /api/v1/auth/password-recovery
type PasswordRecoveryRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"` // 注册邮箱
}

// ConfirmPasswordRecoveryRequest 确认密码恢复请求。
// 对应 POST /api/v1/auth/password-recovery/confirm
type ConfirmPasswordRecoveryRequest struct {
	Token       string `json:"token" binding:"required" example:"recovery-token-xxx"` // 恢复令牌
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpass123"` // 新密码
}

// MultiFactorVerificationRequest 多因素验证请求。
// 对应 GET /api/v1/auth/multi-factor-verification
type MultiFactorVerificationRequest struct {
	Code string `form:"code" binding:"required" example:"123456"` // MFA 验证码
}

// ============================================================
// 用户信息相关请求
// ============================================================

// UpdateUserInfoRequest 更新用户信息请求。
// 对应 PUT /api/v1/users/info
type UpdateUserInfoRequest struct {
	Email string `json:"email" binding:"omitempty,email" example:"new@example.com"` // 邮箱（可选更新）
	Phone string `json:"phone" binding:"omitempty" example:"13800138000"`          // 手机号（可选更新）
}

// ChangePasswordRequest 修改密码请求。
// 对应 PUT /api/v1/users/password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6" example:"oldpass"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpass"` // 新密码
}

// ============================================================
// 地址相关请求
// ============================================================

// AddAddressRequest 添加地址请求。
// 对应 POST /api/v1/users/addresses
type AddAddressRequest struct {
	FirstName       string `json:"first_name" binding:"required" example:"张"`              // 名
	LastName        string `json:"last_name" binding:"required" example:"三"`               // 姓
	Email           string `json:"email" binding:"omitempty,email" example:"zhang@example.com"` // 邮箱
	Phone           string `json:"phone" binding:"omitempty" example:"13800138000"`        // 电话
	Fax             string `json:"fax" binding:"omitempty"`                               // 传真
	Company         string `json:"company" binding:"omitempty" example:"ABC公司"`           // 公司
	CountryID       uint   `json:"country_id" binding:"required" example:"1"`              // 国家 ID
	StateProvinceID uint   `json:"state_province_id" binding:"omitempty" example:"10"`     // 省/州 ID
	City            string `json:"city" binding:"required" example:"北京"`                  // 城市
	Address1        string `json:"address1" binding:"required" example:"朝阳区xxx路"`         // 地址行1
	Address2        string `json:"address2" binding:"omitempty" example:"1号楼"`            // 地址行2
	ZipPostalCode   string `json:"zip_postal_code" binding:"omitempty" example:"100000"`   // 邮编
	IsDefault       bool   `json:"is_default" example:"false"`                            // 是否默认地址
}

// UpdateAddressRequest 编辑地址请求。
// 对应 PUT /api/v1/users/addresses/:id
type UpdateAddressRequest struct {
	FirstName       string `json:"first_name" binding:"required" example:"张"`              // 名
	LastName        string `json:"last_name" binding:"required" example:"三"`               // 姓
	Email           string `json:"email" binding:"omitempty,email" example:"zhang@example.com"` // 邮箱
	Phone           string `json:"phone" binding:"omitempty" example:"13800138000"`        // 电话
	Fax             string `json:"fax" binding:"omitempty"`                               // 传真
	Company         string `json:"company" binding:"omitempty" example:"ABC公司"`           // 公司
	CountryID       uint   `json:"country_id" binding:"required" example:"1"`              // 国家 ID
	StateProvinceID uint   `json:"state_province_id" binding:"omitempty" example:"10"`     // 省/州 ID
	City            string `json:"city" binding:"required" example:"北京"`                  // 城市
	Address1        string `json:"address1" binding:"required" example:"朝阳区xxx路"`         // 地址行1
	Address2        string `json:"address2" binding:"omitempty" example:"1号楼"`            // 地址行2
	ZipPostalCode   string `json:"zip_postal_code" binding:"omitempty" example:"100000"`   // 邮编
	IsDefault       bool   `json:"is_default" example:"false"`                            // 是否默认地址
}

// ============================================================
// 头像相关请求
// ============================================================

// UploadAvatarRequest 上传头像请求。
// 对应 POST /api/v1/users/avatar/upload
// 注意：头像上传使用 multipart/form-data，此处仅定义元数据字段。
// 实际文件通过 c.FormFile("file") 获取。
type UploadAvatarRequest struct {
	// 文件通过 multipart form 上传，此处无需定义文件字段
}

// ============================================================
// 其他请求
// ============================================================

// CheckUsernameRequest 检查用户名可用性请求。
// 对应 POST /api/v1/users/check-username
type CheckUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64" example:"newuser"` // 待检查的用户名
}

// RemoveExternalAssociationRequest 移除外部关联请求。
// 对应 POST /api/v1/users/external-association/remove
type RemoveExternalAssociationRequest struct {
	Provider string `json:"provider" binding:"required" example:"wechat"` // 第三方平台名称
}

// Package biz 业务逻辑层。
// 包含用户服务的领域实体、仓储接口和用例实现。
// 本层负责核心业务规则，不依赖任何具体的数据访问实现。
package biz

import (
	"context"
	"time"
)

// ============================================================
// 领域实体
// ============================================================

// User 用户领域实体。
// 对应 nopCommerce 中的 Customer 核心概念，包含用户基本信息、认证信息和状态。
type User struct {
	ID                    uint      // 用户唯一标识
	Username              string    // 用户名（唯一）
	Email                 string    // 邮箱地址
	PasswordHash          string    // 密码哈希值（不对外暴露）
	PasswordSalt          string    // 密码盐值
	Active                bool      // 是否激活
	IsSystemAccount       bool      // 是否系统内置账号
	SystemName            string    // 系统名称（内置账号标识）
	LastLoginAt           *time.Time // 最后登录时间
	LastActivityAt        *time.Time // 最后活动时间
	AvatarURL             string    // 头像 URL
	Phone                 string    // 手机号
	PasswordRecoveryToken string    // 密码恢复令牌
	MFASecret             string    // 多因素认证密钥
	MFAEnabled            bool      // 是否启用多因素认证
	CreatedAt             time.Time // 创建时间
	UpdatedAt             time.Time // 更新时间
}

// Address 用户地址领域实体。
// 对应 nopCommerce 中的 Address 概念，一个用户可拥有多个地址。
type Address struct {
	ID             uint      // 地址唯一标识
	UserID         uint      // 所属用户 ID
	FirstName      string    // 名
	LastName       string    // 姓
	Email          string    // 邮箱
	Phone          string    // 电话
	Fax            string    // 传真
	Company        string    // 公司
	CountryID      uint      // 国家 ID
	StateProvinceID uint     // 省/州 ID
	City           string    // 城市
	Address1       string    // 地址行1
	Address2       string    // 地址行2
	ZipPostalCode  string    // 邮编
	IsDefault      bool      // 是否默认地址
	CreatedAt      time.Time // 创建时间
	UpdatedAt      time.Time // 更新时间
}

// ExternalAssociation 外部关联领域实体。
// 记录用户与第三方平台（如微信、Google 等）的绑定关系。
type ExternalAssociation struct {
	ID           uint      // 关联唯一标识
	UserID       uint      // 所属用户 ID
	Provider     string    // 第三方平台名称（如 wechat、google）
	ProviderUID  string    // 第三方平台用户标识
	AccessToken  string    // 访问令牌
	CreatedAt    time.Time // 创建时间
}

// DownloadableProduct 可下载产品领域实体。
// 记录用户已购买的可下载产品信息。
type DownloadableProduct struct {
	ID              uint      // 记录唯一标识
	UserID          uint      // 所属用户 ID
	OrderItemID     uint      // 订单项 ID
	ProductID       uint      // 产品 ID
	DownloadCount   int       // 已下载次数
	MaxDownloads    int       // 最大下载次数（0 表示不限）
	IsActivated     bool      // 是否已激活下载
	ExpiresAt       *time.Time // 过期时间（nil 表示永不过期）
	CreatedAt       time.Time // 创建时间
}

// ============================================================
// 仓储接口
// ============================================================

// UserRepository 用户仓储接口。
// 定义用户实体的持久化操作契约，由 data 层实现。
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error
	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id uint) (*User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)
	// Update 更新用户信息
	Update(ctx context.Context, user *User) error
	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, id uint, passwordHash, passwordSalt string) error
	// SetPasswordRecoveryToken 设置密码恢复令牌
	SetPasswordRecoveryToken(ctx context.Context, id uint, token string) error
	// GetByPasswordRecoveryToken 根据密码恢复令牌获取用户
	GetByPasswordRecoveryToken(ctx context.Context, token string) (*User, error)
	// UpdateLastLogin 更新最后登录时间
	UpdateLastLogin(ctx context.Context, id uint) error
	// CheckUsernameAvailable 检查用户名是否可用
	CheckUsernameAvailable(ctx context.Context, username string) (bool, error)
	// UpdateAvatar 更新用户头像
	UpdateAvatar(ctx context.Context, id uint, avatarURL string) error
	// RemoveAvatar 移除用户头像
	RemoveAvatar(ctx context.Context, id uint) error
}

// AddressRepository 地址仓储接口。
// 定义用户地址的持久化操作契约。
type AddressRepository interface {
	// ListByUserID 获取用户的所有地址
	ListByUserID(ctx context.Context, userID uint) ([]*Address, error)
	// GetByID 根据 ID 获取地址
	GetByID(ctx context.Context, id uint) (*Address, error)
	// Create 创建地址
	Create(ctx context.Context, address *Address) error
	// Update 更新地址
	Update(ctx context.Context, address *Address) error
	// Delete 删除地址
	Delete(ctx context.Context, id uint) error
}

// ExternalAssociationRepository 外部关联仓储接口。
// 定义用户外部关联的持久化操作契约。
type ExternalAssociationRepository interface {
	// Remove 移除外部关联
	Remove(ctx context.Context, userID uint, provider string) error
}

// DownloadableProductRepository 可下载产品仓储接口。
// 定义可下载产品的持久化操作契约。
type DownloadableProductRepository interface {
	// ListByUserID 获取用户的可下载产品列表
	ListByUserID(ctx context.Context, userID uint) ([]*DownloadableProduct, error)
}

// ============================================================
// 用例
// ============================================================

// UserUseCase 用户用例。
// 封装用户服务的核心业务逻辑，包括认证、信息管理和地址管理。
type UserUseCase struct {
	userRepo    UserRepository
	addressRepo AddressRepository
	extRepo     ExternalAssociationRepository
	dlRepo      DownloadableProductRepository
}

// NewUserUseCase 创建用户用例。
// 依赖注入所有需要的仓储接口。
func NewUserUseCase(
	userRepo UserRepository,
	addressRepo AddressRepository,
	extRepo ExternalAssociationRepository,
	dlRepo DownloadableProductRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		extRepo:     extRepo,
		dlRepo:      dlRepo,
	}
}

// ---------- 认证相关 ----------

// Login 用户登录。
// 根据用户名/邮箱查找用户，验证密码，更新登录时间。
func (uc *UserUseCase) Login(ctx context.Context, username, password string) (*User, error) {
	// 先尝试按用户名查找，再按邮箱查找
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		user, err = uc.userRepo.GetByEmail(ctx, username)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	// 验证密码（实际项目中应使用 bcrypt 等安全哈希比较）
	if user.PasswordHash != password {
		return nil, ErrInvalidCredentials
	}

	// 检查账号是否激活
	if !user.Active {
		return nil, ErrUserInactive
	}

	// 更新最后登录时间
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID)

	return user, nil
}

// Logout 用户登出。
// 清理服务端会话状态（如需要），当前为无状态 JWT 模式则无需特殊处理。
func (uc *UserUseCase) Logout(ctx context.Context, userID uint) error {
	// 无状态 JWT 模式下登出由客户端删除 token 即可
	// 如需服务端黑名单机制可在此扩展
	return nil
}

// Register 用户注册。
// 创建新用户，检查用户名和邮箱唯一性。
func (uc *UserUseCase) Register(ctx context.Context, username, email, password string) (*User, error) {
	// 检查用户名是否已存在
	available, err := uc.userRepo.CheckUsernameAvailable(ctx, username)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrUsernameTaken
	}

	// 检查邮箱是否已注册
	existing, _ := uc.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	now := time.Now()
	user := &User{
		Username:        username,
		Email:           email,
		PasswordHash:    password, // 实际项目中应使用 bcrypt.GenerateFromPassword
		Active:          true,
		LastActivityAt:  &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// PasswordRecovery 密码恢复（发送恢复链接）。
// 生成恢复令牌并关联到用户，实际项目中应发送邮件。
func (uc *UserUseCase) PasswordRecovery(ctx context.Context, email string) error {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// 安全考虑：不暴露邮箱是否存在
		return nil
	}

	// 生成恢复令牌（实际项目中应使用 crypto/rand 生成安全令牌）
	token := "recovery-token-" + email
	return uc.userRepo.SetPasswordRecoveryToken(ctx, user.ID, token)
}

// ConfirmPasswordRecovery 确认密码恢复。
// 验证恢复令牌并更新密码。
func (uc *UserUseCase) ConfirmPasswordRecovery(ctx context.Context, token, newPassword string) error {
	user, err := uc.userRepo.GetByPasswordRecoveryToken(ctx, token)
	if err != nil {
		return ErrInvalidRecoveryToken
	}

	// 更新密码（实际项目中应使用 bcrypt.GenerateFromPassword）
	passwordHash := newPassword
	passwordSalt := ""
	return uc.userRepo.UpdatePassword(ctx, user.ID, passwordHash, passwordSalt)
}

// MultiFactorVerification 多因素验证。
// 验证用户的多因素认证码。
func (uc *UserUseCase) MultiFactorVerification(ctx context.Context, userID uint, code string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.MFAEnabled {
		return ErrMFANotEnabled
	}

	// 实际项目中应使用 TOTP 库验证 code
	// 此处为简化实现
	if code == "" {
		return ErrInvalidMFACode
	}

	return nil
}

// ---------- 用户信息相关 ----------

// GetUserInfo 获取用户信息。
func (uc *UserUseCase) GetUserInfo(ctx context.Context, userID uint) (*User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

// UpdateUserInfo 更新用户信息。
func (uc *UserUseCase) UpdateUserInfo(ctx context.Context, userID uint, email, phone string) (*User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if email != "" {
		user.Email = email
	}
	if phone != "" {
		user.Phone = phone
	}
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword 修改密码。
func (uc *UserUseCase) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if user.PasswordHash != oldPassword {
		return ErrInvalidOldPassword
	}

	// 更新密码（实际项目中应使用 bcrypt.GenerateFromPassword）
	return uc.userRepo.UpdatePassword(ctx, userID, newPassword, "")
}

// ---------- 地址相关 ----------

// ListAddresses 获取用户地址列表。
func (uc *UserUseCase) ListAddresses(ctx context.Context, userID uint) ([]*Address, error) {
	return uc.addressRepo.ListByUserID(ctx, userID)
}

// AddAddress 添加用户地址。
func (uc *UserUseCase) AddAddress(ctx context.Context, address *Address) error {
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()
	return uc.addressRepo.Create(ctx, address)
}

// UpdateAddress 更新用户地址。
func (uc *UserUseCase) UpdateAddress(ctx context.Context, address *Address) error {
	address.UpdatedAt = time.Now()
	return uc.addressRepo.Update(ctx, address)
}

// DeleteAddress 删除用户地址。
func (uc *UserUseCase) DeleteAddress(ctx context.Context, addressID uint) error {
	return uc.addressRepo.Delete(ctx, addressID)
}

// ---------- 头像相关 ----------

// GetAvatar 获取用户头像信息。
func (uc *UserUseCase) GetAvatar(ctx context.Context, userID uint) (*User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

// UploadAvatar 上传用户头像。
// avatarURL 为上传后得到的图片访问地址。
func (uc *UserUseCase) UploadAvatar(ctx context.Context, userID uint, avatarURL string) error {
	return uc.userRepo.UpdateAvatar(ctx, userID, avatarURL)
}

// RemoveAvatar 移除用户头像。
func (uc *UserUseCase) RemoveAvatar(ctx context.Context, userID uint) error {
	return uc.userRepo.RemoveAvatar(ctx, userID)
}

// ---------- 其他 ----------

// CheckUsernameAvailability 检查用户名可用性。
func (uc *UserUseCase) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	return uc.userRepo.CheckUsernameAvailable(ctx, username)
}

// GetDownloadableProducts 获取用户的可下载产品列表。
func (uc *UserUseCase) GetDownloadableProducts(ctx context.Context, userID uint) ([]*DownloadableProduct, error) {
	return uc.dlRepo.ListByUserID(ctx, userID)
}

// RemoveExternalAssociation 移除外部关联。
func (uc *UserUseCase) RemoveExternalAssociation(ctx context.Context, userID uint, provider string) error {
	return uc.extRepo.Remove(ctx, userID, provider)
}

// ============================================================
// 业务错误定义
// ============================================================

// 业务错误变量，用于在 biz 层统一返回业务异常。
// handler 层根据错误类型选择合适的 HTTP 响应。
var (
	ErrInvalidCredentials   = NewBizError("用户名或密码错误")
	ErrUserInactive         = NewBizError("用户未激活")
	ErrUsernameTaken        = NewBizError("用户名已被占用")
	ErrEmailTaken           = NewBizError("邮箱已被注册")
	ErrInvalidRecoveryToken = NewBizError("无效的密码恢复令牌")
	ErrMFANotEnabled        = NewBizError("未启用多因素认证")
	ErrInvalidMFACode       = NewBizError("无效的多因素验证码")
	ErrInvalidOldPassword   = NewBizError("旧密码不正确")
)

// BizError 业务错误类型。
// 封装业务逻辑层错误，便于上层识别和处理。
type BizError struct {
	Message string
}

// NewBizError 创建业务错误。
func NewBizError(message string) *BizError {
	return &BizError{Message: message}
}

// Error 实现 error 接口。
func (e *BizError) Error() string {
	return e.Message
}

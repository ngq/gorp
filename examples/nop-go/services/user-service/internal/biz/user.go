// Package biz 定义用户服务的业务领域层，包含领域模型、仓储接口和用例。
// 本文件合并了原 user 服务的核心业务逻辑：用户、地址、外部关联、可下载产品。
package biz

import (
	"context"
	"fmt"
	"time"
)

// ======================== 错误定义 ========================

// BizError 业务错误类型，携带错误码和消息
type BizError struct {
	Code    int    // 业务错误码
	Message string // 错误描述
}

func (e *BizError) Error() string {
	return fmt.Sprintf("biz error: code=%d, message=%s", e.Code, e.Message)
}

// NewBizError 创建业务错误
func NewBizError(code int, message string) *BizError {
	return &BizError{Code: code, Message: message}
}

// 通用错误码定义
var (
	ErrUserNotFound                   = NewBizError(40401, "用户不存在")
	ErrUserAlreadyExists              = NewBizError(40901, "用户已存在")
	ErrAddressNotFound                = NewBizError(40402, "地址不存在")
	ErrExternalAssociationNotFound    = NewBizError(40403, "外部关联不存在")
	ErrDownloadableProductNotFound    = NewBizError(40404, "可下载产品不存在")
	ErrInvalidParameter               = NewBizError(40001, "参数无效")
	ErrCreateUserFailed               = NewBizError(50001, "创建用户失败")
	ErrUpdateUserFailed               = NewBizError(50002, "更新用户失败")
	ErrDeleteUserFailed               = NewBizError(50003, "删除用户失败")
	ErrCreateAddressFailed            = NewBizError(50004, "创建地址失败")
	ErrUpdateAddressFailed            = NewBizError(50005, "更新地址失败")
	ErrDeleteAddressFailed            = NewBizError(50006, "删除地址失败")
	ErrCreateExternalAssociationFailed = NewBizError(50007, "创建外部关联失败")
	ErrDeleteExternalAssociationFailed = NewBizError(50008, "删除外部关联失败")
	ErrCreateDownloadableProductFailed = NewBizError(50009, "创建可下载产品失败")
	ErrDeleteDownloadableProductFailed = NewBizError(50010, "删除可下载产品失败")
)

// ======================== 领域模型 ========================

// User 用户领域模型
type User struct {
	ID        uint      // 用户唯一标识
	Username  string    // 用户名
	Email     string    // 邮箱
	Phone     string    // 手机号
	Password  string    // 密码（加密存储）
	Nickname  string    // 昵称
	Avatar    string    // 头像 URL
	Status    int       // 状态：0-禁用 1-启用
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// Address 用户地址领域模型
type Address struct {
	ID           uint   // 地址唯一标识
	UserID       uint   // 所属用户 ID
	RecipientName string // 收件人姓名
	Phone        string // 收件人手机号
	Province     string // 省份
	City         string // 城市
	District     string // 区/县
	Detail       string // 详细地址
	IsDefault    bool   // 是否默认地址
}

// ExternalAssociation 外部关联领域模型（用户与第三方平台的绑定关系）
type ExternalAssociation struct {
	ID           uint   // 关联唯一标识
	UserID       uint   // 所属用户 ID
	Platform     string // 第三方平台标识（如 wechat、alipay）
	ExternalID   string // 第三方平台用户标识
	ExternalData string // 第三方平台返回的附加数据
}

// DownloadableProduct 可下载产品领域模型
type DownloadableProduct struct {
	ID          uint   // 产品唯一标识
	UserID      uint   // 所属用户 ID
	ProductID   string // 产品标识
	ProductName string // 产品名称
	DownloadURL string // 下载地址
	ExpireAt    *time.Time // 过期时间，nil 表示永不过期
}

// ======================== 仓储接口 ========================

// UserRepository 用户仓储接口，定义用户数据的持久化操作
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) (*User, error)
	// Update 更新用户
	Update(ctx context.Context, user *User) (*User, error)
	// Delete 删除用户
	Delete(ctx context.Context, id uint) error
	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id uint) (*User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)
	// List 获取用户列表
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
}

// AddressRepository 地址仓储接口，定义用户地址的持久化操作
type AddressRepository interface {
	// Create 创建地址
	Create(ctx context.Context, address *Address) (*Address, error)
	// Update 更新地址
	Update(ctx context.Context, address *Address) (*Address, error)
	// Delete 删除地址
	Delete(ctx context.Context, id uint) error
	// GetByID 根据 ID 获取地址
	GetByID(ctx context.Context, id uint) (*Address, error)
	// ListByUserID 获取用户的地址列表
	ListByUserID(ctx context.Context, userID uint) ([]*Address, error)
}

// ExternalAssociationRepository 外部关联仓储接口
type ExternalAssociationRepository interface {
	// Create 创建外部关联
	Create(ctx context.Context, ea *ExternalAssociation) (*ExternalAssociation, error)
	// Delete 删除外部关联
	Delete(ctx context.Context, id uint) error
	// GetByID 根据 ID 获取外部关联
	GetByID(ctx context.Context, id uint) (*ExternalAssociation, error)
	// ListByUserID 获取用户的外部关联列表
	ListByUserID(ctx context.Context, userID uint) ([]*ExternalAssociation, error)
}

// DownloadableProductRepository 可下载产品仓储接口
type DownloadableProductRepository interface {
	// Create 创建可下载产品
	Create(ctx context.Context, dp *DownloadableProduct) (*DownloadableProduct, error)
	// Delete 删除可下载产品
	Delete(ctx context.Context, id uint) error
	// GetByID 根据 ID 获取可下载产品
	GetByID(ctx context.Context, id uint) (*DownloadableProduct, error)
	// ListByUserID 获取用户的可下载产品列表
	ListByUserID(ctx context.Context, userID uint) ([]*DownloadableProduct, error)
}

// ======================== 用例 ========================

// UserUseCase 用户业务用例，编排用户相关的业务流程
type UserUseCase struct {
	userRepo                   UserRepository
	addressRepo                AddressRepository
	externalAssociationRepo    ExternalAssociationRepository
	downloadableProductRepo    DownloadableProductRepository
}

// NewUserUseCase 创建用户用例实例
func NewUserUseCase(
	userRepo UserRepository,
	addressRepo AddressRepository,
	externalAssociationRepo ExternalAssociationRepository,
	downloadableProductRepo DownloadableProductRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo:                userRepo,
		addressRepo:             addressRepo,
		externalAssociationRepo: externalAssociationRepo,
		downloadableProductRepo: downloadableProductRepo,
	}
}

// ---- 用户 CRUD ----

// CreateUser 创建用户
func (uc *UserUseCase) CreateUser(ctx context.Context, user *User) (*User, error) {
	// 检查用户名是否已存在
	existing, _ := uc.userRepo.GetByUsername(ctx, user.Username)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}
	created, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, ErrCreateUserFailed
	}
	return created, nil
}

// UpdateUser 更新用户信息
func (uc *UserUseCase) UpdateUser(ctx context.Context, user *User) (*User, error) {
	_, err := uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	updated, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, ErrUpdateUserFailed
	}
	return updated, nil
}

// DeleteUser 删除用户
func (uc *UserUseCase) DeleteUser(ctx context.Context, id uint) error {
	_, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return ErrDeleteUserFailed
	}
	return nil
}

// GetUser 根据 ID 获取用户
func (uc *UserUseCase) GetUser(ctx context.Context, id uint) (*User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ListUsers 获取用户列表
func (uc *UserUseCase) ListUsers(ctx context.Context, offset, limit int) ([]*User, int64, error) {
	users, total, err := uc.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ---- 地址 CRUD ----

// CreateAddress 创建地址
func (uc *UserUseCase) CreateAddress(ctx context.Context, address *Address) (*Address, error) {
	// 校验用户是否存在
	_, err := uc.userRepo.GetByID(ctx, address.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	created, err := uc.addressRepo.Create(ctx, address)
	if err != nil {
		return nil, ErrCreateAddressFailed
	}
	return created, nil
}

// UpdateAddress 更新地址
func (uc *UserUseCase) UpdateAddress(ctx context.Context, address *Address) (*Address, error) {
	_, err := uc.addressRepo.GetByID(ctx, address.ID)
	if err != nil {
		return nil, ErrAddressNotFound
	}
	updated, err := uc.addressRepo.Update(ctx, address)
	if err != nil {
		return nil, ErrUpdateAddressFailed
	}
	return updated, nil
}

// DeleteAddress 删除地址
func (uc *UserUseCase) DeleteAddress(ctx context.Context, id uint) error {
	_, err := uc.addressRepo.GetByID(ctx, id)
	if err != nil {
		return ErrAddressNotFound
	}
	if err := uc.addressRepo.Delete(ctx, id); err != nil {
		return ErrDeleteAddressFailed
	}
	return nil
}

// ListAddresses 获取用户地址列表
func (uc *UserUseCase) ListAddresses(ctx context.Context, userID uint) ([]*Address, error) {
	return uc.addressRepo.ListByUserID(ctx, userID)
}

// ---- 外部关联 CRUD ----

// CreateExternalAssociation 创建外部关联
func (uc *UserUseCase) CreateExternalAssociation(ctx context.Context, ea *ExternalAssociation) (*ExternalAssociation, error) {
	// 校验用户是否存在
	_, err := uc.userRepo.GetByID(ctx, ea.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	created, err := uc.externalAssociationRepo.Create(ctx, ea)
	if err != nil {
		return nil, ErrCreateExternalAssociationFailed
	}
	return created, nil
}

// DeleteExternalAssociation 删除外部关联
func (uc *UserUseCase) DeleteExternalAssociation(ctx context.Context, id uint) error {
	_, err := uc.externalAssociationRepo.GetByID(ctx, id)
	if err != nil {
		return ErrExternalAssociationNotFound
	}
	if err := uc.externalAssociationRepo.Delete(ctx, id); err != nil {
		return ErrDeleteExternalAssociationFailed
	}
	return nil
}

// ListExternalAssociations 获取用户外部关联列表
func (uc *UserUseCase) ListExternalAssociations(ctx context.Context, userID uint) ([]*ExternalAssociation, error) {
	return uc.externalAssociationRepo.ListByUserID(ctx, userID)
}

// ---- 可下载产品 CRUD ----

// CreateDownloadableProduct 创建可下载产品
func (uc *UserUseCase) CreateDownloadableProduct(ctx context.Context, dp *DownloadableProduct) (*DownloadableProduct, error) {
	// 校验用户是否存在
	_, err := uc.userRepo.GetByID(ctx, dp.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	created, err := uc.downloadableProductRepo.Create(ctx, dp)
	if err != nil {
		return nil, ErrCreateDownloadableProductFailed
	}
	return created, nil
}

// DeleteDownloadableProduct 删除可下载产品
func (uc *UserUseCase) DeleteDownloadableProduct(ctx context.Context, id uint) error {
	_, err := uc.downloadableProductRepo.GetByID(ctx, id)
	if err != nil {
		return ErrDownloadableProductNotFound
	}
	if err := uc.downloadableProductRepo.Delete(ctx, id); err != nil {
		return ErrDeleteDownloadableProductFailed
	}
	return nil
}

// ListDownloadableProducts 获取用户可下载产品列表
func (uc *UserUseCase) ListDownloadableProducts(ctx context.Context, userID uint) ([]*DownloadableProduct, error) {
	return uc.downloadableProductRepo.ListByUserID(ctx, userID)
}

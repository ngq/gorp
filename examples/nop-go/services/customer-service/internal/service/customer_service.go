// Package service 客户服务接口定义
// 用于测试从 Service 接口生成 Proto
package service

import (
	"context"
)

// CustomerServiceRPC 客户服务 RPC 接口
//
// 中文说明：
// - 定义客户相关的 RPC 方法；
// - 用于生成 gRPC Proto 定义；
// - 与 HTTP Handler 结构体 CustomerService 分离。
type CustomerServiceRPC interface {
	// Register 注册新客户
	Register(ctx context.Context, req *RegisterRequest) (*Customer, error)

	// Login 客户登录
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// GetCustomer 获取客户信息
	GetCustomer(ctx context.Context, req *GetCustomerRequest) (*Customer, error)

	// UpdateProfile 更新客户资料
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*Customer, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*Empty, error)

	// ListCustomers 客户列表
	ListCustomers(ctx context.Context, req *ListCustomersRequest) (*ListCustomersResponse, error)

	// ValidateCustomer 验证客户（供其他服务调用）
	ValidateCustomer(ctx context.Context, req *ValidateCustomerRequest) (*Empty, error)

	// GetPreferences 获取客户偏好设置
	GetPreferences(ctx context.Context, req *GetCustomerRequest) (*CustomerPreferences, error)
}

// AddressServiceRPC 地址服务 RPC 接口
//
// 中文说明：
// - 定义地址相关的 RPC 方法；
// - 用于生成 gRPC Proto 定义；
// - 与 HTTP Handler 分离。
type AddressServiceRPC interface {
	// CreateAddress 创建地址
	CreateAddress(ctx context.Context, req *Address) (*Address, error)

	// GetAddress 获取地址
	GetAddress(ctx context.Context, req *GetAddressRequest) (*Address, error)

	// ListAddresses 获取客户地址列表
	ListAddresses(ctx context.Context, req *ListAddressesRequest) (*ListAddressesResponse, error)

	// UpdateAddress 更新地址
	UpdateAddress(ctx context.Context, req *Address) (*Address, error)

	// DeleteAddress 删除地址
	DeleteAddress(ctx context.Context, req *DeleteAddressRequest) (*Empty, error)

	// SetDefaultBilling 设置默认账单地址
	SetDefaultBilling(ctx context.Context, req *SetDefaultAddressRequest) (*Empty, error)

	// SetDefaultShipping 设置默认配送地址
	SetDefaultShipping(ctx context.Context, req *SetDefaultAddressRequest) (*Empty, error)
}

// ========== 请求/响应类型 ==========

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Customer *Customer `json:"customer"`
	Token    string    `json:"token"`
}

// GetCustomerRequest 获取客户请求
type GetCustomerRequest struct {
	Id uint64 `json:"id"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Id        uint64 `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Gender    string `json:"gender"`
	Birthday  string `json:"birthday"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Id          uint64 `json:"id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ListCustomersRequest 客户列表请求
type ListCustomersRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

// ListCustomersResponse 客户列表响应
type ListCustomersResponse struct {
	Customers []*Customer `json:"customers"`
	Total     int64       `json:"total"`
}

// CustomerPreferences 客户偏好设置（测试 Map 类型）
type CustomerPreferences struct {
	CustomerId uint64            `json:"customer_id"`
	Settings   map[string]string `json:"settings"`
	Tags       map[string]int32  `json:"tags"`
}

// ValidateCustomerRequest 验证客户请求
type ValidateCustomerRequest struct {
	Id uint64 `json:"id"`
}

// GetAddressRequest 获取地址请求
type GetAddressRequest struct {
	Id uint64 `json:"id"`
}

// ListAddressesRequest 地址列表请求
type ListAddressesRequest struct {
	CustomerId uint64 `json:"customer_id"`
}

// ListAddressesResponse 地址列表响应
type ListAddressesResponse struct {
	Addresses []*Address `json:"addresses"`
}

// DeleteAddressRequest 删除地址请求
type DeleteAddressRequest struct {
	Id uint64 `json:"id"`
}

// SetDefaultAddressRequest 设置默认地址请求
type SetDefaultAddressRequest struct {
	CustomerId uint64 `json:"customer_id"`
	AddressId  uint64 `json:"address_id"`
}

// Customer 客户信息
type Customer struct {
	Id            uint64 `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Gender        string `json:"gender"`
	IsActive      bool   `json:"is_active"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Address 地址信息
type Address struct {
	Id              uint64 `json:"id"`              // 地址ID
	CustomerId      uint64 `json:"customer_id"`     // 客户ID
	FirstName       string `json:"first_name"`      // 名
	LastName        string `json:"last_name"`       // 姓
	Email           string `json:"email"`           // 邮箱
	Company         string `json:"company"`         // 公司
	CountryId       uint64 `json:"country_id"`      // 国家ID
	StateProvinceId uint64 `json:"state_province_id"` // 省/州ID
	City            string `json:"city"`            // 城市
	Address1        string `json:"address1"`        // 地址行1
	Address2        string `json:"address2"`        // 地址行2
	ZipPostalCode   string `json:"zip_postal_code"` // 邮编
	PhoneNumber     string `json:"phone_number"`    // 电话
	FaxNumber       string `json:"fax_number"`      // 传真
}

// Empty 空响应
type Empty struct{}

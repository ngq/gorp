// Package pb 客户服务 gRPC 接口定义
package customer

import (
	"context"
)

// CustomerServiceServer 客户服务服务端接口
type CustomerServiceServer interface {
	// ValidateCustomer 验证客户
	ValidateCustomer(ctx context.Context, req *ValidateCustomerRequest) (*ValidateCustomerResponse, error)

	// GetCustomer 获取客户信息
	GetCustomer(ctx context.Context, req *GetCustomerRequest) (*GetCustomerResponse, error)

	// GetCustomerAddresses 获取客户地址
	GetCustomerAddresses(ctx context.Context, req *GetCustomerAddressesRequest) (*GetCustomerAddressesResponse, error)
}

// ValidateCustomerRequest 验证客户请求
type ValidateCustomerRequest struct {
	CustomerID uint64
}

// ValidateCustomerResponse 验证客户响应
type ValidateCustomerResponse struct {
	Valid          bool
	ErrorMessage   string
	CustomerRoleID uint64
}

// GetCustomerRequest 获取客户请求
type GetCustomerRequest struct {
	CustomerID uint64
}

// GetCustomerResponse 获取客户响应
type GetCustomerResponse struct {
	ID        uint64
	Username  string
	Email     string
	Phone     string
	FirstName string
	LastName  string
	IsActive  bool
	RoleIDs   []uint64
}

// GetCustomerAddressesRequest 获取客户地址请求
type GetCustomerAddressesRequest struct {
	CustomerID uint64
}

// GetCustomerAddressesResponse 获取客户地址响应
type GetCustomerAddressesResponse struct {
	Addresses []*Address
}

// Address 地址
type Address struct {
	ID               uint64
	FirstName        string
	LastName         string
	Email            string
	Phone            string
	Country          string
	State            string
	City             string
	Address1         string
	Address2         string
	ZipCode          string
	IsDefaultBilling bool
	IsDefaultShipping bool
}

// UnimplementedCustomerServiceServer 未实现的服务端基类
type UnimplementedCustomerServiceServer struct{}

func (UnimplementedCustomerServiceServer) ValidateCustomer(ctx context.Context, req *ValidateCustomerRequest) (*ValidateCustomerResponse, error) {
	return nil, nil
}
func (UnimplementedCustomerServiceServer) GetCustomer(ctx context.Context, req *GetCustomerRequest) (*GetCustomerResponse, error) {
	return nil, nil
}
func (UnimplementedCustomerServiceServer) GetCustomerAddresses(ctx context.Context, req *GetCustomerAddressesRequest) (*GetCustomerAddressesResponse, error) {
	return nil, nil
}

// CustomerServiceClient 客户端接口
type CustomerServiceClient interface {
	ValidateCustomer(ctx context.Context, req *ValidateCustomerRequest) (*ValidateCustomerResponse, error)
	GetCustomer(ctx context.Context, req *GetCustomerRequest) (*GetCustomerResponse, error)
	GetCustomerAddresses(ctx context.Context, req *GetCustomerAddressesRequest) (*GetCustomerAddressesResponse, error)
}

// NewCustomerServiceClient 创建客户端
func NewCustomerServiceClient(conn interface{}) CustomerServiceClient {
	return &customerClient{}
}

type customerClient struct{}

func (c *customerClient) ValidateCustomer(ctx context.Context, req *ValidateCustomerRequest) (*ValidateCustomerResponse, error) {
	return nil, nil
}
func (c *customerClient) GetCustomer(ctx context.Context, req *GetCustomerRequest) (*GetCustomerResponse, error) {
	return nil, nil
}
func (c *customerClient) GetCustomerAddresses(ctx context.Context, req *GetCustomerAddressesRequest) (*GetCustomerAddressesResponse, error) {
	return nil, nil
}
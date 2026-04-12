// Package grpcsvc 客户服务 gRPC 实现
package grpcsvc

import (
	"context"

	"nop-go/shared/customer"
)

// CustomerGRPCServer 客户服务 gRPC 服务端
type CustomerGRPCServer struct {
	customer.UnimplementedCustomerServiceServer
}

// NewCustomerGRPCServer 创建客户 gRPC 服务端
func NewCustomerGRPCServer() *CustomerGRPCServer {
	return &CustomerGRPCServer{}
}

// ValidateCustomer 验证客户
func (s *CustomerGRPCServer) ValidateCustomer(ctx context.Context, req *customer.ValidateCustomerRequest) (*customer.ValidateCustomerResponse, error) {
	// TODO: 实现客户验证逻辑
	return &customer.ValidateCustomerResponse{
		Valid:          true,
		CustomerRoleID: 1,
	}, nil
}

// GetCustomer 获取客户信息
func (s *CustomerGRPCServer) GetCustomer(ctx context.Context, req *customer.GetCustomerRequest) (*customer.GetCustomerResponse, error) {
	// TODO: 实现获取客户信息
	return &customer.GetCustomerResponse{
		ID:        req.CustomerID,
		IsActive:  true,
	}, nil
}

// GetCustomerAddresses 获取客户地址
func (s *CustomerGRPCServer) GetCustomerAddresses(ctx context.Context, req *customer.GetCustomerAddressesRequest) (*customer.GetCustomerAddressesResponse, error) {
	// TODO: 实现获取客户地址
	return &customer.GetCustomerAddressesResponse{
		Addresses: []*customer.Address{},
	}, nil
}
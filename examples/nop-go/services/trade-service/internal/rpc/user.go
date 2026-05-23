// Package rpc 提供用户服务的 gRPC 客户端封装。
// trade-service 结算时需要通过此客户端调用 user-service 获取用户信息和收货地址。
package rpc

import (
	"context"

	userv1 "nop-go/services/user-service/api/user/v1"

	"google.golang.org/grpc"
)

// UserClient 用户服务 gRPC 客户端。
type UserClient struct {
	conn *grpc.ClientConn
	cli  userv1.UserRPCClient
}

// NewUserClient 创建用户服务 gRPC 客户端。
func NewUserClient(addr string, opts ...grpc.DialOption) (*UserClient, error) {
	conn, err := Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &UserClient{conn: conn, cli: userv1.NewUserRPCClient(conn)}, nil
}

// Close 关闭连接。
func (c *UserClient) Close() error {
	return c.conn.Close()
}

// GetUser 根据 ID 获取用户基本信息。
func (c *UserClient) GetUser(ctx context.Context, id uint32) (*userv1.GetUserResp, error) {
	return c.cli.GetUser(ctx, &userv1.GetUserReq{Id: id})
}

// ListUserAddresses 获取用户地址列表。
func (c *UserClient) ListUserAddresses(ctx context.Context, userID uint32) (*userv1.ListUserAddressesResp, error) {
	return c.cli.ListUserAddresses(ctx, &userv1.ListUserAddressesReq{UserId: userID})
}
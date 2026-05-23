// Package rpc 提供管理服务的 gRPC 客户端封装。
// trade-service 结算时需要通过此客户端调用 admin-service 获取折扣规则和权限校验。
package rpc

import (
	"context"

	adminv1 "nop-go/services/admin-service/api/admin/v1"

	"google.golang.org/grpc"
)

// AdminClient 管理服务 gRPC 客户端。
type AdminClient struct {
	conn *grpc.ClientConn
	cli  adminv1.AdminRPCClient
}

// NewAdminClient 创建管理服务 gRPC 客户端。
func NewAdminClient(addr string, opts ...grpc.DialOption) (*AdminClient, error) {
	conn, err := Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &AdminClient{conn: conn, cli: adminv1.NewAdminRPCClient(conn)}, nil
}

// Close 关闭连接。
func (c *AdminClient) Close() error {
	return c.conn.Close()
}

// GetDiscount 根据 ID 获取优惠规则。
func (c *AdminClient) GetDiscount(ctx context.Context, id uint32) (*adminv1.GetDiscountResp, error) {
	return c.cli.GetDiscount(ctx, &adminv1.GetDiscountReq{Id: id})
}

// GetDiscountByCode 根据优惠码获取优惠规则。
func (c *AdminClient) GetDiscountByCode(ctx context.Context, code string) (*adminv1.GetDiscountResp, error) {
	return c.cli.GetDiscountByCode(ctx, &adminv1.GetDiscountByCodeReq{Code: code})
}

// CheckPermission 检查权限。
func (c *AdminClient) CheckPermission(ctx context.Context, roleID uint32, resource, action string) (*adminv1.CheckPermissionResp, error) {
	return c.cli.CheckPermission(ctx, &adminv1.CheckPermissionReq{
		RoleId:   roleID,
		Resource: resource,
		Action:   action,
	})
}
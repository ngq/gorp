// Package grpc 实现 admin-service 的 gRPC 服务端。
// AdminRPCServer 实现 adminv1.UnimplementedAdminRPCServer 接口，
// 将 gRPC 请求委托给已有的 DiscountService 和 SecurityService 处理。
package grpc

import (
	"context"

	adminv1 "nop-go/services/admin-service/api/admin/v1"
	"nop-go/services/admin-service/internal/service"

	"google.golang.org/grpc"
)

// AdminRPCServer 管理服务 gRPC 服务端实现。
type AdminRPCServer struct {
	adminv1.UnimplementedAdminRPCServer
	discountSvc *service.DiscountService
	securitySvc *service.SecurityService
}

// NewAdminRPCServer 创建管理服务 gRPC 服务端。
func NewAdminRPCServer(discountSvc *service.DiscountService, securitySvc *service.SecurityService) *AdminRPCServer {
	return &AdminRPCServer{discountSvc: discountSvc, securitySvc: securitySvc}
}

// GetDiscount 根据 ID 获取优惠规则 —— 供 trade-service 结算时计算折扣。
func (s *AdminRPCServer) GetDiscount(ctx context.Context, req *adminv1.GetDiscountReq) (*adminv1.GetDiscountResp, error) {
	discount, err := s.discountSvc.GetDiscount(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &adminv1.GetDiscountResp{
		Id:            uint32(discount.ID),
		Name:          discount.Name,
		Code:          discount.Code,
		Type:          int32(discount.Type),
		Value:         discount.Value,
		MinAmount:     discount.MinAmount,
		MaxDiscount:   discount.MaxDiscount,
		StartTime:     discount.StartTime,
		EndTime:       discount.EndTime,
		TotalQuota:    int32(discount.TotalQuota),
		UsedQuota:     int32(discount.UsedQuota),
		PerUserLimit:  int32(discount.PerUserLimit),
		Status:        int32(discount.Status),
		Description:   discount.Description,
	}, nil
}

// GetDiscountByCode 根据优惠码获取优惠规则 —— 供 trade-service 结算时验证优惠码。
func (s *AdminRPCServer) GetDiscountByCode(ctx context.Context, req *adminv1.GetDiscountByCodeReq) (*adminv1.GetDiscountResp, error) {
	discount, err := s.discountSvc.GetDiscountByCode(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return &adminv1.GetDiscountResp{
		Id:            uint32(discount.ID),
		Name:          discount.Name,
		Code:          discount.Code,
		Type:          int32(discount.Type),
		Value:         discount.Value,
		MinAmount:     discount.MinAmount,
		MaxDiscount:   discount.MaxDiscount,
		StartTime:     discount.StartTime,
		EndTime:       discount.EndTime,
		TotalQuota:    int32(discount.TotalQuota),
		UsedQuota:     int32(discount.UsedQuota),
		PerUserLimit:  int32(discount.PerUserLimit),
		Status:        int32(discount.Status),
		Description:   discount.Description,
	}, nil
}

// CheckPermission 检查权限 —— 供 admin 权限检查和 gateway 接口鉴权。
func (s *AdminRPCServer) CheckPermission(ctx context.Context, req *adminv1.CheckPermissionReq) (*adminv1.CheckPermissionResp, error) {
	acls, err := s.securitySvc.GetACLsByRoleID(ctx, uint(req.GetRoleId()))
	if err != nil {
		return nil, err
	}
	// 在 ACL 列表中查找匹配的规则
	for _, acl := range acls {
		if acl.Resource == req.GetResource() && acl.Action == req.GetAction() {
			return &adminv1.CheckPermissionResp{
				Allowed: acl.Effect == "allow",
				Effect:  acl.Effect,
			}, nil
		}
	}
	// 默认拒绝
	return &adminv1.CheckPermissionResp{Allowed: false, Effect: "deny"}, nil
}

// RegisterAdminService 注册管理 gRPC 服务到 gRPC Server。
func RegisterAdminService(server *grpc.Server, discountSvc *service.DiscountService, securitySvc *service.SecurityService) {
	adminv1.RegisterAdminRPCServer(server, NewAdminRPCServer(discountSvc, securitySvc))
}
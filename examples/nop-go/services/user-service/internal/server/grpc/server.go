// Package grpc 实现 user-service 的 gRPC 服务端。
// UserRPCServer 实现 userv1.UnimplementedUserRPCServer 接口，
// 将 gRPC 请求委托给已有的 UserService 处理。
package grpc

import (
	"context"

	userv1 "nop-go/services/user-service/api/user/v1"
	"nop-go/services/user-service/internal/service"

	"google.golang.org/grpc"
)

// UserRPCServer 用户服务 gRPC 服务端实现。
type UserRPCServer struct {
	userv1.UnimplementedUserRPCServer
	userSvc *service.UserService
}

// NewUserRPCServer 创建用户服务 gRPC 服务端。
func NewUserRPCServer(userSvc *service.UserService) *UserRPCServer {
	return &UserRPCServer{userSvc: userSvc}
}

// GetUser 根据 ID 获取用户基本信息 —— 供 trade-service 和 admin-service 跨服务调用。
func (s *UserRPCServer) GetUser(ctx context.Context, req *userv1.GetUserReq) (*userv1.GetUserResp, error) {
	user, err := s.userSvc.GetUser(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &userv1.GetUserResp{
		Id:        uint32(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    int32(user.Status),
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ListUserAddresses 获取用户地址列表 —— 供 trade-service 结算时获取收货地址。
func (s *UserRPCServer) ListUserAddresses(ctx context.Context, req *userv1.ListUserAddressesReq) (*userv1.ListUserAddressesResp, error) {
	addrs, err := s.userSvc.ListAddresses(ctx, uint(req.GetUserId()))
	if err != nil {
		return nil, err
	}
	items := make([]*userv1.AddressItem, 0, len(addrs))
	for _, a := range addrs {
		items = append(items, &userv1.AddressItem{
			Id:            uint32(a.ID),
			UserId:        uint32(a.UserID),
			RecipientName: a.RecipientName,
			Phone:         a.Phone,
			Province:      a.Province,
			City:          a.City,
			District:      a.District,
			Detail:        a.Detail,
			IsDefault:     a.IsDefault,
		})
	}
	return &userv1.ListUserAddressesResp{Items: items}, nil
}

// RegisterUserService 注册用户 gRPC 服务到 gRPC Server。
func RegisterUserService(server *grpc.Server, userSvc *service.UserService) {
	userv1.RegisterUserRPCServer(server, NewUserRPCServer(userSvc))
}
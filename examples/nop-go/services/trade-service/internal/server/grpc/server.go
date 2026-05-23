// Package grpc 实现 trade-service 的 gRPC 服务端。
// TradeRPCServer 实现 tradev1.UnimplementedTradeRPCServer 接口，
// 将 gRPC 请求委托给已有的 OrderService 处理。
package grpc

import (
	"context"

	tradev1 "nop-go/services/trade-service/api/trade/v1"
	"nop-go/services/trade-service/internal/service"

	"google.golang.org/grpc"
)

// TradeRPCServer 交易服务 gRPC 服务端实现。
type TradeRPCServer struct {
	tradev1.UnimplementedTradeRPCServer
	orderSvc *service.OrderService
}

// NewTradeRPCServer 创建交易服务 gRPC 服务端。
func NewTradeRPCServer(orderSvc *service.OrderService) *TradeRPCServer {
	return &TradeRPCServer{orderSvc: orderSvc}
}

// GetOrder 根据 ID 获取订单信息 —— 供 gateway 或其他服务查询订单状态。
func (s *TradeRPCServer) GetOrder(ctx context.Context, req *tradev1.GetOrderReq) (*tradev1.GetOrderResp, error) {
	order, err := s.orderSvc.OrderUC.GetOrder(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &tradev1.GetOrderResp{
		Id:            order.ID,
		UserId:        order.UserID,
		Status:        order.Status,
		TotalAmount:   order.TotalAmount,
		Currency:      order.Currency,
		ShippingAddr:  order.ShippingAddr,
		PaymentMethod: order.PaymentMethod,
		CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     order.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// RegisterTradeService 注册交易 gRPC 服务到 gRPC Server。
func RegisterTradeService(server *grpc.Server, orderSvc *service.OrderService) {
	tradev1.RegisterTradeRPCServer(server, NewTradeRPCServer(orderSvc))
}
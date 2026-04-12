// Package grpcsvc 价格服务 gRPC 实现
package grpcsvc

import (
	"context"

	"nop-go/shared/price"
)

// PriceGRPCServer 价格服务 gRPC 服务端
type PriceGRPCServer struct {
	price.UnimplementedPriceServiceServer
}

// NewPriceGRPCServer 创建价格 gRPC 服务端
func NewPriceGRPCServer() *PriceGRPCServer {
	return &PriceGRPCServer{}
}

// CalculateOrderPrice 计算订单价格
func (s *PriceGRPCServer) CalculateOrderPrice(ctx context.Context, req *price.CalculateOrderPriceRequest) (*price.CalculateOrderPriceResponse, error) {
	// TODO: 实现价格计算逻辑
	var subtotal float64
	for _, item := range req.Items {
		subtotal += 100.0 * float64(item.Quantity) // 简化: 每个商品 100 元
	}

	return &price.CalculateOrderPriceResponse{
		Subtotal:  subtotal,
		Total:     subtotal,
		Currency:  "CNY",
	}, nil
}

// GetProductPrice 获取商品价格
func (s *PriceGRPCServer) GetProductPrice(ctx context.Context, req *price.GetProductPriceRequest) (*price.GetProductPriceResponse, error) {
	// TODO: 实现获取商品价格
	return &price.GetProductPriceResponse{
		ProductID:  req.ProductID,
		UnitPrice:  100.0,
		FinalPrice: 100.0,
	}, nil
}

// ValidateCoupon 验证优惠券
func (s *PriceGRPCServer) ValidateCoupon(ctx context.Context, req *price.ValidateCouponRequest) (*price.ValidateCouponResponse, error) {
	// TODO: 实现优惠券验证
	return &price.ValidateCouponResponse{
		Valid: false,
	}, nil
}
// Package pb 价格服务 gRPC 接口定义
package price

import (
	"context"
)

// PriceServiceServer 价格服务服务端接口
type PriceServiceServer interface {
	// CalculateOrderPrice 计算订单总价
	CalculateOrderPrice(ctx context.Context, req *CalculateOrderPriceRequest) (*CalculateOrderPriceResponse, error)

	// GetProductPrice 获取商品价格
	GetProductPrice(ctx context.Context, req *GetProductPriceRequest) (*GetProductPriceResponse, error)

	// ValidateCoupon 验证优惠券
	ValidateCoupon(ctx context.Context, req *ValidateCouponRequest) (*ValidateCouponResponse, error)
}

// CalculateOrderPriceRequest 计算订单价格请求
type CalculateOrderPriceRequest struct {
	CustomerID    uint64
	CustomerRoleID uint64
	Items         []*OrderItem
	CouponCode    string
	RewardPoints  int32   // 使用积分
	ShippingAddress *Address
}

// OrderItem 订单商品项
type OrderItem struct {
	ProductID  uint64
	SKU        string
	Quantity   int32
	Attributes map[string]string // 商品属性
}

// Address 地址
type Address struct {
	Country string
	State   string
	City    string
	ZipCode string
}

// CalculateOrderPriceResponse 计算订单价格响应
type CalculateOrderPriceResponse struct {
	Subtotal            float64           // 商品小计
	DiscountAmount      float64           // 折扣金额
	TaxAmount           float64           // 税费
	ShippingAmount      float64           // 运费
	RewardPointsDiscount float64          // 积分抵扣
	Total               float64           // 总价
	Currency            string
	Breakdown           []*PriceBreakdown
}

// PriceBreakdown 价格明细
type PriceBreakdown struct {
	ProductID   uint64
	ProductName string
	Quantity    int32
	UnitPrice   float64
	Subtotal    float64
	Discount    float64
	Tax         float64
	Total       float64
}

// GetProductPriceRequest 获取商品价格请求
type GetProductPriceRequest struct {
	ProductID      uint64
	CustomerRoleID uint64
	Quantity       int32
}

// GetProductPriceResponse 获取商品价格响应
type GetProductPriceResponse struct {
	ProductID   uint64
	UnitPrice   float64
	FinalPrice  float64 // 优惠后价格
	TaxRate     float64
	OnSale      bool
}

// ValidateCouponRequest 验证优惠券请求
type ValidateCouponRequest struct {
	CouponCode   string
	CustomerID   uint64
	OrderAmount  float64
	ProductIDs   []uint64
}

// ValidateCouponResponse 验证优惠券响应
type ValidateCouponResponse struct {
	Valid          bool
	ErrorMessage   string
	DiscountAmount float64
	DiscountType   string // percentage, fixed
}

// UnimplementedPriceServiceServer 未实现的服务端基类
type UnimplementedPriceServiceServer struct{}

func (UnimplementedPriceServiceServer) CalculateOrderPrice(ctx context.Context, req *CalculateOrderPriceRequest) (*CalculateOrderPriceResponse, error) {
	return nil, nil
}
func (UnimplementedPriceServiceServer) GetProductPrice(ctx context.Context, req *GetProductPriceRequest) (*GetProductPriceResponse, error) {
	return nil, nil
}
func (UnimplementedPriceServiceServer) ValidateCoupon(ctx context.Context, req *ValidateCouponRequest) (*ValidateCouponResponse, error) {
	return nil, nil
}

// PriceServiceClient 客户端接口
type PriceServiceClient interface {
	CalculateOrderPrice(ctx context.Context, req *CalculateOrderPriceRequest) (*CalculateOrderPriceResponse, error)
	GetProductPrice(ctx context.Context, req *GetProductPriceRequest) (*GetProductPriceResponse, error)
	ValidateCoupon(ctx context.Context, req *ValidateCouponRequest) (*ValidateCouponResponse, error)
}

// NewPriceServiceClient 创建客户端
func NewPriceServiceClient(conn interface{}) PriceServiceClient {
	return &priceClient{}
}

type priceClient struct{}

func (c *priceClient) CalculateOrderPrice(ctx context.Context, req *CalculateOrderPriceRequest) (*CalculateOrderPriceResponse, error) {
	return nil, nil
}
func (c *priceClient) GetProductPrice(ctx context.Context, req *GetProductPriceRequest) (*GetProductPriceResponse, error) {
	return nil, nil
}
func (c *priceClient) ValidateCoupon(ctx context.Context, req *ValidateCouponRequest) (*ValidateCouponResponse, error) {
	return nil, nil
}
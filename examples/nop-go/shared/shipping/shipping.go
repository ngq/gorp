// Package pb 配送服务 gRPC 接口定义
package shipping

import (
	"context"
)

// ShippingServiceServer 配送服务服务端接口
type ShippingServiceServer interface {
	// CalculateShipping 计算运费
	CalculateShipping(ctx context.Context, req *CalculateShippingRequest) (*CalculateShippingResponse, error)

	// CreateShipment 创建运单
	CreateShipment(ctx context.Context, req *CreateShipmentRequest) (*CreateShipmentResponse, error)

	// GetTrackingInfo 获取物流追踪
	GetTrackingInfo(ctx context.Context, req *GetTrackingInfoRequest) (*GetTrackingInfoResponse, error)

	// CancelShipment 取消运单
	CancelShipment(ctx context.Context, req *CancelShipmentRequest) (*CancelShipmentResponse, error)
}

// CalculateShippingRequest 计算运费请求
type CalculateShippingRequest struct {
	OrderID      uint64
	FromAddress  *Address
	ToAddress    *Address
	Packages     []*Package
	Currency     string
}

// Address 地址
type Address struct {
	Country  string
	State    string
	City     string
	Address1 string
	Address2 string
	ZipCode  string
	Phone    string
	Name     string
}

// Package 包裹
type Package struct {
	Weight  float64 // kg
	Length  float64 // cm
	Width   float64
	Height  float64
	Value   float64
	Quantity int32
}

// CalculateShippingResponse 计算运费响应
type CalculateShippingResponse struct {
	Options      []*ShippingOption
	ErrorMessage string
}

// ShippingOption 配送选项
type ShippingOption struct {
	Code          string
	Name          string
	Amount        float64
	Currency      string
	EstimatedDays int32
	Description   string
}

// CreateShipmentRequest 创建运单请求
type CreateShipmentRequest struct {
	OrderID             uint64
	ShippingMethodCode  string
	FromAddress         *Address
	ToAddress           *Address
	Packages            []*Package
}

// CreateShipmentResponse 创建运单响应
type CreateShipmentResponse struct {
	Success        bool
	ShipmentID     uint64
	TrackingNumber string
	LabelURL       string
	ErrorMessage   string
}

// GetTrackingInfoRequest 获取物流追踪请求
type GetTrackingInfoRequest struct {
	TrackingNumber string
}

// GetTrackingInfoResponse 获取物流追踪响应
type GetTrackingInfoResponse struct {
	TrackingNumber    string
	Status            string
	Events            []*TrackingEvent
	EstimatedDelivery string
}

// TrackingEvent 物流事件
type TrackingEvent struct {
	Time        string
	Location    string
	Description string
	Status      string
}

// CancelShipmentRequest 取消运单请求
type CancelShipmentRequest struct {
	ShipmentID uint64
}

// CancelShipmentResponse 取消运单响应
type CancelShipmentResponse struct {
	Success      bool
	ErrorMessage string
}

// UnimplementedShippingServiceServer 未实现的服务端基类
type UnimplementedShippingServiceServer struct{}

func (UnimplementedShippingServiceServer) CalculateShipping(ctx context.Context, req *CalculateShippingRequest) (*CalculateShippingResponse, error) {
	return nil, nil
}
func (UnimplementedShippingServiceServer) CreateShipment(ctx context.Context, req *CreateShipmentRequest) (*CreateShipmentResponse, error) {
	return nil, nil
}
func (UnimplementedShippingServiceServer) GetTrackingInfo(ctx context.Context, req *GetTrackingInfoRequest) (*GetTrackingInfoResponse, error) {
	return nil, nil
}
func (UnimplementedShippingServiceServer) CancelShipment(ctx context.Context, req *CancelShipmentRequest) (*CancelShipmentResponse, error) {
	return nil, nil
}

// ShippingServiceClient 客户端接口
type ShippingServiceClient interface {
	CalculateShipping(ctx context.Context, req *CalculateShippingRequest) (*CalculateShippingResponse, error)
	CreateShipment(ctx context.Context, req *CreateShipmentRequest) (*CreateShipmentResponse, error)
	GetTrackingInfo(ctx context.Context, req *GetTrackingInfoRequest) (*GetTrackingInfoResponse, error)
	CancelShipment(ctx context.Context, req *CancelShipmentRequest) (*CancelShipmentResponse, error)
}

// NewShippingServiceClient 创建客户端
func NewShippingServiceClient(conn interface{}) ShippingServiceClient {
	return &shippingClient{}
}

type shippingClient struct{}

func (c *shippingClient) CalculateShipping(ctx context.Context, req *CalculateShippingRequest) (*CalculateShippingResponse, error) {
	return nil, nil
}
func (c *shippingClient) CreateShipment(ctx context.Context, req *CreateShipmentRequest) (*CreateShipmentResponse, error) {
	return nil, nil
}
func (c *shippingClient) GetTrackingInfo(ctx context.Context, req *GetTrackingInfoRequest) (*GetTrackingInfoResponse, error) {
	return nil, nil
}
func (c *shippingClient) CancelShipment(ctx context.Context, req *CancelShipmentRequest) (*CancelShipmentResponse, error) {
	return nil, nil
}
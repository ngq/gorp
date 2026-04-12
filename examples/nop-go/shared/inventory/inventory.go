// Package pb 库存服务 gRPC 接口定义
//
// 中文说明:
// - 定义库存服务的 gRPC 接口;
// - 用于服务间同步调用;
// - 支持 DTM SAGA 分布式事务。
package inventory

import (
	"context"
)

// InventoryServiceServer 库存服务服务端接口
type InventoryServiceServer interface {
	// GetStock 获取商品库存
	GetStock(ctx context.Context, req *GetStockRequest) (*GetStockResponse, error)

	// ReserveStock 预留库存
	//
	// 中文说明:
	// - SAGA 事务第一步: 预留库存;
	// - 返回预留 ID,用于后续确认或释放;
	// - Compensate: ReleaseStock。
	ReserveStock(ctx context.Context, req *ReserveStockRequest) (*ReserveStockResponse, error)

	// ConfirmStock 确认库存扣减
	ConfirmStock(ctx context.Context, req *ConfirmStockRequest) (*ConfirmStockResponse, error)

	// ReleaseStock 释放预留库存
	//
	// 中文说明:
	// - SAGA 补偿操作;
	// - 支付失败或订单取消时调用。
	ReleaseStock(ctx context.Context, req *ReleaseStockRequest) (*ReleaseStockResponse, error)

	// CheckAvailability 检查库存是否充足
	CheckAvailability(ctx context.Context, req *CheckAvailabilityRequest) (*CheckAvailabilityResponse, error)
}

// GetStockRequest 获取库存请求
type GetStockRequest struct {
	ProductID   uint64
	WarehouseID uint64 // 0 表示查询所有仓库总量
}

// GetStockResponse 获取库存响应
type GetStockResponse struct {
	ProductID        uint64
	WarehouseID      uint64
	Quantity         int32
	ReservedQuantity int32
	AvailableQuantity int32
}

// ReserveStockRequest 预留库存请求
type ReserveStockRequest struct {
	OrderID       uint64
	ProductID     uint64
	WarehouseID   uint64
	Quantity      int32
	ReservationID string // DTM 事务 ID
}

// ReserveStockResponse 预留库存响应
type ReserveStockResponse struct {
	Success       bool
	ReservationID string
	ErrorMessage  string
}

// ConfirmStockRequest 确认库存请求
type ConfirmStockRequest struct {
	OrderID       uint64
	ReservationID string
}

// ConfirmStockResponse 确认库存响应
type ConfirmStockResponse struct {
	Success      bool
	ErrorMessage string
}

// ReleaseStockRequest 释放库存请求
type ReleaseStockRequest struct {
	OrderID       uint64
	ReservationID string
}

// ReleaseStockResponse 释放库存响应
type ReleaseStockResponse struct {
	Success      bool
	ErrorMessage string
}

// CheckAvailabilityRequest 检查库存请求
type CheckAvailabilityRequest struct {
	Items []*StockItem
}

// StockItem 库存项
type StockItem struct {
	ProductID   uint64
	WarehouseID uint64
	Quantity    int32
}

// CheckAvailabilityResponse 检查库存响应
type CheckAvailabilityResponse struct {
	AllAvailable bool
	Results      []*StockCheckResult
}

// StockCheckResult 库存检查结果
type StockCheckResult struct {
	ProductID         uint64
	Available         bool
	AvailableQuantity int32
	RequestedQuantity int32
}

// UnimplementedInventoryServiceServer 未实现的服务端基类
//
// 中文说明:
// - 用于向前兼容;
// - 新增方法时不会破坏现有实现。
type UnimplementedInventoryServiceServer struct{}

func (UnimplementedInventoryServiceServer) GetStock(ctx context.Context, req *GetStockRequest) (*GetStockResponse, error) {
	return nil, nil
}
func (UnimplementedInventoryServiceServer) ReserveStock(ctx context.Context, req *ReserveStockRequest) (*ReserveStockResponse, error) {
	return nil, nil
}
func (UnimplementedInventoryServiceServer) ConfirmStock(ctx context.Context, req *ConfirmStockRequest) (*ConfirmStockResponse, error) {
	return nil, nil
}
func (UnimplementedInventoryServiceServer) ReleaseStock(ctx context.Context, req *ReleaseStockRequest) (*ReleaseStockResponse, error) {
	return nil, nil
}
func (UnimplementedInventoryServiceServer) CheckAvailability(ctx context.Context, req *CheckAvailabilityRequest) (*CheckAvailabilityResponse, error) {
	return nil, nil
}

// InventoryServiceClient 客户端接口
type InventoryServiceClient interface {
	GetStock(ctx context.Context, req *GetStockRequest) (*GetStockResponse, error)
	ReserveStock(ctx context.Context, req *ReserveStockRequest) (*ReserveStockResponse, error)
	ConfirmStock(ctx context.Context, req *ConfirmStockRequest) (*ConfirmStockResponse, error)
	ReleaseStock(ctx context.Context, req *ReleaseStockRequest) (*ReleaseStockResponse, error)
	CheckAvailability(ctx context.Context, req *CheckAvailabilityRequest) (*CheckAvailabilityResponse, error)
}

// NewInventoryServiceClient 创建客户端
func NewInventoryServiceClient(conn interface{}) InventoryServiceClient {
	return &inventoryClient{conn: conn}
}

type inventoryClient struct {
	conn interface{}
}

func (c *inventoryClient) GetStock(ctx context.Context, req *GetStockRequest) (*GetStockResponse, error) {
	return nil, nil
}
func (c *inventoryClient) ReserveStock(ctx context.Context, req *ReserveStockRequest) (*ReserveStockResponse, error) {
	return nil, nil
}
func (c *inventoryClient) ConfirmStock(ctx context.Context, req *ConfirmStockRequest) (*ConfirmStockResponse, error) {
	return nil, nil
}
func (c *inventoryClient) ReleaseStock(ctx context.Context, req *ReleaseStockRequest) (*ReleaseStockResponse, error) {
	return nil, nil
}
func (c *inventoryClient) CheckAvailability(ctx context.Context, req *CheckAvailabilityRequest) (*CheckAvailabilityResponse, error) {
	return nil, nil
}
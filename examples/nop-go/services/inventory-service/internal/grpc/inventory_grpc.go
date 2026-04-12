// Package grpcsvc gRPC 服务端实现
//
// 中文说明:
// - 实现各服务的 gRPC 服务端;
// - 注册到 gorp 框架的 gRPC Provider;
// - 支持 DTM SAGA 分布式事务。
package grpcsvc

import (
	"context"

	"nop-go/shared/inventory"
	"nop-go/services/inventory-service/internal/biz"
)

// InventoryGRPCServer 库存服务 gRPC 服务端
//
// 中文说明:
// - 实现 inventory.InventoryServiceServer 接口;
// - 调用 UseCase 处理业务逻辑;
// - 支持 DTM SAGA 事务的 Try/Confirm/Cancel。
type InventoryGRPCServer struct {
	inventory.UnimplementedInventoryServiceServer
	uc *biz.InventoryUseCase
}

// NewInventoryGRPCServer 创建库存 gRPC 服务端
func NewInventoryGRPCServer(uc *biz.InventoryUseCase) *InventoryGRPCServer {
	return &InventoryGRPCServer{uc: uc}
}

// GetStock 获取库存
func (s *InventoryGRPCServer) GetStock(ctx context.Context, req *inventory.GetStockRequest) (*inventory.GetStockResponse, error) {
	stock, err := s.uc.GetByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if req.WarehouseID > 0 {
		// 查找指定仓库
		for _, inv := range stock {
			if inv.WarehouseID == req.WarehouseID {
				return &inventory.GetStockResponse{
					ProductID:         inv.ProductID,
					WarehouseID:       inv.WarehouseID,
					Quantity:          int32(inv.Quantity),
					ReservedQuantity:  int32(inv.ReservedQuantity),
					AvailableQuantity: int32(inv.Quantity - inv.ReservedQuantity),
				}, nil
			}
		}
		return nil, err
	}

	// 汇总所有仓库
	var totalQty, totalReserved int32
	for _, inv := range stock {
		totalQty += int32(inv.Quantity)
		totalReserved += int32(inv.ReservedQuantity)
	}

	return &inventory.GetStockResponse{
		ProductID:         req.ProductID,
		WarehouseID:       0,
		Quantity:          totalQty,
		ReservedQuantity:  totalReserved,
		AvailableQuantity: totalQty - totalReserved,
	}, nil
}

// ReserveStock 预留库存
//
// 中文说明:
// - SAGA Try 操作;
// - 预留库存,30分钟内未确认则自动释放;
// - 返回 reservation_id 用于后续确认/释放。
func (s *InventoryGRPCServer) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
	err := s.uc.ReserveStock(ctx, &biz.ReserveStockRequest{
		OrderID:     req.OrderID,
		ProductID:   req.ProductID,
		WarehouseID: req.WarehouseID,
		Quantity:    int(req.Quantity),
	})

	if err != nil {
		return &inventory.ReserveStockResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &inventory.ReserveStockResponse{
		Success:       true,
		ReservationID: req.ReservationID,
	}, nil
}

// ConfirmStock 确认库存扣减
//
// 中文说明:
// - SAGA Confirm 操作;
// - 支付成功后调用,确认扣减库存。
func (s *InventoryGRPCServer) ConfirmStock(ctx context.Context, req *inventory.ConfirmStockRequest) (*inventory.ConfirmStockResponse, error) {
	err := s.uc.ConfirmStock(ctx, req.OrderID)
	if err != nil {
		return &inventory.ConfirmStockResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &inventory.ConfirmStockResponse{Success: true}, nil
}

// ReleaseStock 释放预留库存
//
// 中文说明:
// - SAGA Cancel 操作;
// - 支付失败或订单取消时调用。
func (s *InventoryGRPCServer) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
	err := s.uc.ReleaseStock(ctx, req.OrderID)
	if err != nil {
		return &inventory.ReleaseStockResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &inventory.ReleaseStockResponse{Success: true}, nil
}

// CheckAvailability 检查库存是否充足
func (s *InventoryGRPCServer) CheckAvailability(ctx context.Context, req *inventory.CheckAvailabilityRequest) (*inventory.CheckAvailabilityResponse, error) {
	results := make([]*inventory.StockCheckResult, 0, len(req.Items))
	allAvailable := true

	for _, item := range req.Items {
		stock, err := s.uc.GetByProductID(ctx, item.ProductID)
		if err != nil {
			results = append(results, &inventory.StockCheckResult{
				ProductID:         item.ProductID,
				Available:         false,
				AvailableQuantity: 0,
				RequestedQuantity: item.Quantity,
			})
			allAvailable = false
			continue
		}

		var availableQty int32
		for _, inv := range stock {
			if item.WarehouseID == 0 || inv.WarehouseID == item.WarehouseID {
				availableQty += int32(inv.Quantity - inv.ReservedQuantity)
			}
		}

		available := availableQty >= item.Quantity
		if !available {
			allAvailable = false
		}

		results = append(results, &inventory.StockCheckResult{
			ProductID:         item.ProductID,
			Available:         available,
			AvailableQuantity: availableQty,
			RequestedQuantity: item.Quantity,
		})
	}

	return &inventory.CheckAvailabilityResponse{
		AllAvailable: allAvailable,
		Results:      results,
	}, nil
}
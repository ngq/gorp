// Package biz 库存服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"time"

	"nop-go/services/inventory-service/internal/data"
	"nop-go/services/inventory-service/internal/models"
	"nop-go/shared/dlock"
	shareErrors "nop-go/shared/errors"
)

// ReserveStockRequest 预留库存请求
type ReserveStockRequest struct {
	OrderID     uint64
	ProductID   uint64
	WarehouseID uint64
	Quantity    int
}

// InventoryUseCase 库存服务用例
type InventoryUseCase struct {
	inventoryRepo    data.InventoryRepository
	warehouseRepo    data.WarehouseRepository
	reservationRepo  data.StockReservationRepository
	logRepo          data.InventoryLogRepository
	lockMgr          *dlock.LockManager
}

// NewInventoryUseCase 创建库存用例
func NewInventoryUseCase(
	inventoryRepo data.InventoryRepository,
	warehouseRepo data.WarehouseRepository,
	reservationRepo data.StockReservationRepository,
	logRepo data.InventoryLogRepository,
	lockMgr *dlock.LockManager,
) *InventoryUseCase {
	return &InventoryUseCase{
		inventoryRepo:    inventoryRepo,
		warehouseRepo:    warehouseRepo,
		reservationRepo:  reservationRepo,
		logRepo:          logRepo,
		lockMgr:          lockMgr,
	}
}

// ReserveStock 预留库存
//
// 中文说明:
// - 使用分布式锁保证并发安全;
// - SAGA Try 操作;
// - 预留库存,30分钟内未确认则自动释放。
func (uc *InventoryUseCase) ReserveStock(ctx context.Context, req *ReserveStockRequest) error {
	// 获取分布式锁
	//
	// 中文说明:
	// - 防止高并发场景下的超卖;
	// - 锁粒度为 商品+仓库;
	// - 5秒超时。
	lock, err := dlock.AcquireInventoryLock(ctx, uc.lockMgr, req.ProductID, req.WarehouseID, 30*time.Second)
	if err != nil {
		return errors.New("failed to acquire inventory lock")
	}
	defer lock.Release(ctx)

	// 检查库存是否充足
	inv, err := uc.inventoryRepo.GetByProductAndWarehouse(ctx, req.ProductID, req.WarehouseID)
	if err != nil {
		return shareErrors.ErrInventoryNotFound
	}

	availableQty := inv.Quantity - inv.ReservedQuantity
	if availableQty < req.Quantity {
		return shareErrors.ErrInsufficientStock
	}

	// 预留库存
	//
	// 中文说明:
	// - 增加预留数量;
	// - 创建预留记录;
	// - 30分钟后自动过期。
	err = uc.inventoryRepo.ReserveStock(ctx, req.ProductID, req.WarehouseID, req.Quantity)
	if err != nil {
		return shareErrors.ErrInsufficientStock
	}

	reservation := &models.StockReservation{
		OrderID:     req.OrderID,
		ProductID:   req.ProductID,
		WarehouseID: req.WarehouseID,
		Quantity:    req.Quantity,
		Status:      "reserved",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}

	return uc.reservationRepo.Create(ctx, reservation)
}

// ConfirmStock 确认库存扣减
//
// 中文说明:
// - 支付成功后调用;
// - 确认扣减库存;
// - 更新预留记录状态。
func (uc *InventoryUseCase) ConfirmStock(ctx context.Context, orderID uint64) error {
	reservation, err := uc.reservationRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// 获取分布式锁
	lock, err := dlock.AcquireInventoryLock(ctx, uc.lockMgr, reservation.ProductID, reservation.WarehouseID, 30*time.Second)
	if err != nil {
		return errors.New("failed to acquire inventory lock")
	}
	defer lock.Release(ctx)

	// 确认预留
	if err := uc.reservationRepo.Confirm(ctx, orderID); err != nil {
		return err
	}

	// 更新库存数量
	return uc.inventoryRepo.UpdateQuantity(ctx, reservation.ProductID, reservation.WarehouseID, -reservation.Quantity)
}

// ReleaseStock 释放预留库存
//
// 中文说明:
// - SAGA Cancel 操作;
// - 支付失败或订单取消时调用;
// - 释放预留的库存。
func (uc *InventoryUseCase) ReleaseStock(ctx context.Context, orderID uint64) error {
	reservation, err := uc.reservationRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// 获取分布式锁
	lock, err := dlock.AcquireInventoryLock(ctx, uc.lockMgr, reservation.ProductID, reservation.WarehouseID, 30*time.Second)
	if err != nil {
		return errors.New("failed to acquire inventory lock")
	}
	defer lock.Release(ctx)

	// 释放预留状态
	if err := uc.reservationRepo.Release(ctx, orderID); err != nil {
		return err
	}

	// 释放预留数量
	return uc.inventoryRepo.ReleaseStock(ctx, reservation.ProductID, reservation.WarehouseID, reservation.Quantity)
}

// GetInventory 获取库存
func (uc *InventoryUseCase) GetInventory(ctx context.Context, id uint64) (*models.Inventory, error) {
	inv, err := uc.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrInventoryNotFound
	}
	return inv, nil
}

// GetByProductID 按商品获取库存
func (uc *InventoryUseCase) GetByProductID(ctx context.Context, productID uint64) ([]*models.Inventory, error) {
	return uc.inventoryRepo.GetByProductID(ctx, productID)
}

// AdjustStock 调整库存
func (uc *InventoryUseCase) AdjustStock(ctx context.Context, req *models.AdjustStockRequest) error {
	// 获取分布式锁
	lock, err := dlock.AcquireInventoryLock(ctx, uc.lockMgr, req.ProductID, req.WarehouseID, 30*time.Second)
	if err != nil {
		return errors.New("failed to acquire inventory lock")
	}
	defer lock.Release(ctx)

	if err := uc.inventoryRepo.UpdateQuantity(ctx, req.ProductID, req.WarehouseID, req.Quantity); err != nil {
		return err
	}

	changeType := "in"
	if req.Quantity < 0 {
		changeType = "out"
	}

	log := &models.InventoryLog{
		ProductID:     req.ProductID,
		WarehouseID:   req.WarehouseID,
		ChangeType:    changeType,
		Quantity:      req.Quantity,
		ReferenceType: "adjustment",
		Remark:        req.Remark,
	}

	return uc.logRepo.Create(ctx, log)
}

// WarehouseUseCase 仓库用例
type WarehouseUseCase struct {
	warehouseRepo data.WarehouseRepository
}

// NewWarehouseUseCase 创建仓库用例
func NewWarehouseUseCase(warehouseRepo data.WarehouseRepository) *WarehouseUseCase {
	return &WarehouseUseCase{warehouseRepo: warehouseRepo}
}

// CreateWarehouse 创建仓库
func (uc *WarehouseUseCase) CreateWarehouse(ctx context.Context, w *models.Warehouse) error {
	return uc.warehouseRepo.Create(ctx, w)
}

// GetWarehouse 获取仓库
func (uc *WarehouseUseCase) GetWarehouse(ctx context.Context, id uint64) (*models.Warehouse, error) {
	return uc.warehouseRepo.GetByID(ctx, id)
}

// ListWarehouses 列出仓库
func (uc *WarehouseUseCase) ListWarehouses(ctx context.Context) ([]*models.Warehouse, error) {
	return uc.warehouseRepo.List(ctx)
}

// UpdateWarehouse 更新仓库
func (uc *WarehouseUseCase) UpdateWarehouse(ctx context.Context, w *models.Warehouse) error {
	return uc.warehouseRepo.Update(ctx, w)
}

// DeleteWarehouse 删除仓库
func (uc *WarehouseUseCase) DeleteWarehouse(ctx context.Context, id uint64) error {
	return uc.warehouseRepo.Delete(ctx, id)
}
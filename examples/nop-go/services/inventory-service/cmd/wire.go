//go:build wireinject

package main

import (
	"nop-go/shared/dlock"
	"nop-go/services/inventory-service/internal/biz"
	"nop-go/services/inventory-service/internal/data"
	"nop-go/services/inventory-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireInventoryService(db *gorm.DB, lockMgr *dlock.LockManager) (*service.InventoryService, error) {
	panic(wire.Build(
		data.NewInventoryRepository,
		data.NewWarehouseRepository,
		data.NewStockReservationRepository,
		data.NewInventoryLogRepository,
		biz.NewInventoryUseCase,
		biz.NewWarehouseUseCase,
		service.NewInventoryService,
	))
}

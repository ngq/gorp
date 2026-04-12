// Package main 库存服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/shared/dlock"
	"nop-go/services/inventory-service/internal/models"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("inventory-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	inventoryService, err := wireInventoryService(rt.DB, initLockManager(rt))
	if err != nil {
		return err
	}

	inventoryService.RegisterRoutes(rt.Engine)
	return nil
}

func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Inventory{},
		&models.Warehouse{},
		&models.InventoryLog{},
		&models.StockReservation{},
		&models.TierPrice{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}

func initLockManager(rt *bootstrap.HTTPServiceRuntime) *dlock.LockManager {
	// 使用框架的分布式锁能力
	locker, err := container.MakeAppService[contract.DistributedLock](rt.Container, contract.DistributedLockKey)
	if err != nil {
		rt.Logger.Info("分布式锁未配置，使用 noop 实现")
		// 返回 nil，业务层会处理
		return nil
	}
	rt.Logger.Info("分布式锁管理器初始化完成")
	return dlock.NewLockManager(locker)
}

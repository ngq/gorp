// Package main 搴撳瓨鏈嶅姟鍏ュ彛
package main

import (
	"fmt"
	"os"

	"nop-go/services/inventory-service/internal/models"
	"nop-go/shared/bootstrap"
	"nop-go/shared/dlock"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
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
		return fmt.Errorf("琛ㄧ粨鏋勮縼绉诲け璐? %w", err)
	}
	return nil
}

func initLockManager(rt *bootstrap.HTTPServiceRuntime) *dlock.LockManager {
	// 浣跨敤妗嗘灦鐨勫垎甯冨紡閿佽兘鍔?
	locker, err := container.MakeAppService[datacontract.DistributedLock](rt.Container, datacontract.DistributedLockKey)
	if err != nil {
		rt.Logger.Info("鍒嗗竷寮忛攣鏈厤缃紝浣跨敤 noop 瀹炵幇")
		// 杩斿洖 nil锛屼笟鍔″眰浼氬鐞?
		return nil
	}
	rt.Logger.Info("鍒嗗竷寮忛攣绠＄悊鍣ㄥ垵濮嬪寲瀹屾垚")
	return dlock.NewLockManager(locker)
}

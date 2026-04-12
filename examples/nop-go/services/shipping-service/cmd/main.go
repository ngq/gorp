// Package main 物流配送服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/shipping-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("shipping-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	shippingService, err := wireShippingService(rt.DB)
	if err != nil {
		return err
	}

	shippingService.RegisterRoutes(rt.Engine)
	return nil
}

func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Shipment{},
		&models.ShipmentItem{},
		&models.ShippingMethod{},
		&models.ShipmentTracking{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}

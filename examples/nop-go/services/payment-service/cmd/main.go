// Package main 支付服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/shared/plugin"
	"nop-go/services/payment-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("payment-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	paymentService, err := wirePaymentService(rt.DB, plugin.GetRegistry())
	if err != nil {
		return err
	}

	paymentService.RegisterRoutes(rt.Engine)
	return nil
}

func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Payment{},
		&models.PaymentTransaction{},
		&models.Refund{},
		&models.PaymentMethod{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}

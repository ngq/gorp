// Package main 联盟推广服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/affiliate-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("affiliate-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	config := affiliateConfigFromRuntime(rt)
	affService, err := wireAffiliateService(rt.DB, bootstrap.MustMakeJWTService(rt.Container), config)
	if err != nil {
		return err
	}

	affService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Affiliate{},
		&models.AffiliateOrder{},
		&models.AffiliateReferral{},
		&models.AffiliateCommission{},
		&models.AffiliatePayout{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}
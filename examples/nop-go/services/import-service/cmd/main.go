// Package main 导入导出服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/import-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("import-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	config := importConfigFromRuntime(rt)
	importService, err := wireImportService(rt.DB, bootstrap.MustMakeJWTService(rt.Container), config)
	if err != nil {
		return err
	}

	importService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.ImportProfile{},
		&models.ExportProfile{},
		&models.ImportHistory{},
		&models.ExportHistory{},
		&models.ImportError{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}
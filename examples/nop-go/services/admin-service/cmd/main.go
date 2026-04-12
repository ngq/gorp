// Package main 后台管理服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/admin-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("admin-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	adminService, err := wireAdminService(rt.DB, bootstrap.MustMakeJWTService(rt.Container))
	if err != nil {
		return err
	}

	adminService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移。
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.AdminUser{},
		&models.AdminRole{},
		&models.AdminPermission{},
		&models.Setting{},
		&models.ActivityLog{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}

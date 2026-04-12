// Package main 媒体服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/media-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("media-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	storageType := rt.Config.GetString("storage.type")
	localPath := rt.Config.GetString("storage.local.path")
	urlPrefix := rt.Config.GetString("storage.local.url_prefix")
	if localPath == "" {
		localPath = "./uploads"
	}
	if urlPrefix == "" {
		urlPrefix = "/media"
	}

	if storageType == "local" || storageType == "" {
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return fmt.Errorf("创建上传目录失败: %w", err)
		}
	}

	mediaService, err := wireMediaService(rt.DB, bootstrap.MustMakeJWTService(rt.Container), localPath, urlPrefix, storageType)
	if err != nil {
		return err
	}

	mediaService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Picture{},
		&models.ProductPicture{},
		&models.CategoryPicture{},
		&models.ManufacturerPicture{},
		&models.VendorPicture{},
		&models.Document{},
		&models.DownloadDownload{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}
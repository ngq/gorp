// Package main 商品目录服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/catalog-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("catalog-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	catalogService, err := wireCatalogService(rt.DB)
	if err != nil {
		return err
	}

	catalogService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移。
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Product{},
		&models.Category{},
		&models.Manufacturer{},
		&models.ProductCategory{},
		&models.ProductAttribute{},
		&models.ProductAttributeValue{},
		&models.ProductAttributeMapping{},
		&models.ProductPicture{},
		&models.ProductReview{},
		&models.ProductSpecificationAttribute{},
		&models.ProductSpecificationAttributeOption{},
		&models.ProductSpecificationAttributeMapping{},
		&models.RelatedProduct{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}

// Package main 购物车服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
	"nop-go/services/cart-service/internal/models"

	"gorm.io/gorm"
)

func main() {
	if err := bootstrap.BootHTTPService("cart-service", bootstrap.Options{}, migrate, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
	return autoMigrate(rt.DB)
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	cartService, err := wireCartService(rt.DB, bootstrap.MustMakeJWTService(rt.Container))
	if err != nil {
		return err
	}

	cartService.RegisterRoutes(rt.Engine)
	return nil
}

// autoMigrate 执行数据库表结构迁移。
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.ShoppingCart{},
		&models.CartItem{},
		&models.Wishlist{},
		&models.WishlistItem{},
	)
	if err != nil {
		return fmt.Errorf("表结构迁移失败: %w", err)
	}
	return nil
}

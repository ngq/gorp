package main

import (
	"fmt"
	"os"

	orderdata "nop-go/services/order/internal/data"
	orderhttp "nop-go/services/order/internal/server/http"
	"github.com/ngq/gorp"
	_ "nop-go/shared" // 微服务治理组件统一导入
)

func main() {
	if err := gorp.Run(
		gorp.GRPC(),
		gorp.WithMicroGovernance(),
		gorp.WithMigrate(migrate),
		gorp.WithSetup(setup),
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	// 自动迁移所有订单服务相关表
	return rt.DB.AutoMigrate(
		&orderdata.OrderPO{},
		&orderdata.OrderItemPO{},
		&orderdata.ShipmentPO{},
		&orderdata.ShoppingCartItemPO{},
		&orderdata.WishlistItemPO{},
		&orderdata.ReturnRequestPO{},
	)
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("order-service requires gorm database")
	}
	services, err := wireOrderServices(rt.DB)
	if err != nil {
		return err
	}
	orderhttp.RegisterRoutes(rt.Router, services)
	return nil
}

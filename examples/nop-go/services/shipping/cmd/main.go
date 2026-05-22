// Package main 提供 shipping 服务的启动入口
package main

import (
	"fmt"
	"os"

	shippingdata "nop-go/services/shipping/internal/data"
	shippinghttp "nop-go/services/shipping/internal/server/http"
	shippingservice "nop-go/services/shipping/internal/service"

	gorp "github.com/ngq/gorp"
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

// migrate 自动迁移配送服务相关表
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&shippingdata.ProviderPO{},
		&shippingdata.MethodPO{},
		&shippingdata.DeliveryDatePO{},
		&shippingdata.WarehousePO{},
	)
}

// setup 初始化配送服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("shipping-service requires gorm database")
	}
	svc := shippingservice.NewShippingService(rt.DB)
	shippinghttp.RegisterRoutes(rt.Router, svc.Handler)
	return nil
}
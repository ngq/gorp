// Package main 提供 discount 服务的启动入口
package main

import (
	"fmt"
	"os"

	discountdata "nop-go/services/discount/internal/data"
discounthttp "nop-go/services/discount/internal/server/http"
	discountservice "nop-go/services/discount/internal/service"

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

// migrate 自动迁移折扣服务相关表
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&discountdata.DiscountPO{},
		&discountdata.DiscountProductPO{},
		&discountdata.DiscountCategoryPO{},
		&discountdata.DiscountManufacturerPO{},
		&discountdata.DiscountUsageHistoryPO{},
	)
}

// setup 初始化折扣服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("discount-service requires gorm database")
	}
	svc := discountservice.NewDiscountService(rt.DB)
	discounthttp.RegisterRoutes(rt.Router, svc.Handler)
	return nil
}
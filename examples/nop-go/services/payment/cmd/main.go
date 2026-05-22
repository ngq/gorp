// Package main 提供 payment 服务的启动入口
package main

import (
	"fmt"
	"os"

	paymentdata "nop-go/services/payment/internal/data"
	paymenthttp "nop-go/services/payment/internal/server/http"
	paymentservice "nop-go/services/payment/internal/service"

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

// migrate 自动迁移支付服务相关表
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&paymentdata.PaymentMethodPO{},
		&paymentdata.MethodRestrictionPO{},
	)
}

// setup 初始化支付服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("payment-service requires gorm database")
	}
	services := paymentservice.NewServices(rt.DB)
	handler := paymentservice.NewPaymentHandler(services.Payment)
	paymenthttp.RegisterRoutes(rt.Router, handler)
	return nil
}
// Package main 交易服务启动入口
// trade-service 合并了 order + payment + shipping + tax 四大业务域
package main

import (
	"fmt"
	"os"

	tradedata "nop-go/services/trade-service/internal/data"
	tradegrpc "nop-go/services/trade-service/internal/server/grpc"
	tradehttp "nop-go/services/trade-service/internal/server/http"
	tradeservice "nop-go/services/trade-service/internal/service"

	gorp "github.com/ngq/gorp"
	"google.golang.org/grpc"
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

// migrate 自动迁移交易服务所有相关表
// 包含订单、购物车、心愿单、退换货、支付、物流、税务全部 PO
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		// 订单相关
		&tradedata.OrderPO{},
		&tradedata.OrderItemPO{},
		&tradedata.CartItemPO{},
		&tradedata.WishlistItemPO{},
		&tradedata.ReturnRequestPO{},
		// 支付相关
		&tradedata.PaymentPO{},
		&tradedata.PaymentMethodPO{},
		// 物流相关（ShippingProvider 原名 Provider，已重命名）
		&tradedata.ShippingProviderPO{},
		&tradedata.ShippingOrderPO{},
		&tradedata.ShippingEventPO{},
		&tradedata.ShippingRatePO{},
		// 税务相关（TaxProvider/TaxCategory 原 Provider/Category，已重命名）
		&tradedata.TaxProviderPO{},
		&tradedata.TaxCategoryPO{},
		&tradedata.TaxRatePO{},
		&tradedata.TaxTransactionPO{},
	)
}

// setup 初始化交易服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("trade-service requires gorm database")
	}
	services := tradeservice.NewServices(rt.DB)
	tradehttp.RegisterRoutes(rt.Router, services)

	// 注册 gRPC 服务
	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		tradegrpc.RegisterTradeService(server, services.Order)
		return nil
	})
}
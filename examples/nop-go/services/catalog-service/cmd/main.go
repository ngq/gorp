// Package main 是 catalog-service 的入口。
// catalog-service 合并了 product、directory、media、seo 四个服务，
// 提供商品、目录、媒体、SEO 统一管理能力。
package main

import (
	"fmt"
	"os"

	catalogdata "nop-go/services/catalog-service/internal/data"
	cataloggrpc "nop-go/services/catalog-service/internal/server/grpc"
	cataloghttp "nop-go/services/catalog-service/internal/server/http"

	"github.com/ngq/gorp"
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

// migrate 自动迁移数据库表结构。
// 包含四个服务的所有 PO 类型：
//   - product: ProductPO, CategoryPO, ManufacturerPO, ProductReviewPO, RecentlyViewedPO
//   - directory: CountryPO, StatePO, CurrencyPO
//   - media: MediaPO
//   - seo: SeoPO
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		// product 服务的 PO
		&catalogdata.ProductPO{},
		&catalogdata.CategoryPO{},
		&catalogdata.ManufacturerPO{},
		&catalogdata.ProductReviewPO{},
		&catalogdata.RecentlyViewedPO{},
		// directory 服务的 PO
		&catalogdata.CountryPO{},
		&catalogdata.StatePO{},
		&catalogdata.CurrencyPO{},
		// media 服务的 PO
		&catalogdata.MediaPO{},
		// seo 服务的 PO
		&catalogdata.SeoPO{},
	)
}

// setup 初始化业务服务并注册路由。
// 通过 wire 注入所有依赖，注册合并后的 HTTP 路由和 gRPC 服务。
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("catalog-service requires gorm database")
	}
	services, err := wireCatalogServices(rt.DB)
	if err != nil {
		return err
	}
	cataloghttp.RegisterRoutes(rt.Router, services)

	// 注册 gRPC 服务
	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		cataloggrpc.RegisterCatalogService(server, services.Product, services.Directory)
		return nil
	})
}

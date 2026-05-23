package main

import (
	"fmt"
	"os"

	contentdata "nop-go/services/content-service/internal/data"
	contentgrpc "nop-go/services/content-service/internal/server/grpc"
	contenthttp "nop-go/services/content-service/internal/server/http"
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

// migrate 自动迁移所有持久化对象。
// 合并了 content / localization / affiliate 三个子服务的所有 PO 类型。
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	// 自动迁移所有表：内容 + 本地化 + 推广
	return rt.DB.AutoMigrate(
		// ---- 内容服务 PO ----
		&contentdata.BlogPO{},
		&contentdata.NewsPO{},
		&contentdata.TopicPO{},
		&contentdata.PollPO{},
		// ---- 本地化服务 PO ----
		&contentdata.LanguagePO{},
		&contentdata.LocaleResourcePO{},
		// ---- 推广服务 PO ----
		&contentdata.AffiliatePO{},
		&contentdata.AffiliateOrderPO{},
		&contentdata.AffiliateCustomerPO{},
	)
}

// setup 初始化内容服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("content-service requires gorm database")
	}
	services, err := wireContentServices(rt.DB)
	if err != nil {
		return err
	}
	contenthttp.RegisterRoutes(rt.Router, services)

	// 注册 gRPC 服务
	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		contentgrpc.RegisterContentService(server, services.Content)
		return nil
	})
}
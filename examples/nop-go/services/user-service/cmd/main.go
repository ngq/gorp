package main

import (
	"fmt"
	"os"

	userdata "nop-go/services/user-service/internal/data"
	usergrpc "nop-go/services/user-service/internal/server/grpc"
	userhttp "nop-go/services/user-service/internal/server/http"
	"nop-go/services/user-service/internal/service"

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

// migrate 数据库自动迁移。
// 合并了原 user 和 gdpr 两个服务的所有持久化对象：
// - users: 用户基本信息表
// - addresses: 用户地址表
// - external_associations: 外部关联绑定表
// - downloadable_products: 可下载产品表
// - gdprs: GDPR 数据删除请求表
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&userdata.UserPO{},
		&userdata.AddressPO{},
		&userdata.ExternalAssociationPO{},
		&userdata.DownloadableProductPO{},
		&userdata.GdprPO{},
	)
}

// setup 初始化服务容器并注册路由。
// 将用户子服务和 GDPR 子服务统一注入到同一个 Services 容器，
// 然后注册合并后的 HTTP 路由和 gRPC 服务。
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("user-service requires gorm database")
	}

	// 初始化服务容器（包含用户和 GDPR 两个子服务的完整依赖注入）
	services := service.NewServices(rt.DB)

	// 注册合并后的 HTTP 路由
	userhttp.RegisterRoutes(rt.Router, services)

	// 注册 gRPC 服务
	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		usergrpc.RegisterUserService(server, services.User)
		return nil
	})
}
package main

import (
	"fmt"
	"os"

	userdata "nop-go/services/user/internal/data"
	userhttp "nop-go/services/user/internal/server/http"
	"nop-go/services/user/internal/service"

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

// migrate 数据库自动迁移。
// 包含用户、地址、外部关联、可下载产品四张表。
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&userdata.UserPO{},
		&userdata.AddressPO{},
		&userdata.ExternalAssociationPO{},
		&userdata.DownloadableProductPO{},
	)
}

// setup 初始化服务容器并注册路由。
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("user-service requires gorm database")
	}

	// 初始化服务容器（包含仓储、用例、服务的完整依赖注入）
	services := service.NewServices(rt.DB)

	// 注册 HTTP 路由
	userhttp.RegisterRoutes(rt.Router, services)
	return nil
}
// Package main 提供 tax 服务的启动入口
package main

import (
	"fmt"
	"os"

	taxdata "nop-go/services/tax/internal/data"
	taxhttp "nop-go/services/tax/internal/server/http"
	taxservice "nop-go/services/tax/internal/service"

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

// migrate 自动迁移税务服务相关表
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&taxdata.ProviderPO{},
		&taxdata.CategoryPO{},
	)
}

// setup 初始化税务服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("tax-service requires gorm database")
	}
	svc := taxservice.NewTaxService(rt.DB)
	taxhttp.RegisterRoutes(rt.Router, svc.Handler)
	return nil
}
// Package main 提供 admin-service 的启动入口
//
// admin-service 是管理后台统一服务，合并了6个模块：
// 1. 安全（权限 + ACL）
// 2. 插件
// 3. 门店
// 4. 日志（活动日志 + 系统日志）
// 5. 优惠（优惠 + 使用记录）
// 6. 供应商
package main

import (
	"fmt"
	"os"

	"nop-go/services/admin-service/internal/data"
	admingrpc "nop-go/services/admin-service/internal/server/grpc"
	adminhttp "nop-go/services/admin-service/internal/server/http"

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

// migrate 自动迁移管理后台服务相关表
// 包含6个模块的所有 PO 类型
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		// 安全模块
		&data.PermissionPO{},
		&data.ACLPO{},
		// 插件模块
		&data.PluginPO{},
		// 门店模块
		&data.StorePO{},
		// 日志模块
		&data.ActivityLogPO{},
		&data.SystemLogPO{},
		// 优惠模块
		&data.DiscountPO{},
		&data.DiscountUsagePO{},
		// 供应商模块
		&data.VendorPO{},
	)
}

// setup 初始化管理后台服务依赖并注册路由
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("admin-service requires gorm database")
	}
	services, err := wireAdminServices(rt.DB)
	if err != nil {
		return err
	}
	adminhttp.RegisterRoutes(rt.Router, services)

	// 注册 gRPC 服务
	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		admingrpc.RegisterAdminService(server, services.Discount, services.Security)
		return nil
	})
}
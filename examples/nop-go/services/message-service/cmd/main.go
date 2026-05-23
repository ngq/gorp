// Package main 消息服务入口。
//
// message-service 是消息模板管理微服务，提供消息模板的 CRUD、测试、复制等能力。
// 启动时通过 gorp.Run 初始化框架基础设施（HTTP/gRPC、数据库、治理组件等），
// 并通过 wire 注入服务依赖、注册路由。
package main

import (
	"fmt"
	"os"

	messagedata "nop-go/services/message-service/internal/data"
	messagegrpc "nop-go/services/message-service/internal/server/grpc"
	messagehttp "nop-go/services/message-service/internal/server/http"

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
//
// 框架启动时调用，将 MessageTemplatePO 对应的表自动创建/更新到数据库。
// 若数据库未配置则跳过。
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(&messagedata.MessageTemplatePO{})
}

// setup 初始化业务服务并注册路由。
func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("message-service requires gorm database")
	}
	services, err := wireMessageServices(rt.DB)
	if err != nil {
		return err
	}
	messagehttp.RegisterRoutes(rt.Router, services)

	// 注册 gRPC 服务
	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		messagegrpc.RegisterMessageService(server, services.Message)
		return nil
	})
}

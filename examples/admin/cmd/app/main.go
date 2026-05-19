package main

import (
	"fmt"
	"os"

	apphttp "admin/app/http"
	"admin/internal/biz"
	"admin/internal/data"
	"admin/internal/service"

	"github.com/ngq/gorp"
)

// main 是项目主入口。
//
// 中文说明：
// - 用户日常开发时，直接 `go run ./cmd/app` 或 IDE 调试即可启动 HTTP 服务；
// - 使用框架统一入口 gorp，简化导入；
// - 默认主线直接走 typed runtime + direct constructor，不要求先理解 ServiceProvider / ExtraProviders。
func main() {
	if err := gorp.Run(
		"admin",
		gorp.HTTP(),
		gorp.WithMonolithMode(),
		gorp.WithMigrate(migrate),
		gorp.WithSetup(setup),
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&data.DemoPO{},
		&data.UserPO{},
		&data.RolePO{},
	)
}

func setup(rt *gorp.HTTPRuntime) error {
	// 创建数据层
	d := data.NewData(rt.DB)
	demoRepo := data.NewDemoRepo(d)
	userRepo := data.NewUserRepo(d)
	roleRepo := data.NewRoleRepo(d)

	// 创建业务层
	b := biz.NewBiz(demoRepo, userRepo, roleRepo)

	// 创建服务层
	svc := service.NewServices(b)

	// 注册 HTTP 路由
	apphttp.RegisterRoutes(rt.Router, svc)

	// 打印启动信息
	fmt.Printf("[INFO] admin example initialized\n")
	fmt.Printf("[INFO] default accounts: admin/admin123, editor/editor123\n")

	return nil
}

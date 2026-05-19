package main

import (
	"fmt"
	"os"

	"github.com/ngq/gorp"
	"google.golang.org/grpc"
	userdata "grpc-demo/services/user/internal/data"
	usergrpc "grpc-demo/services/user/internal/server/grpc"
	userhttp "grpc-demo/services/user/internal/server/http"
)

// main 是 user 服务的主入口。
func main() {
	if err := gorp.Run(
		"user-service",
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
	return rt.DB.AutoMigrate(&userdata.UserPO{})
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("user-service requires gorm database")
	}
	services, err := wireUserServices(rt.DB)
	if err != nil {
		return err
	}
	userhttp.RegisterRoutes(rt.Router, services)

	registrar, err := gorp.MakeGRPCServerRegistrar(rt.Container)
	if err != nil {
		return err
	}
	return registrar.RegisterProto(func(server *grpc.Server) error {
		usergrpc.RegisterUserService(server, services)
		return nil
	})
}

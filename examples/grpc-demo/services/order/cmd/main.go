package main

import (
	"fmt"
	"os"

	"github.com/ngq/gorp"
	orderdata "grpc-demo/services/order/internal/data"
	orderhttp "grpc-demo/services/order/internal/server/http"
)

// main 是 order 服务的主入口。
func main() {
	if err := gorp.Run(
		"order-service",
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
	return rt.DB.AutoMigrate(&orderdata.OrderPO{})
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("order-service requires gorm database")
	}
	connFactory, err := gorp.MakeGRPCConnFactory(rt.Container)
	if err != nil {
		return err
	}
	dlock, err := gorp.MakeDistributedLock(rt.Container)
	if err != nil {
		return err
	}
	services, err := wireOrderServices(rt.DB, connFactory, dlock)
	if err != nil {
		return err
	}
	orderhttp.RegisterRoutes(rt.Router, services)
	return nil
}

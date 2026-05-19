package main

import (
	"fmt"
	"os"

	"github.com/ngq/gorp"
	productdata "grpc-demo/services/product/internal/data"
	producthttp "grpc-demo/services/product/internal/server/http"
)

// main 是 product 服务的主入口。
func main() {
	if err := gorp.Run(
		"product-service",
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
	return rt.DB.AutoMigrate(&productdata.ProductPO{})
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("product-service requires gorm database")
	}
	publisher, err := gorp.MakeMessagePublisher(rt.Container)
	if err != nil {
		return err
	}
	subscriber, err := gorp.MakeMessageSubscriber(rt.Container)
	if err != nil {
		return err
	}
	services, err := wireProductServices(rt.DB, publisher, subscriber)
	if err != nil {
		return err
	}
	producthttp.RegisterRoutes(rt.Router, services)
	return nil
}

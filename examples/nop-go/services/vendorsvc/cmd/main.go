package main

import (
	"fmt"
	"os"

	vendordata "nop-go/services/vendorsvc/internal/data"
	vendorhttp "nop-go/services/vendorsvc/internal/server/http"
	"github.com/ngq/gorp"
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

func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(&vendordata.VendorPO{})
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("vendor-service requires gorm database")
	}
	services, err := wireVendorServices(rt.DB)
	if err != nil {
		return err
	}
	vendorhttp.RegisterRoutes(rt.Router, services)
	return nil
}

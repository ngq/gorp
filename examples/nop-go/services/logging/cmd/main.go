package main

import (
	"fmt"
	"os"

	loggingdata "nop-go/services/logging/internal/data"
	logginghttp "nop-go/services/logging/internal/server/http"
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
	return rt.DB.AutoMigrate(&loggingdata.LoggingPO{})
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("logging-service requires gorm database")
	}
	services, err := wireLoggingServices(rt.DB)
	if err != nil {
		return err
	}
	logginghttp.RegisterRoutes(rt.Router, services)
	return nil
}

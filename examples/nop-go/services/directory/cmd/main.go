package main

import (
	"fmt"
	"os"

	directorydata "nop-go/services/directory/internal/data"
	directoryhttp "nop-go/services/directory/internal/server/http"
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
	// 自动迁移 directory 服务的三类持久化对象：国家、省/州、货币
	return rt.DB.AutoMigrate(
		&directorydata.CountryPO{},
		&directorydata.StatePO{},
		&directorydata.CurrencyPO{},
	)
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("directory-service requires gorm database")
	}
	services, err := wireDirectoryServices(rt.DB)
	if err != nil {
		return err
	}
	directoryhttp.RegisterRoutes(rt.Router, services)
	return nil
}

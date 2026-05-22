package main

import (
	"fmt"
	"os"

	localizationdata "nop-go/services/localization/internal/data"
	localizationhttp "nop-go/services/localization/internal/server/http"
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
	return rt.DB.AutoMigrate(&localizationdata.LanguagePO{}, &localizationdata.LocaleResourcePO{})
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("localization-service requires gorm database")
	}
	services, err := wireLocalizationServices(rt.DB)
	if err != nil {
		return err
	}
	localizationhttp.RegisterRoutes(rt.Router, services)
	return nil
}

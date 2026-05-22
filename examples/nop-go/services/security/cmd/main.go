package main

import (
	"fmt"
	"os"

	securitydata "nop-go/services/security/internal/data"
	securityhttp "nop-go/services/security/internal/server/http"
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

// migrate 自动迁移安全服务相关表
func migrate(rt *gorp.HTTPRuntime) error {
	if rt == nil || rt.DB == nil {
		return nil
	}
	return rt.DB.AutoMigrate(
		&securitydata.PermissionPO{},
		&securitydata.ACLRecordPO{},
	)
}

func setup(rt *gorp.HTTPRuntime) error {
	if rt.DB == nil {
		return fmt.Errorf("security-service requires gorm database")
	}
	services, err := wireSecurityServices(rt.DB)
	if err != nil {
		return err
	}
	securityhttp.RegisterRoutes(rt.Router, services)
	return nil
}

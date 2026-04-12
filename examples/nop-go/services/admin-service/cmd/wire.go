//go:build wireinject

package main

import (
	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/data"
	"nop-go/services/admin-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

func wireAdminService(db *gorm.DB, jwtSvc contract.JWTService) (*service.AdminService, error) {
	panic(wire.Build(
		data.NewAdminUserRepository,
		data.NewAdminRoleRepository,
		data.NewAdminPermissionRepository,
		data.NewSettingRepository,
		data.NewActivityLogRepository,
		biz.NewAdminUserUseCase,
		biz.NewAdminRoleUseCase,
		biz.NewSettingUseCase,
		biz.NewActivityLogUseCase,
		service.NewAdminService,
	))
}

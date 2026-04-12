//go:build wireinject

package main

import (
	"nop-go/services/theme-service/internal/biz"
	"nop-go/services/theme-service/internal/data"
	"nop-go/services/theme-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

func wireThemeService(db *gorm.DB, jwtSvc contract.JWTService, config biz.ThemeConfig) (*service.ThemeService, error) {
	panic(wire.Build(
		data.NewThemeRepository,
		data.NewThemeVariableRepository,
		data.NewThemeConfigurationRepository,
		data.NewCustomerThemeSettingRepository,
		data.NewThemeFileRepository,
		biz.NewThemeUseCase,
		biz.NewThemeConfigurationUseCase,
		biz.NewCustomerThemeUseCase,
		biz.NewThemeFileUseCase,
		service.NewThemeService,
	))
}

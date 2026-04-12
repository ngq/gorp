//go:build wireinject

package main

import (
	"nop-go/services/localization-service/internal/biz"
	"nop-go/services/localization-service/internal/data"
	"nop-go/services/localization-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

// wireLocalizationService 使用 Wire 生成 localization-service 的装配代码。
func wireLocalizationService(db *gorm.DB, jwtSvc contract.JWTService) (*service.LocalizationService, error) {
	panic(wire.Build(
		data.NewLanguageRepository,
		data.NewLocaleStringResourceRepository,
		data.NewCurrencyRepository,
		biz.NewLocalizationUseCase,
		service.NewLocalizationService,
	))
}

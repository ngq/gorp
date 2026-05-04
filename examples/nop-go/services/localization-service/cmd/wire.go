//go:build wireinject

package main

import (
	"nop-go/services/localization-service/internal/biz"
	"nop-go/services/localization-service/internal/data"
	"nop-go/services/localization-service/internal/service"

	"github.com/google/wire"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"gorm.io/gorm"
)

// wireLocalizationService 浣跨敤 Wire 鐢熸垚 localization-service 鐨勮閰嶄唬鐮併€?
func wireLocalizationService(db *gorm.DB, jwtSvc securitycontract.JWTService) (*service.LocalizationService, error) {
	panic(wire.Build(
		data.NewLanguageRepository,
		data.NewLocaleStringResourceRepository,
		data.NewCurrencyRepository,
		biz.NewLocalizationUseCase,
		service.NewLocalizationService,
	))
}

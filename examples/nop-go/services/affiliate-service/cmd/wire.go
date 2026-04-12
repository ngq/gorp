//go:build wireinject

package main

import (
	"nop-go/services/affiliate-service/internal/biz"
	"nop-go/services/affiliate-service/internal/data"
	"nop-go/services/affiliate-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

func wireAffiliateService(db *gorm.DB, jwtSvc contract.JWTService, config biz.AffiliateConfig) (*service.AffiliateService, error) {
	panic(wire.Build(
		data.NewAffiliateRepository,
		data.NewAffiliateOrderRepository,
		data.NewAffiliateReferralRepository,
		data.NewAffiliateCommissionRepository,
		data.NewAffiliatePayoutRepository,
		biz.NewAffiliateUseCase,
		service.NewAffiliateService,
	))
}

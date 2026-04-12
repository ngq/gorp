//go:build wireinject

package main

import (
	"nop-go/services/price-service/internal/biz"
	"nop-go/services/price-service/internal/data"
	"nop-go/services/price-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wirePriceService(db *gorm.DB) (*service.PriceService, error) {
	panic(wire.Build(
		data.NewTaxRateRepository,
		data.NewDiscountRepository,
		data.NewProductPriceRepository,
		biz.NewTaxUseCase,
		biz.NewDiscountUseCase,
		biz.NewPriceUseCase,
		service.NewPriceService,
	))
}

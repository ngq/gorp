//go:build wireinject

package main

import (
	"nop-go/services/shipping-service/internal/biz"
	"nop-go/services/shipping-service/internal/data"
	"nop-go/services/shipping-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireShippingService(db *gorm.DB) (*service.ShippingService, error) {
	panic(wire.Build(
		data.NewShipmentRepository,
		data.NewShipmentItemRepository,
		data.NewShippingMethodRepository,
		data.NewShipmentTrackingRepository,
		biz.NewShipmentUseCase,
		biz.NewShippingMethodUseCase,
		service.NewShippingService,
	))
}

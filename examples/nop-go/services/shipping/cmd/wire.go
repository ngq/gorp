//go:build wireinject

package main

import (
	"nop-go/services/shipping/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireShippingServices(db *gorm.DB) (*service.ShippingService, error) {
	panic(wire.Build(
		service.NewShippingService,
	))
}

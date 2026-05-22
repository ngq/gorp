//go:build wireinject

package main

import (
	"nop-go/services/payment/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wirePaymentServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

//go:build wireinject

package main

import (
	"nop-go/services/vendor/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireVendorServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

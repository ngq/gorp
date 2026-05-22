//go:build wireinject

package main

import (
	"nop-go/services/localization/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireLocalizationServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

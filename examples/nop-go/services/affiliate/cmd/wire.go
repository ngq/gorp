//go:build wireinject

package main

import (
	"nop-go/services/affiliate/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireAffiliateServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

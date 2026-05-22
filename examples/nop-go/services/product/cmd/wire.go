//go:build wireinject

package main

import (
	"nop-go/services/product/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireProductServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

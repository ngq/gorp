//go:build wireinject

package main

import (
	"nop-go/services/media/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireMediaServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

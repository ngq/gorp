//go:build wireinject

package main

import (
	"nop-go/services/logging/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireLoggingServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

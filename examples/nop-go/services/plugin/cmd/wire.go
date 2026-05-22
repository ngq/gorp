//go:build wireinject

package main

import (
	"nop-go/services/plugin/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wirePluginServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

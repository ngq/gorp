//go:build wireinject

package main

import (
	"nop-go/services/content/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireContentServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

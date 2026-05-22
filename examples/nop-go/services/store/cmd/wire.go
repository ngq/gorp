//go:build wireinject

package main

import (
	"nop-go/services/store/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireStoreServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

//go:build wireinject

package main

import (
	"nop-go/services/directory/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireDirectoryServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

//go:build wireinject

package main

import (
	"nop-go/services/user/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireUserServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

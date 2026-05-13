//go:build wireinject

package main

import (
	"grpc-demo/services/user/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireUserServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

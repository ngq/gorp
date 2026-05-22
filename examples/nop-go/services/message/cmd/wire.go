//go:build wireinject

package main

import (
	"nop-go/services/message/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireMessageServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

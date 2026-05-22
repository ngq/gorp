//go:build wireinject

package main

import (
	"nop-go/services/gdpr/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireGdprServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

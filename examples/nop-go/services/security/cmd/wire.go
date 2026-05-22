//go:build wireinject

package main

import (
	"nop-go/services/security/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireSecurityServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

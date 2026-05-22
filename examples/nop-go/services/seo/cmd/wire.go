//go:build wireinject

package main

import (
	"nop-go/services/seo/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireSeoServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

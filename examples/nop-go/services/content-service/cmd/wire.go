//go:build wireinject

package main

import (
	"nop-go/services/content-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireContentServices Wire 注入入口，生成服务集合
func wireContentServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}
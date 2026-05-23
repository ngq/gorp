//go:build wireinject

// Package main 提供 Wire 依赖注入入口。
// catalog-service 合并四个子服务，Wire 注入统一 Services 结构。
package main

import (
	"nop-go/services/catalog-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireCatalogServices 通过 Wire 生成依赖注入代码。
// 合并 product + directory + media + seo 四个服务的依赖注入。
func wireCatalogServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

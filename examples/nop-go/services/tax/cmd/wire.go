//go:build wireinject

package main

import (
	"nop-go/services/tax/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireTaxServices 通过 Wire 注入数据库连接，构建税务服务实例
func wireTaxServices(db *gorm.DB) (*service.TaxService, error) {
	panic(wire.Build(
		service.NewTaxService,
	))
}

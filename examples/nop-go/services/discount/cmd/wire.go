//go:build wireinject

package main

import (
	"nop-go/services/discount/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireDiscountServices 通过 Wire 注入数据库连接，构建折扣服务实例
func wireDiscountServices(db *gorm.DB) (*service.DiscountService, error) {
	panic(wire.Build(
		service.NewDiscountService,
	))
}

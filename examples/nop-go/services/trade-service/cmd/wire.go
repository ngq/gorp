//go:build wireinject

// Wire 注入模板，生成 wire_gen.go
package main

import (
	"nop-go/services/trade-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireTradeServices 通过 Wire 注入数据库连接，构建交易服务容器
func wireTradeServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}
//go:build wireinject

// Package main 提供 Wire 依赖注入入口
package main

import (
	"nop-go/services/admin-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireAdminServices 通过 Wire 注入数据库连接，构建管理后台服务实例
func wireAdminServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}
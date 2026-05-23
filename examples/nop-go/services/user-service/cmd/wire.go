//go:build wireinject

package main

import (
	"nop-go/services/user-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireUserServices 使用 Wire 生成用户服务的依赖注入代码。
// 合并了原 user 和 gdpr 服务的所有依赖链路。
func wireUserServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}
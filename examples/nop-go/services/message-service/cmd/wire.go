//go:build wireinject

// Package main Wire 依赖注入构建器。
//
// 此文件仅包含 Wire 注入模板，由 wire 工具生成 wire_gen.go 实际实现。
// 编译时通过 build tag 排除，不会参与最终构建。
package main

import (
	"nop-go/services/message-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireMessageServices 通过 Wire 注入消息服务依赖。
//
// Wire 会根据 Build 列表中的 Provider 自动推导依赖图，
// 生成 wire_gen.go 中的实际实现代码。
func wireMessageServices(db *gorm.DB) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}

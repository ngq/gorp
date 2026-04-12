// Package bootstrap 服务启动封装
// 直接使用框架能力，提供类型别名简化调用
package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"

	gormpkg "gorm.io/gorm"
)

// 类型别名，简化调用
type HTTPServiceOptions = bootstrap.HTTPServiceOptions
type HTTPServiceRuntime = bootstrap.HTTPServiceRuntime

// 函数别名，直接委托给框架
var (
	// NewHTTPServiceRuntime 创建 HTTP 服务运行时
	NewHTTPServiceRuntime = bootstrap.NewHTTPServiceRuntime
	// BootHTTPService 启动 HTTP 服务
	BootHTTPService = bootstrap.BootHTTPService
	// RegisterHealthCheck 注册健康检查端点
	RegisterHealthCheck = bootstrap.RegisterHealthCheck
	// RegisterMetricsEndpoint 注册 Prometheus 指标端点
	RegisterMetricsEndpoint = bootstrap.RegisterMetricsEndpoint
	// RunHTTP 启动 HTTP 服务
	RunHTTP = bootstrap.RunHTTP
)

// Options 初始化选项（兼容旧代码）
type Options = bootstrap.HTTPServiceOptions

// MustMakeJWTService 从容器获取 JWT 服务
func MustMakeJWTService(c contract.Container) contract.JWTService {
	return container.MustMakeJWTService(c)
}

// MustMakeLogger 从容器获取 Logger
func MustMakeLogger(c contract.Container) contract.Logger {
	return container.MustMakeLogger(c)
}

// MustMakeGorm 从容器获取 Gorm DB
func MustMakeGorm(c contract.Container) *gormpkg.DB {
	return container.MustMakeGorm(c)
}

// MustMakeEngine 从容器获取 Gin Engine
func MustMakeEngine(c contract.Container) *gin.Engine {
	return container.MustMakeEngine(c)
}

// MustMakeConfig 从容器获取 Config
func MustMakeConfig(c contract.Container) contract.Config {
	return container.MustMakeConfig(c)
}
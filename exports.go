// Package gorp 提供 gorp 框架的统一入口。
//
// 中文说明：
// - 通过此包可以直接访问框架的核心功能，无需分别导入子包；
// - 例如：gorp.BootHTTPService；
// - 业务层日志主线改为使用 `gorp/log`，而不是 `gorp.MustMakeLogger`；
// - `MustMake*` 仍保留给启动装配层与 framework/internal helper 使用，不作为业务默认主线。
//
// 使用示例：
//
//	import "github.com/ngq/gorp"
//	import glog "github.com/ngq/gorp/log"
//
//	func main() {
//	    err := gorp.BootHTTPService("my-service", gorp.HTTPServiceOptions{}, migrate, setup)
//	    if err != nil {
//	        panic(err)
//	    }
//	}
//
//	func migrate(rt *gorp.HTTPServiceRuntime) error {
//	    return gorp.AutoMigrateModels(rt, &User{}, &Order{})
//	}
//
//	func setup(rt *gorp.HTTPServiceRuntime) error {
//	    glog.Info("service setup")
//	    // 业务默认优先使用 typed runtime 与轻量 helper
//	    if rt.DB == nil {
//	        return fmt.Errorf("database required")
//	    }
//	    return nil
//	}
package gorp

import (
	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

// HTTPServiceOptions HTTP 服务启动选项。
type HTTPServiceOptions = bootstrap.HTTPServiceOptions

// HTTPServiceRuntime HTTP 服务运行时。
type HTTPServiceRuntime = bootstrap.HTTPServiceRuntime

// Container 依赖注入容器接口。
type Container = contract.Container

// ServiceProvider 服务提供者接口。
type ServiceProvider = contract.ServiceProvider

// BootHTTPService 启动 HTTP 服务。
func BootHTTPService(serviceName string, opts HTTPServiceOptions, migrate func(*HTTPServiceRuntime) error, setup func(*HTTPServiceRuntime) error) error {
	return bootstrap.BootHTTPService(serviceName, opts, migrate, setup)
}

// RunHTTP 运行 HTTP 服务。
func RunHTTP(c Container, logger contract.Logger) error {
	return bootstrap.RunHTTP(c, logger)
}

// NewHTTPServiceRuntime 创建 HTTP 服务运行时。
func NewHTTPServiceRuntime(serviceName string, opts HTTPServiceOptions) (*HTTPServiceRuntime, error) {
	return bootstrap.NewHTTPServiceRuntime(serviceName, opts)
}

// RegisterHealthCheck 注册健康检查端点。
func RegisterHealthCheck(engine *gin.Engine, serviceName string) {
	bootstrap.RegisterHealthCheck(engine, serviceName)
}

// RegisterMetricsEndpoint 注册 Prometheus 指标端点。
func RegisterMetricsEndpoint(engine *gin.Engine) {
	bootstrap.RegisterMetricsEndpoint(engine)
}

// AutoMigrateModels 自动迁移数据库模型。
func AutoMigrateModels(runtime *HTTPServiceRuntime, models ...any) error {
	return bootstrap.AutoMigrateModels(runtime, models)
}

// MustMakeLogger 从容器获取日志服务。
//
// 中文说明：
// - 这是 framework/internal helper；
// - 业务层统一日志入口直接使用 `gorp/log` 或 `framework/log`。
func MustMakeLogger(c Container) contract.Logger {
	return container.MustMakeLogger(c)
}

// MustMakeGorm 从容器获取 Gorm 数据库连接。
func MustMakeGorm(c Container) *gorm.DB {
	return container.MustMakeGorm(c)
}

// MustMakeEngine 从容器获取 Gin Engine。
func MustMakeEngine(c Container) *gin.Engine {
	return container.MustMakeEngine(c)
}

// MustMakeConfig 从容器获取配置服务。
func MustMakeConfig(c Container) contract.Config {
	return container.MustMakeConfig(c)
}

// MustMakeJWTService 从容器获取 JWT 服务。
func MustMakeJWTService(c Container) contract.JWTService {
	return container.MustMakeJWTService(c)
}

// MustMakeValidator 从容器获取验证器。
//
// 中文说明：
// - 适用于启动阶段或明确要求校验能力已接入的路径；
// - 业务默认仍优先使用 `framework/provider/gin` 提供的 ValidateBody / ValidateQuery / ValidateForm。
func MustMakeValidator(c Container) contract.Validator {
	return container.MustMakeValidator(c)
}

// MustMakeRetry 从容器获取重试服务。
//
// 中文说明：
// - 适用于启动阶段或明确要求重试能力已接入的路径；
// - 业务默认仍优先使用 `framework/provider/gin` 提供的 DoWithRetry。
func MustMakeRetry(c Container) contract.Retry {
	return container.MustMakeRetry(c)
}

// MustMakeCache 从容器获取缓存服务。
func MustMakeCache(c Container) contract.Cache {
	return container.MustMakeCache(c)
}

// MustMakeGRPCConnFactory 从容器获取 gRPC 连接工厂。
//
// 中文说明：
// - 适用于启动装配层或 framework/internal helper；
// - 业务默认优先使用 typed runtime 与 `container.MakeGRPCConnFactory` 轻入口。
func MustMakeGRPCConnFactory(c Container) contract.GRPCConnFactory {
	return container.MustMakeGRPCConnFactory(c)
}

// MustMakeGRPCServerRegistrar 从容器获取 gRPC 服务注册器。
//
// 中文说明：
// - 适用于启动装配层或 framework/internal helper；
// - 业务默认优先使用 typed runtime 与 `container.MakeGRPCServerRegistrar` 轻入口。
func MustMakeGRPCServerRegistrar(c Container) contract.GRPCServerRegistrar {
	return container.MustMakeGRPCServerRegistrar(c)
}

// MustMakeMessagePublisher 从容器获取消息发布者。
//
// 中文说明：
// - 适用于启动装配层或 framework/internal helper；
// - 业务默认优先使用 `container.MakeMessagePublisher` 轻入口。
func MustMakeMessagePublisher(c Container) contract.MessagePublisher {
	return container.MustMakeMessagePublisher(c)
}

// MustMakeMessageSubscriber 从容器获取消息订阅者。
//
// 中文说明：
// - 适用于启动装配层或 framework/internal helper；
// - 业务默认优先使用 `container.MakeMessageSubscriber` 轻入口。
func MustMakeMessageSubscriber(c Container) contract.MessageSubscriber {
	return container.MustMakeMessageSubscriber(c)
}

// MustMakeDistributedLock 从容器获取分布式锁。
//
// 中文说明：
// - 适用于启动装配层或 framework/internal helper；
// - 业务默认优先使用 `container.MakeDistributedLock` 轻入口。
func MustMakeDistributedLock(c Container) contract.DistributedLock {
	return container.MustMakeDistributedLock(c)
}

// Package container provides runtime dependency injection container for gorp framework.
// This file exposes strongly typed helper accessors for common capabilities.
// All helpers delegate to the generic MakeWith[T] / MustMakeWith[T] functions,
// eliminating bare type assertions and providing type-mismatch error messages.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 本文件在通用运行时容器之上暴露强类型辅助访问入口。
// 所有 helper 委托给泛型 MakeWith[T] / MustMakeWith[T]，
// 消除裸类型断言，并在类型不匹配时提供可读错误信息。
package container

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	gormdb "gorm.io/gorm"
)

// MakeDBRuntime resolves the database runtime object from the container.
//
// MakeDBRuntime 从容器中解析数据库运行时对象。
func MakeDBRuntime(c runtimecontract.Container) (any, error) {
	return c.Make(datacontract.DBRuntimeKey)
}

// MakeRedis resolves the Redis capability from the container.
//
// MakeRedis 从容器中解析 Redis 能力。
func MakeRedis(c runtimecontract.Container) (datacontract.Redis, error) {
	return MakeWith[datacontract.Redis](c, datacontract.RedisKey)
}

// MakeCache resolves the cache capability from the container.
//
// MakeCache 从容器中解析缓存能力。
func MakeCache(c runtimecontract.Container) (datacontract.Cache, error) {
	return MakeWith[datacontract.Cache](c, datacontract.CacheKey)
}

// MakeGormDB resolves the Gorm database handle from the container.
//
// MakeGormDB 从容器中解析 Gorm 数据库句柄。
func MakeGormDB(c runtimecontract.Container) (*gormdb.DB, error) {
	return MakeWith[*gormdb.DB](c, datacontract.GormKey)
}

// MakeSQLX resolves the SQLX database handle from the container.
//
// MakeSQLX 从容器中解析 SQLX 数据库句柄。
func MakeSQLX(c runtimecontract.Container) (*sqlx.DB, error) {
	return MakeWith[*sqlx.DB](c, datacontract.SQLXKey)
}

// MakeMessagePublisher resolves the message publisher from the container.
//
// MakeMessagePublisher 从容器中解析消息发布能力。
func MakeMessagePublisher(c runtimecontract.Container) (integrationcontract.MessagePublisher, error) {
	return MakeWith[integrationcontract.MessagePublisher](c, integrationcontract.MessagePublisherKey)
}

// MakeMessageSubscriber resolves the message subscriber from the container.
//
// MakeMessageSubscriber 从容器中解析消息订阅能力。
func MakeMessageSubscriber(c runtimecontract.Container) (integrationcontract.MessageSubscriber, error) {
	return MakeWith[integrationcontract.MessageSubscriber](c, integrationcontract.MessageSubscriberKey)
}

// MakeDistributedLock resolves the distributed lock capability from the container.
//
// MakeDistributedLock 从容器中解析分布式锁能力。
func MakeDistributedLock(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	return MakeWith[datacontract.DistributedLock](c, datacontract.DistributedLockKey)
}

// MakeGRPCConnFactory resolves the gRPC connection factory from the container.
//
// MakeGRPCConnFactory 从容器中解析 gRPC 连接工厂。
func MakeGRPCConnFactory(c runtimecontract.Container) (transportcontract.GRPCConnFactory, error) {
	return MakeWith[transportcontract.GRPCConnFactory](c, transportcontract.GRPCConnFactoryKey)
}

// MakeGRPCServerRegistrar resolves the gRPC server registrar from the container.
//
// MakeGRPCServerRegistrar 从容器中解析 gRPC 服务注册器。
func MakeGRPCServerRegistrar(c runtimecontract.Container) (transportcontract.GRPCServerRegistrar, error) {
	return MakeWith[transportcontract.GRPCServerRegistrar](c, transportcontract.GRPCServerRegistrarKey)
}

// MakeCron resolves the cron capability from the container.
//
// MakeCron 从容器中解析 cron 能力。
func MakeCron(c runtimecontract.Container) (runtimecontract.Cron, error) {
	return MakeWith[runtimecontract.Cron](c, runtimecontract.CronKey)
}

// MakeLogger resolves the logger capability from the container.
//
// MakeLogger 从容器中解析日志能力。
func MakeLogger(c runtimecontract.Container) (observabilitycontract.Logger, error) {
	return MakeWith[observabilitycontract.Logger](c, observabilitycontract.LogKey)
}

// MakeValidator resolves the validator capability from the container.
//
// MakeValidator 从容器中解析校验器能力。
func MakeValidator(c runtimecontract.Container) (datacontract.Validator, error) {
	return MakeWith[datacontract.Validator](c, datacontract.ValidatorKey)
}

// MakeRetry resolves the retry capability from the container.
//
// MakeRetry 从容器中解析重试能力。
func MakeRetry(c runtimecontract.Container) (resiliencecontract.Retry, error) {
	return MakeWith[resiliencecontract.Retry](c, resiliencecontract.RetryKey)
}

// MakeHost resolves the host capability from the container.
//
// MakeHost 从容器中解析 host 能力。
func MakeHost(c runtimecontract.Container) (runtimecontract.Host, error) {
	return MakeWith[runtimecontract.Host](c, runtimecontract.HostKey)
}

// MakeHTTP resolves the HTTP service from the container.
//
// MakeHTTP 从容器中解析 HTTP 服务。
func MakeHTTP(c runtimecontract.Container) (transportcontract.HTTP, error) {
	return MakeWith[transportcontract.HTTP](c, transportcontract.HTTPKey)
}

// MakeHTTPRouter resolves the HTTP router facade from the container.
//
// MakeHTTPRouter 从容器中解析 HTTP 路由门面。
func MakeHTTPRouter(c runtimecontract.Container) (transportcontract.HTTPRouter, error) {
	httpSvc, err := MakeHTTP(c)
	if err != nil {
		return nil, err
	}
	return httpSvc.Router(), nil
}

// MustMakeLogger resolves the logger and panics on failure.
//
// MustMakeLogger 解析日志能力，失败时 panic。
func MustMakeLogger(c runtimecontract.Container) observabilitycontract.Logger {
	return MustMakeWith[observabilitycontract.Logger](c, observabilitycontract.LogKey)
}

// MustMakeGorm resolves the Gorm database handle and panics on failure.
//
// MustMakeGorm 解析 Gorm 数据库句柄，失败时 panic。
func MustMakeGorm(c runtimecontract.Container) *gormdb.DB {
	return MustMakeWith[*gormdb.DB](c, datacontract.GormKey)
}

// MustMakeHTTPRouter resolves the HTTP router facade and panics on failure.
//
// MustMakeHTTPRouter 解析 HTTP 路由门面，失败时 panic。
func MustMakeHTTPRouter(c runtimecontract.Container) transportcontract.HTTPRouter {
	httpSvc := MustMakeHTTP(c)
	return httpSvc.Router()
}

// MustMakeHTTP resolves the HTTP service and panics on failure.
//
// MustMakeHTTP 解析 HTTP 服务，失败时 panic。
func MustMakeHTTP(c runtimecontract.Container) transportcontract.HTTP {
	return MustMakeWith[transportcontract.HTTP](c, transportcontract.HTTPKey)
}

// MustMakeMessagePublisher resolves the message publisher and panics on failure.
//
// MustMakeMessagePublisher 解析消息发布能力，失败时 panic。
func MustMakeMessagePublisher(c runtimecontract.Container) integrationcontract.MessagePublisher {
	return MustMakeWith[integrationcontract.MessagePublisher](c, integrationcontract.MessagePublisherKey)
}

// MustMakeMessageSubscriber resolves the message subscriber and panics on failure.
//
// MustMakeMessageSubscriber 解析消息订阅能力，失败时 panic。
func MustMakeMessageSubscriber(c runtimecontract.Container) integrationcontract.MessageSubscriber {
	return MustMakeWith[integrationcontract.MessageSubscriber](c, integrationcontract.MessageSubscriberKey)
}

// MustMakeDistributedLock resolves the distributed lock capability and panics on failure.
//
// MustMakeDistributedLock 解析分布式锁能力，失败时 panic。
func MustMakeDistributedLock(c runtimecontract.Container) datacontract.DistributedLock {
	return MustMakeWith[datacontract.DistributedLock](c, datacontract.DistributedLockKey)
}

// MustMakeGRPCConnFactory resolves the gRPC connection factory and panics on failure.
//
// MustMakeGRPCConnFactory 解析 gRPC 连接工厂，失败时 panic。
func MustMakeGRPCConnFactory(c runtimecontract.Container) transportcontract.GRPCConnFactory {
	return MustMakeWith[transportcontract.GRPCConnFactory](c, transportcontract.GRPCConnFactoryKey)
}

// MustMakeGRPCServerRegistrar resolves the gRPC server registrar and panics on failure.
//
// MustMakeGRPCServerRegistrar 解析 gRPC 服务注册器，失败时 panic。
func MustMakeGRPCServerRegistrar(c runtimecontract.Container) transportcontract.GRPCServerRegistrar {
	return MustMakeWith[transportcontract.GRPCServerRegistrar](c, transportcontract.GRPCServerRegistrarKey)
}

// MustMakeValidator resolves the validator capability and panics on failure.
//
// MustMakeValidator 解析校验器能力，失败时 panic。
func MustMakeValidator(c runtimecontract.Container) datacontract.Validator {
	return MustMakeWith[datacontract.Validator](c, datacontract.ValidatorKey)
}

// MustMakeRetry resolves the retry capability and panics on failure.
//
// MustMakeRetry 解析重试能力，失败时 panic。
func MustMakeRetry(c runtimecontract.Container) resiliencecontract.Retry {
	return MustMakeWith[resiliencecontract.Retry](c, resiliencecontract.RetryKey)
}

// MustMakeConfig resolves the config capability and panics on failure.
//
// MustMakeConfig 解析配置能力，失败时 panic。
func MustMakeConfig(c runtimecontract.Container) datacontract.Config {
	return MustMakeWith[datacontract.Config](c, datacontract.ConfigKey)
}

// MustMakeCache resolves the cache capability and panics on failure.
//
// MustMakeCache 解析缓存能力，失败时 panic。
func MustMakeCache(c runtimecontract.Container) datacontract.Cache {
	return MustMakeWith[datacontract.Cache](c, datacontract.CacheKey)
}

// PingDBRuntime performs a best-effort health check against different DB runtime shapes.
//
// PingDBRuntime 针对不同数据库运行时形态执行尽力而为的健康检查。
func PingDBRuntime(dbAny any) error {
	switch db := dbAny.(type) {
	case *gormdb.DB:
		// Gorm requires unwrapping to the underlying *sql.DB before a raw ping can happen.
		// Gorm 需要先解包到底层 *sql.DB，才能执行原生 ping。
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	case *sql.DB:
		return db.Ping()
	case interface{ Ping() error }:
		// Fall back to any custom runtime object that already exposes Ping().
		// 回退支持任何已经暴露 Ping() 的自定义运行时对象。
		return db.Ping()
	default:
		// Unknown runtime shapes are treated as non-pingable rather than hard failures.
		// 未知运行时形态按"不可 ping 但不报错"处理，避免额外制造硬失败。
		return nil
	}
}

// MustMakeJWT resolves the JWT service and panics on failure.
//
// MustMakeJWT 解析 JWT 服务，失败时 panic。
func MustMakeJWT(c runtimecontract.Container) securitycontract.JWTService {
	return MustMakeJWTService(c)
}

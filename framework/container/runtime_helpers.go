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

// ============================================================
// Get 系列：返回 error，不 panic
// ============================================================

// GetDBRuntime resolves the database runtime object from the container.
//
// GetDBRuntime 从容器中解析数据库运行时对象。
func GetDBRuntime(c runtimecontract.Container) (any, error) {
	return c.Make(datacontract.DBRuntimeKey)
}

// GetRedis resolves the Redis capability from the container.
//
// GetRedis 从容器中解析 Redis 能力。
func GetRedis(c runtimecontract.Container) (datacontract.Redis, error) {
	return MakeWith[datacontract.Redis](c, datacontract.RedisKey)
}

// GetCache resolves the cache capability from the container.
//
// GetCache 从容器中解析缓存能力。
func GetCache(c runtimecontract.Container) (datacontract.Cache, error) {
	return MakeWith[datacontract.Cache](c, datacontract.CacheKey)
}

// GetGorm resolves the Gorm database handle from the container.
//
// GetGorm 从容器中解析 Gorm 数据库句柄。
func GetGorm(c runtimecontract.Container) (*gormdb.DB, error) {
	return MakeWith[*gormdb.DB](c, datacontract.GormKey)
}

// GetSQLX resolves the SQLX database handle from the container.
//
// GetSQLX 从容器中解析 SQLX 数据库句柄。
func GetSQLX(c runtimecontract.Container) (*sqlx.DB, error) {
	return MakeWith[*sqlx.DB](c, datacontract.SQLXKey)
}

// GetDB resolves both Gorm and SQLX database handles from the container.
//
// GetDB 从容器中解析 Gorm 和 SQLX 数据库句柄。
func GetDB(c runtimecontract.Container) (*gormdb.DB, *sqlx.DB, error) {
	gormDB, err := GetGorm(c)
	if err != nil {
		return nil, nil, err
	}
	sqlxDB, err := GetSQLX(c)
	if err != nil {
		return nil, nil, err
	}
	return gormDB, sqlxDB, nil
}

// GetMessagePublisher resolves the message publisher from the container.
//
// GetMessagePublisher 从容器中解析消息发布能力。
func GetMessagePublisher(c runtimecontract.Container) (integrationcontract.MessagePublisher, error) {
	return MakeWith[integrationcontract.MessagePublisher](c, integrationcontract.MessagePublisherKey)
}

// GetMessageSubscriber resolves the message subscriber from the container.
//
// GetMessageSubscriber 从容器中解析消息订阅能力。
func GetMessageSubscriber(c runtimecontract.Container) (integrationcontract.MessageSubscriber, error) {
	return MakeWith[integrationcontract.MessageSubscriber](c, integrationcontract.MessageSubscriberKey)
}

// GetDistributedLock resolves the distributed lock capability from the container.
//
// GetDistributedLock 从容器中解析分布式锁能力。
func GetDistributedLock(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	return MakeWith[datacontract.DistributedLock](c, datacontract.DistributedLockKey)
}

// GetGRPCConnFactory resolves the gRPC connection factory from the container.
//
// GetGRPCConnFactory 从容器中解析 gRPC 连接工厂。
func GetGRPCConnFactory(c runtimecontract.Container) (transportcontract.GRPCConnFactory, error) {
	return MakeWith[transportcontract.GRPCConnFactory](c, transportcontract.GRPCConnFactoryKey)
}

// GetGRPCServerRegistrar resolves the gRPC server registrar from the container.
//
// GetGRPCServerRegistrar 从容器中解析 gRPC 服务注册器。
func GetGRPCServerRegistrar(c runtimecontract.Container) (transportcontract.GRPCServerRegistrar, error) {
	return MakeWith[transportcontract.GRPCServerRegistrar](c, transportcontract.GRPCServerRegistrarKey)
}

// GetCron resolves the cron capability from the container.
//
// GetCron 从容器中解析定时任务能力。
func GetCron(c runtimecontract.Container) (runtimecontract.Cron, error) {
	return MakeWith[runtimecontract.Cron](c, runtimecontract.CronKey)
}

// GetLogger resolves the logger capability from the container.
//
// GetLogger 从容器中解析日志能力。
func GetLogger(c runtimecontract.Container) (observabilitycontract.Logger, error) {
	return MakeWith[observabilitycontract.Logger](c, observabilitycontract.LogKey)
}

// GetValidator resolves the validator capability from the container.
//
// GetValidator 从容器中解析参数校验能力。
func GetValidator(c runtimecontract.Container) (datacontract.Validator, error) {
	return MakeWith[datacontract.Validator](c, datacontract.ValidatorKey)
}

// GetRetry resolves the retry capability from the container.
//
// GetRetry 从容器中解析重试策略能力。
func GetRetry(c runtimecontract.Container) (resiliencecontract.Retry, error) {
	return MakeWith[resiliencecontract.Retry](c, resiliencecontract.RetryKey)
}

// GetHost resolves the host capability from the container.
//
// GetHost 从容器中解析 host 能力。
func GetHost(c runtimecontract.Container) (runtimecontract.Host, error) {
	return MakeWith[runtimecontract.Host](c, runtimecontract.HostKey)
}

// GetHTTP resolves the HTTP service from the container.
//
// GetHTTP 从容器中解析 HTTP 服务。
func GetHTTP(c runtimecontract.Container) (transportcontract.HTTP, error) {
	return MakeWith[transportcontract.HTTP](c, transportcontract.HTTPKey)
}

// GetHTTPRouter resolves the HTTP router facade from the container.
//
// GetHTTPRouter 从容器中解析 HTTP 路由门面。
func GetHTTPRouter(c runtimecontract.Container) (transportcontract.HTTPRouter, error) {
	httpSvc, err := GetHTTP(c)
	if err != nil {
		return nil, err
	}
	return httpSvc.Router(), nil
}

// GetJWT resolves the JWT service from the container.
//
// GetJWT 从容器中解析 JWT 服务。
func GetJWT(c runtimecontract.Container) (securitycontract.JWTService, error) {
	return MakeWith[securitycontract.JWTService](c, securitycontract.AuthJWTKey)
}

// ============================================================
// GetOrPanic 系列：失败时 panic
// ============================================================

// GetLoggerOrPanic resolves the logger and panics on failure.
//
// GetLoggerOrPanic 解析日志能力，失败时 panic。
func GetLoggerOrPanic(c runtimecontract.Container) observabilitycontract.Logger {
	return MustMakeWith[observabilitycontract.Logger](c, observabilitycontract.LogKey)
}

// GetGormOrPanic resolves the Gorm database handle and panics on failure.
//
// GetGormOrPanic 解析 Gorm 数据库句柄，失败时 panic。
func GetGormOrPanic(c runtimecontract.Container) *gormdb.DB {
	return MustMakeWith[*gormdb.DB](c, datacontract.GormKey)
}

// GetSQLXOrPanic resolves the SQLX database handle and panics on failure.
//
// GetSQLXOrPanic 解析 SQLX 数据库句柄，失败时 panic。
func GetSQLXOrPanic(c runtimecontract.Container) *sqlx.DB {
	return MustMakeWith[*sqlx.DB](c, datacontract.SQLXKey)
}

// GetDBOrPanic resolves both Gorm and SQLX database handles from the container.
// Panics if either is not available.
//
// GetDBOrPanic 从容器中解析 Gorm 和 SQLX 数据库句柄，失败时 panic。
// 这两个句柄指向同一个数据库连接，业务可根据场景选择使用。
func GetDBOrPanic(c runtimecontract.Container) (*gormdb.DB, *sqlx.DB) {
	return MustMakeWith[*gormdb.DB](c, datacontract.GormKey), MustMakeWith[*sqlx.DB](c, datacontract.SQLXKey)
}

// GetHTTPRouterOrPanic resolves the HTTP router facade and panics on failure.
//
// GetHTTPRouterOrPanic 解析 HTTP 路由门面，失败时 panic。
func GetHTTPRouterOrPanic(c runtimecontract.Container) transportcontract.HTTPRouter {
	httpSvc := GetHTTPOrPanic(c)
	return httpSvc.Router()
}

// GetHTTPOrPanic resolves the HTTP service and panics on failure.
//
// GetHTTPOrPanic 解析 HTTP 服务，失败时 panic。
func GetHTTPOrPanic(c runtimecontract.Container) transportcontract.HTTP {
	return MustMakeWith[transportcontract.HTTP](c, transportcontract.HTTPKey)
}

// GetMessagePublisherOrPanic resolves the message publisher and panics on failure.
//
// GetMessagePublisherOrPanic 解析消息发布能力，失败时 panic。
func GetMessagePublisherOrPanic(c runtimecontract.Container) integrationcontract.MessagePublisher {
	return MustMakeWith[integrationcontract.MessagePublisher](c, integrationcontract.MessagePublisherKey)
}

// GetMessageSubscriberOrPanic resolves the message subscriber and panics on failure.
//
// GetMessageSubscriberOrPanic 解析消息订阅能力，失败时 panic。
func GetMessageSubscriberOrPanic(c runtimecontract.Container) integrationcontract.MessageSubscriber {
	return MustMakeWith[integrationcontract.MessageSubscriber](c, integrationcontract.MessageSubscriberKey)
}

// GetDistributedLockOrPanic resolves the distributed lock capability and panics on failure.
//
// GetDistributedLockOrPanic 解析分布式锁能力，失败时 panic。
func GetDistributedLockOrPanic(c runtimecontract.Container) datacontract.DistributedLock {
	return MustMakeWith[datacontract.DistributedLock](c, datacontract.DistributedLockKey)
}

// GetGRPCConnFactoryOrPanic resolves the gRPC connection factory and panics on failure.
//
// GetGRPCConnFactoryOrPanic 解析 gRPC 连接工厂，失败时 panic。
func GetGRPCConnFactoryOrPanic(c runtimecontract.Container) transportcontract.GRPCConnFactory {
	return MustMakeWith[transportcontract.GRPCConnFactory](c, transportcontract.GRPCConnFactoryKey)
}

// GetGRPCServerRegistrarOrPanic resolves the gRPC server registrar and panics on failure.
//
// GetGRPCServerRegistrarOrPanic 解析 gRPC 服务注册器，失败时 panic。
func GetGRPCServerRegistrarOrPanic(c runtimecontract.Container) transportcontract.GRPCServerRegistrar {
	return MustMakeWith[transportcontract.GRPCServerRegistrar](c, transportcontract.GRPCServerRegistrarKey)
}

// GetValidatorOrPanic resolves the validator capability and panics on failure.
//
// GetValidatorOrPanic 解析参数校验能力，失败时 panic。
func GetValidatorOrPanic(c runtimecontract.Container) datacontract.Validator {
	return MustMakeWith[datacontract.Validator](c, datacontract.ValidatorKey)
}

// GetRetryOrPanic resolves the retry capability and panics on failure.
//
// GetRetryOrPanic 解析重试策略能力，失败时 panic。
func GetRetryOrPanic(c runtimecontract.Container) resiliencecontract.Retry {
	return MustMakeWith[resiliencecontract.Retry](c, resiliencecontract.RetryKey)
}

// GetConfigOrPanic resolves the config capability and panics on failure.
//
// GetConfigOrPanic 解析配置能力，失败时 panic。
func GetConfigOrPanic(c runtimecontract.Container) datacontract.Config {
	return MustMakeWith[datacontract.Config](c, datacontract.ConfigKey)
}

// GetCacheOrPanic resolves the cache capability and panics on failure.
//
// GetCacheOrPanic 解析缓存能力，失败时 panic。
func GetCacheOrPanic(c runtimecontract.Container) datacontract.Cache {
	return MustMakeWith[datacontract.Cache](c, datacontract.CacheKey)
}

// GetCronOrPanic resolves the cron capability and panics on failure.
//
// GetCronOrPanic 解析定时任务能力，失败时 panic。
func GetCronOrPanic(c runtimecontract.Container) runtimecontract.Cron {
	return MustMakeWith[runtimecontract.Cron](c, runtimecontract.CronKey)
}

// GetJWTOrPanic resolves the JWT service and panics on failure.
//
// GetJWTOrPanic 解析 JWT 服务，失败时 panic。
func GetJWTOrPanic(c runtimecontract.Container) securitycontract.JWTService {
	return MustMakeWith[securitycontract.JWTService](c, securitycontract.AuthJWTKey)
}

// GetMiddlewareRegistry resolves the middleware registry from the container.
//
// GetMiddlewareRegistry 从容器中解析中间件注册表。
func GetMiddlewareRegistry(c runtimecontract.Container) (transportcontract.MiddlewareRegistry, error) {
	return MakeWith[transportcontract.MiddlewareRegistry](c, transportcontract.MiddlewareRegistryKey)
}

// GetMiddlewareRegistryOrPanic resolves the middleware registry and panics on failure.
//
// GetMiddlewareRegistryOrPanic 解析中间件注册表，失败时 panic。
func GetMiddlewareRegistryOrPanic(c runtimecontract.Container) transportcontract.MiddlewareRegistry {
	return MustMakeWith[transportcontract.MiddlewareRegistry](c, transportcontract.MiddlewareRegistryKey)
}

// GetWebSocketServer resolves the WebSocket server from the container.
//
// GetWebSocketServer 从容器中解析 WebSocket 服务器。
func GetWebSocketServer(c runtimecontract.Container) (transportcontract.WebSocketServer, error) {
	return MakeWith[transportcontract.WebSocketServer](c, transportcontract.WebSocketKey)
}

// GetWebSocketServerOrPanic resolves the WebSocket server and panics on failure.
//
// GetWebSocketServerOrPanic 解析 WebSocket 服务器，失败时 panic。
func GetWebSocketServerOrPanic(c runtimecontract.Container) transportcontract.WebSocketServer {
	return MustMakeWith[transportcontract.WebSocketServer](c, transportcontract.WebSocketKey)
}

// GetServiceAuthenticator resolves the service authenticator from the container.
//
// GetServiceAuthenticator 从容器中解析服务认证器。
func GetServiceAuthenticator(c runtimecontract.Container) (securitycontract.ServiceAuthenticator, error) {
	return MakeWith[securitycontract.ServiceAuthenticator](c, securitycontract.ServiceAuthKey)
}

// GetServiceAuthenticatorOrPanic resolves the service authenticator and panics on failure.
//
// GetServiceAuthenticatorOrPanic 解析服务认证器，失败时 panic。
func GetServiceAuthenticatorOrPanic(c runtimecontract.Container) securitycontract.ServiceAuthenticator {
	return MustMakeWith[securitycontract.ServiceAuthenticator](c, securitycontract.ServiceAuthKey)
}

// GetServiceTokenIssuer resolves the service token issuer from the container.
//
// GetServiceTokenIssuer 从容器中解析服务令牌签发器。
func GetServiceTokenIssuer(c runtimecontract.Container) (securitycontract.ServiceTokenIssuer, error) {
	return MakeWith[securitycontract.ServiceTokenIssuer](c, securitycontract.ServiceAuthKey)
}

// GetServiceTokenIssuerOrPanic resolves the service token issuer and panics on failure.
//
// GetServiceTokenIssuerOrPanic 解析服务令牌签发器，失败时 panic。
func GetServiceTokenIssuerOrPanic(c runtimecontract.Container) securitycontract.ServiceTokenIssuer {
	return MustMakeWith[securitycontract.ServiceTokenIssuer](c, securitycontract.ServiceAuthKey)
}

// ============================================================
// 兼容别名（保留旧命名，内部调用新命名）
// ============================================================

// MakeDBRuntime is an alias for GetDBRuntime.
//
// MakeDBRuntime 是 GetDBRuntime 的别名。
func MakeDBRuntime(c runtimecontract.Container) (any, error) {
	return GetDBRuntime(c)
}

// MakeRedis is an alias for GetRedis.
//
// MakeRedis 是 GetRedis 的别名。
func MakeRedis(c runtimecontract.Container) (datacontract.Redis, error) {
	return GetRedis(c)
}

// MakeCache is an alias for GetCache.
//
// MakeCache 是 GetCache 的别名。
func MakeCache(c runtimecontract.Container) (datacontract.Cache, error) {
	return GetCache(c)
}

// MakeGormDB is an alias for GetGorm.
//
// MakeGormDB 是 GetGorm 的别名。
func MakeGormDB(c runtimecontract.Container) (*gormdb.DB, error) {
	return GetGorm(c)
}

// MakeSQLX is an alias for GetSQLX.
//
// MakeSQLX 是 GetSQLX 的别名。
func MakeSQLX(c runtimecontract.Container) (*sqlx.DB, error) {
	return GetSQLX(c)
}

// MakeMessagePublisher is an alias for GetMessagePublisher.
//
// MakeMessagePublisher 是 GetMessagePublisher 的别名。
func MakeMessagePublisher(c runtimecontract.Container) (integrationcontract.MessagePublisher, error) {
	return GetMessagePublisher(c)
}

// MakeMessageSubscriber is an alias for GetMessageSubscriber.
//
// MakeMessageSubscriber 是 GetMessageSubscriber 的别名。
func MakeMessageSubscriber(c runtimecontract.Container) (integrationcontract.MessageSubscriber, error) {
	return GetMessageSubscriber(c)
}

// MakeDistributedLock is an alias for GetDistributedLock.
//
// MakeDistributedLock 是 GetDistributedLock 的别名。
func MakeDistributedLock(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	return GetDistributedLock(c)
}

// MakeGRPCConnFactory is an alias for GetGRPCConnFactory.
//
// MakeGRPCConnFactory 是 GetGRPCConnFactory 的别名。
func MakeGRPCConnFactory(c runtimecontract.Container) (transportcontract.GRPCConnFactory, error) {
	return GetGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar is an alias for GetGRPCServerRegistrar.
//
// MakeGRPCServerRegistrar 是 GetGRPCServerRegistrar 的别名。
func MakeGRPCServerRegistrar(c runtimecontract.Container) (transportcontract.GRPCServerRegistrar, error) {
	return GetGRPCServerRegistrar(c)
}

// MakeCron is an alias for GetCron.
//
// MakeCron 是 GetCron 的别名。
func MakeCron(c runtimecontract.Container) (runtimecontract.Cron, error) {
	return GetCron(c)
}

// MakeLogger is an alias for GetLogger.
//
// MakeLogger 是 GetLogger 的别名。
func MakeLogger(c runtimecontract.Container) (observabilitycontract.Logger, error) {
	return GetLogger(c)
}

// MakeValidator is an alias for GetValidator.
//
// MakeValidator 是 GetValidator 的别名。
func MakeValidator(c runtimecontract.Container) (datacontract.Validator, error) {
	return GetValidator(c)
}

// MakeRetry is an alias for GetRetry.
//
// MakeRetry 是 GetRetry 的别名。
func MakeRetry(c runtimecontract.Container) (resiliencecontract.Retry, error) {
	return GetRetry(c)
}

// MakeHost is an alias for GetHost.
//
// MakeHost 是 GetHost 的别名。
func MakeHost(c runtimecontract.Container) (runtimecontract.Host, error) {
	return GetHost(c)
}

// MakeHTTP is an alias for GetHTTP.
//
// MakeHTTP 是 GetHTTP 的别名。
func MakeHTTP(c runtimecontract.Container) (transportcontract.HTTP, error) {
	return GetHTTP(c)
}

// MakeHTTPRouter is an alias for GetHTTPRouter.
//
// MakeHTTPRouter 是 GetHTTPRouter 的别名。
func MakeHTTPRouter(c runtimecontract.Container) (transportcontract.HTTPRouter, error) {
	return GetHTTPRouter(c)
}

// MustMakeLogger is an alias for GetLoggerOrPanic.
//
// MustMakeLogger 是 GetLoggerOrPanic 的别名。
func MustMakeLogger(c runtimecontract.Container) observabilitycontract.Logger {
	return GetLoggerOrPanic(c)
}

// MustMakeGorm is an alias for GetGormOrPanic.
//
// MustMakeGorm 是 GetGormOrPanic 的别名。
func MustMakeGorm(c runtimecontract.Container) *gormdb.DB {
	return GetGormOrPanic(c)
}

// MustMakeGormDB is an alias for GetGormOrPanic.
//
// MustMakeGormDB 是 GetGormOrPanic 的别名。
func MustMakeGormDB(c runtimecontract.Container) *gormdb.DB {
	return GetGormOrPanic(c)
}

// MustMakeSQLX is an alias for GetSQLXOrPanic.
//
// MustMakeSQLX 是 GetSQLXOrPanic 的别名。
func MustMakeSQLX(c runtimecontract.Container) *sqlx.DB {
	return GetSQLXOrPanic(c)
}

// MustMakeDB is an alias for GetDBOrPanic.
//
// MustMakeDB 是 GetDBOrPanic 的别名。
func MustMakeDB(c runtimecontract.Container) (*gormdb.DB, *sqlx.DB) {
	return GetDBOrPanic(c)
}

// MustMakeHTTPRouter is an alias for GetHTTPRouterOrPanic.
//
// MustMakeHTTPRouter 是 GetHTTPRouterOrPanic 的别名。
func MustMakeHTTPRouter(c runtimecontract.Container) transportcontract.HTTPRouter {
	return GetHTTPRouterOrPanic(c)
}

// MustMakeHTTP is an alias for GetHTTPOrPanic.
//
// MustMakeHTTP 是 GetHTTPOrPanic 的别名。
func MustMakeHTTP(c runtimecontract.Container) transportcontract.HTTP {
	return GetHTTPOrPanic(c)
}

// MustMakeMessagePublisher is an alias for GetMessagePublisherOrPanic.
//
// MustMakeMessagePublisher 是 GetMessagePublisherOrPanic 的别名。
func MustMakeMessagePublisher(c runtimecontract.Container) integrationcontract.MessagePublisher {
	return GetMessagePublisherOrPanic(c)
}

// MustMakeMessageSubscriber is an alias for GetMessageSubscriberOrPanic.
//
// MustMakeMessageSubscriber 是 GetMessageSubscriberOrPanic 的别名。
func MustMakeMessageSubscriber(c runtimecontract.Container) integrationcontract.MessageSubscriber {
	return GetMessageSubscriberOrPanic(c)
}

// MustMakeDistributedLock is an alias for GetDistributedLockOrPanic.
//
// MustMakeDistributedLock 是 GetDistributedLockOrPanic 的别名。
func MustMakeDistributedLock(c runtimecontract.Container) datacontract.DistributedLock {
	return GetDistributedLockOrPanic(c)
}

// MustMakeGRPCConnFactory is an alias for GetGRPCConnFactoryOrPanic.
//
// MustMakeGRPCConnFactory 是 GetGRPCConnFactoryOrPanic 的别名。
func MustMakeGRPCConnFactory(c runtimecontract.Container) transportcontract.GRPCConnFactory {
	return GetGRPCConnFactoryOrPanic(c)
}

// MustMakeGRPCServerRegistrar is an alias for GetGRPCServerRegistrarOrPanic.
//
// MustMakeGRPCServerRegistrar 是 GetGRPCServerRegistrarOrPanic 的别名。
func MustMakeGRPCServerRegistrar(c runtimecontract.Container) transportcontract.GRPCServerRegistrar {
	return GetGRPCServerRegistrarOrPanic(c)
}

// MustMakeValidator is an alias for GetValidatorOrPanic.
//
// MustMakeValidator 是 GetValidatorOrPanic 的别名。
func MustMakeValidator(c runtimecontract.Container) datacontract.Validator {
	return GetValidatorOrPanic(c)
}

// MustMakeRetry is an alias for GetRetryOrPanic.
//
// MustMakeRetry 是 GetRetryOrPanic 的别名。
func MustMakeRetry(c runtimecontract.Container) resiliencecontract.Retry {
	return GetRetryOrPanic(c)
}

// MustMakeConfig is an alias for GetConfigOrPanic.
//
// MustMakeConfig 是 GetConfigOrPanic 的别名。
func MustMakeConfig(c runtimecontract.Container) datacontract.Config {
	return GetConfigOrPanic(c)
}

// MustMakeCache is an alias for GetCacheOrPanic.
//
// MustMakeCache 是 GetCacheOrPanic 的别名。
func MustMakeCache(c runtimecontract.Container) datacontract.Cache {
	return GetCacheOrPanic(c)
}

// MustMakeCron is an alias for GetCronOrPanic.
//
// MustMakeCron 是 GetCronOrPanic 的别名。
func MustMakeCron(c runtimecontract.Container) runtimecontract.Cron {
	return GetCronOrPanic(c)
}

// MustMakeJWT is an alias for GetJWTOrPanic.
//
// MustMakeJWT 是 GetJWTOrPanic 的别名。
func MustMakeJWT(c runtimecontract.Container) securitycontract.JWTService {
	return GetJWTOrPanic(c)
}

// ============================================================
// 其他辅助函数
// ============================================================

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
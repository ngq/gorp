package container

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ngq/gorp/framework/contract"
	gormdb "gorm.io/gorm"
)

// MakeDBRuntime 获取统一数据库运行时实例。
//
// 中文说明：
// - 优先使用 contract.DBRuntimeKey，屏蔽业务层直接感知多 backend 细节；
// - 返回 any，由调用方根据需要做最小类型判断；
// - 适合健康检查、通用探针、低层 runtime 诊断场景。
func MakeDBRuntime(c contract.Container) (any, error) {
	return c.Make(contract.DBRuntimeKey)
}

// MakeRedis 获取 Redis 服务，失败返回 error。
func MakeRedis(c contract.Container) (contract.Redis, error) {
	v, err := c.Make(contract.RedisKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.Redis), nil
}

// MakeCache 获取统一缓存服务，失败返回 error。
//
// 中文说明：
// - 优先暴露 contract.Cache，而不是让业务层自己决定 memory/redis driver；
// - 适合作为 starter 默认缓存接入位；
// - 与 MakeRedis 并存：需要 Redis 原语时拿 Redis，需要缓存语义时拿 Cache。
func MakeCache(c contract.Container) (contract.Cache, error) {
	v, err := c.Make(contract.CacheKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.Cache), nil
}

// MakeGormDB 获取 GORM 实例，失败返回 error。
func MakeGormDB(c contract.Container) (*gormdb.DB, error) {
	v, err := c.Make(contract.GormKey)
	if err != nil {
		return nil, err
	}
	return v.(*gormdb.DB), nil
}

// MakeSQLX 获取 SQLX 实例，失败返回 error。
func MakeSQLX(c contract.Container) (*sqlx.DB, error) {
	v, err := c.Make(contract.SQLXKey)
	if err != nil {
		return nil, err
	}
	return v.(*sqlx.DB), nil
}

// MakeMessagePublisher 获取消息发布者，失败返回 error。
//
// 中文说明：
// - 用于业务侧最小接入消息发布能力；
// - 让样板不需要直接记忆 `MessagePublisherKey`；
// - 适合 starter / 模板作为默认 publish 入口。
func MakeMessagePublisher(c contract.Container) (contract.MessagePublisher, error) {
	v, err := c.Make(contract.MessagePublisherKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.MessagePublisher), nil
}

// MakeMessageSubscriber 获取消息订阅者，失败返回 error。
//
// 中文说明：
// - 用于业务侧最小接入消息消费能力；
// - 让样板不需要直接记忆 `MessageSubscriberKey`；
// - 适合 starter / 模板作为默认 consume 入口。
func MakeMessageSubscriber(c contract.Container) (contract.MessageSubscriber, error) {
	v, err := c.Make(contract.MessageSubscriberKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.MessageSubscriber), nil
}

// MakeDistributedLock 获取分布式锁服务，失败返回 error。
//
// 中文说明：
// - 用于业务侧最小接入分布式锁能力；
// - 让样板不需要直接记忆 `DistributedLockKey`；
// - 适合 starter / 模板作为默认锁语义入口。
func MakeDistributedLock(c contract.Container) (contract.DistributedLock, error) {
	v, err := c.Make(contract.DistributedLockKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.DistributedLock), nil
}

// MakeGRPCConnFactory 获取 Proto-first gRPC 连接工厂，失败返回 error。
//
// 中文说明：
// - 业务侧可通过它按服务名获取 framework 托管的 `*grpc.ClientConn`；
// - 这是 Proto-first gRPC 客户端主线的标准入口；
// - 与旧统一 `RPCClient` 相比，这里直接返回连接，便于继续使用 `pb.NewXxxClient(conn)`。
func MakeGRPCConnFactory(c contract.Container) (contract.GRPCConnFactory, error) {
	v, err := c.Make(contract.GRPCConnFactoryKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.GRPCConnFactory), nil
}

// MakeGRPCServerRegistrar 获取 Proto-first gRPC 服务端注册器，失败返回 error。
//
// 中文说明：
// - 业务侧可通过它把 `pb.RegisterXxxServer(...)` 挂到 framework 托管的 `grpc.Server`；
// - 这是 Proto-first gRPC 服务端主线的标准入口；
// - 与旧统一 `RPCServer` 相比，这里直接表达标准 gRPC register 心智。
func MakeGRPCServerRegistrar(c contract.Container) (contract.GRPCServerRegistrar, error) {
	v, err := c.Make(contract.GRPCServerRegistrarKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.GRPCServerRegistrar), nil
}

// MakeCron 获取 Cron 服务，失败返回 error。
func MakeCron(c contract.Container) (contract.Cron, error) {
	v, err := c.Make(contract.CronKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.Cron), nil
}

// MakeLogger 获取日志服务，失败返回 error。
func MakeLogger(c contract.Container) (contract.Logger, error) {
	v, err := c.Make(contract.LogKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.Logger), nil
}

// MakeHost 获取 Host 服务，失败返回 error。
//
// 中文说明：
// - 用于获取框架级 Host 能力（生命周期管理）；
// - 与 MustMakeHost 相比，失败时返回 error 而不 panic。
func MakeHost(c contract.Container) (contract.Host, error) {
	v, err := c.Make(contract.HostKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.Host), nil
}

// MakeHTTP 获取 HTTP 服务，失败返回 error。
//
// 中文说明：
// - 用于获取框架级 HTTP 能力（Gin 服务封装）；
// - 与 MustMakeHTTP 相比，失败时返回 error 而不 panic。
func MakeHTTP(c contract.Container) (contract.HTTP, error) {
	v, err := c.Make(contract.HTTPKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.HTTP), nil
}

// MakeGinEngine 获取 Gin Engine，失败返回 error。
//
// 中文说明：
// - 直接获取 Gin Engine，用于路由注册等场景；
// - 与 MustMakeEngine 相比，失败时返回 error 而不 panic。
func MakeGinEngine(c contract.Container) (*gin.Engine, error) {
	v, err := c.Make(contract.HTTPEngineKey)
	if err != nil {
		return nil, err
	}
	return v.(*gin.Engine), nil
}

// MustMakeLogger 获取日志服务，失败 panic。
//
// 中文说明：
// - 适用于启动阶段，缺少日志说明配置有误；
// - 封装 container.MustMake，便于业务调用。
func MustMakeLogger(c contract.Container) contract.Logger {
	v := c.MustMake(contract.LogKey)
	return v.(contract.Logger)
}

// MustMakeGorm 获取 GORM 实例，失败 panic。
//
// 中文说明：
// - 适用于启动阶段，缺少数据库说明配置有误；
// - 封装 container.MustMake，便于业务调用。
func MustMakeGorm(c contract.Container) *gormdb.DB {
	v := c.MustMake(contract.GormKey)
	return v.(*gormdb.DB)
}

// MustMakeEngine 获取 Gin Engine，失败 panic。
//
// 中文说明：
// - 适用于启动阶段，缺少 Engine 说明配置有误；
// - 封装 container.MustMake，便于业务调用。
func MustMakeEngine(c contract.Container) *gin.Engine {
	v := c.MustMake(contract.HTTPEngineKey)
	return v.(*gin.Engine)
}

// MustMakeMessagePublisher 获取消息发布者，失败 panic。
//
// 中文说明：
// - 适用于启动阶段或明确要求 MQ 发布能力已接入的路径；
// - 样板里推荐用它表达“当前服务具备发布能力”。
func MustMakeMessagePublisher(c contract.Container) contract.MessagePublisher {
	v := c.MustMake(contract.MessagePublisherKey)
	return v.(contract.MessagePublisher)
}

// MustMakeMessageSubscriber 获取消息订阅者，失败 panic。
//
// 中文说明：
// - 适用于启动阶段或明确要求 MQ 消费能力已接入的路径；
// - 样板里推荐用它表达“当前服务具备消费能力”。
func MustMakeMessageSubscriber(c contract.Container) contract.MessageSubscriber {
	v := c.MustMake(contract.MessageSubscriberKey)
	return v.(contract.MessageSubscriber)
}

// MustMakeDistributedLock 获取分布式锁服务，失败 panic。
//
// 中文说明：
// - 适用于启动阶段或明确要求锁能力已接入的路径；
// - 样板里推荐用它表达“当前服务具备锁语义能力”。
func MustMakeDistributedLock(c contract.Container) contract.DistributedLock {
	v := c.MustMake(contract.DistributedLockKey)
	return v.(contract.DistributedLock)
}

// MustMakeGRPCConnFactory 获取 Proto-first gRPC 连接工厂，失败 panic。
//
// 中文说明：
// - 适用于启动阶段或明确要求 gRPC 主线能力已接入的路径；
// - 拿到后可直接配合 `pb.NewXxxClient(conn)` 使用。
func MustMakeGRPCConnFactory(c contract.Container) contract.GRPCConnFactory {
	v := c.MustMake(contract.GRPCConnFactoryKey)
	return v.(contract.GRPCConnFactory)
}

// MustMakeGRPCServerRegistrar 获取 Proto-first gRPC 服务端注册器，失败 panic。
//
// 中文说明：
// - 适用于启动阶段或项目明确要求 gRPC 服务注册能力已接入的路径；
// - 业务注册应优先通过 `RegisterProto(...)` 完成。
func MustMakeGRPCServerRegistrar(c contract.Container) contract.GRPCServerRegistrar {
	v := c.MustMake(contract.GRPCServerRegistrarKey)
	return v.(contract.GRPCServerRegistrar)
}

// MustMakeConfig 获取 Config，失败 panic。
//
// 中文说明：
// - 适用于启动阶段，缺少配置说明配置有误；
// - 封装 container.MustMake，便于业务调用。
func MustMakeConfig(c contract.Container) contract.Config {
	v := c.MustMake(contract.ConfigKey)
	return v.(contract.Config)
}

// MustMakeCache 获取统一缓存服务，失败 panic。
//
// 中文说明：
// - 适用于启动阶段或明确要求 cache 已接入的业务路径；
// - starter/project 若把 cache 作为默认起步能力，可直接复用。
func MustMakeCache(c contract.Container) contract.Cache {
	v := c.MustMake(contract.CacheKey)
	return v.(contract.Cache)
}

// PingDBRuntime 对统一数据库运行时做最小 ping。
//
// 中文说明：
// - 给 framework 探针与默认接入路径使用；
// - 这样业务层不必自己到处复制相同的类型分支判断。
func PingDBRuntime(dbAny any) error {
	switch db := dbAny.(type) {
	case *gormdb.DB:
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	case *sql.DB:
		return db.Ping()
	case interface{ Ping() error }:
		return db.Ping()
	default:
		return nil
	}
}

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

// MustMakeConfig 获取 Config，失败 panic。
//
// 中文说明：
// - 适用于启动阶段，缺少配置说明配置有误；
// - 封装 container.MustMake，便于业务调用。
func MustMakeConfig(c contract.Container) contract.Config {
	v := c.MustMake(contract.ConfigKey)
	return v.(contract.Config)
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

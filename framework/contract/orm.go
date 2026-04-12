package contract

import (
	"context"
	"database/sql"
)

const (
	GormKey = "framework.orm.gorm"
	SQLXKey = "framework.orm.sqlx"

	// ORMBackendKey 暴露当前 database.backend 的选择结果。
	//
	// 中文说明：
	// - 这是“当前项目声明想使用哪种 runtime backend”的轻量入口；
	// - 第一阶段只先提供 backend 名称与能力 key，不强行统一 GORM / SQLX / Ent 的查询 API。
	ORMBackendKey = "framework.orm.backend"

	// DBRuntimeKey 暴露当前 runtime backend 对应的主对象。
	//
	// 中文说明：
	// - 当前阶段默认仍然返回 GORM 实例；
	// - 后续接入 ent 时，这里可以改成根据 backend 返回 *ent.Client 等具体对象；
	// - 上层新代码应优先依赖更细粒度的 capability / repository，而不是长期直接依赖这个 key。
	DBRuntimeKey = "framework.orm.runtime"

	// EntClientKey 暴露 ent backend 的 runtime client 占位入口。
	//
	// 中文说明：
	// - 当前仓库还没有真正引入 ent 依赖，因此这里只先给出正式 key；
	// - 后续项目内或框架内接入 ent client 时，应由 orm.ent provider 返回真实 client 实例。
	EntClientKey = "framework.orm.ent"

	// EntClientFactoryKey 允许项目侧注入 ent client 创建逻辑。
	//
	// 中文说明：
	// - 框架层不直接依赖 ent，因此不持有 `*ent.Client` 类型；
	// - 当 `database.backend=ent` 时，orm.ent provider 会优先解析这个 factory key；
	// - 这样真正的 ent 依赖由业务项目自己决定是否引入。
	EntClientFactoryKey = "framework.orm.ent.factory"

	// MigratorKey 暴露当前 backend 的迁移能力。
	MigratorKey = "framework.orm.migrator"

	// SQLExecutorKey 暴露当前 backend 的原生 SQL 执行能力。
	SQLExecutorKey = "framework.orm.sql_executor"
)

type DBConfig struct {
	Driver  string `mapstructure:"driver"`
	Backend string `mapstructure:"backend"`
	DSN     string `mapstructure:"dsn"`

	MaxOpenConns int `mapstructure:"max_open_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	// ConnMaxLifetime/ConnMaxIdleTime accept Go duration strings, e.g. "1h", "30m", "10s".
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idletime"`
}

// RuntimeBackend 描述当前运行时数据库访问后端。
type RuntimeBackend string

const (
	RuntimeBackendGorm RuntimeBackend = "gorm"
	RuntimeBackendSQLX RuntimeBackend = "sqlx"
	RuntimeBackendEnt  RuntimeBackend = "ent"
)

// NormalizeBackendName 统一 backend 名称，并为旧配置提供默认值。
func NormalizeBackendName(name string) RuntimeBackend {
	switch name {
	case "", string(RuntimeBackendGorm):
		return RuntimeBackendGorm
	case string(RuntimeBackendSQLX):
		return RuntimeBackendSQLX
	case string(RuntimeBackendEnt):
		return RuntimeBackendEnt
	default:
		// 中文说明：
		// - 第一阶段先对未知值宽容处理，统一退回 gorm，避免直接破坏现有启动链路；
		// - 真正启用 ent/sqlx-only 前，再在更高层补显式校验与错误提示。
		return RuntimeBackendGorm
	}
}

// EntClientFactory 定义“由项目侧创建 ent runtime client”的能力。
//
// 中文说明：
// - 返回值使用 `any`，是因为框架层不直接 import ent；
// - 业务项目可以返回 `*ent.Client`，也可以返回自己包一层的 runtime 对象；
// - orm.ent provider 只负责把这个对象挂到 `EntClientKey / DBRuntimeKey` 上。
type EntClientFactory interface {
	CreateEntClient(c Container) (any, error)
}

// Migrator 抽象“当前 backend 是否支持迁移”。
//
// 中文说明：
// - 第一阶段先只定义能力接口，不强行把 GORM / Ent / SQL migration 做成一套查询 API；
// - GORM 可以自然实现为 AutoMigrate；后续 Ent 可映射到自己的 schema migrate。
type Migrator interface {
	AutoMigrate(dst ...any) error
}

// SQLExecutor 抽象原生 SQL 执行能力。
//
// 中文说明：
// - 主要服务于 SQLX / schema inspect / 报表查询这类更偏原生 SQL 的路径；
// - 当前先用最小接口，后续如果 repository/query service 收口后仍有需要，再补 Query 能力。
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type GormDB interface {
	WithContext(ctx context.Context) any
}

type SQLX interface {
	ExecContext(ctx context.Context, query string, args ...any) (any, error)
}

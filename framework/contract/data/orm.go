// Application scenarios:
// - Define shared database and ORM contracts used by runtime providers.
// - Standardize backend selection keys and minimal database execution capabilities.
// - Keep Gorm, SQLX, and Ent integrations behind stable framework-level contracts.
//
// 适用场景：
// - 定义运行时 provider 使用的共享数据库与 ORM 契约。
// - 统一后端选择键和最小数据库执行能力。
// - 让 Gorm、SQLX 和 Ent 集成稳定地收敛在框架级契约之后。
package data

import (
	"context"
	"database/sql"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

const (
	GormKey             = "framework.orm.gorm"
	SQLXKey             = "framework.orm.sqlx"
	ORMBackendKey       = "framework.orm.backend"
	DBRuntimeKey        = "framework.orm.runtime"
	EntClientKey        = "framework.orm.ent"
	EntClientFactoryKey = "framework.orm.ent.factory"
	MigratorKey         = "framework.orm.migrator"
	SQLExecutorKey      = "framework.orm.sql_executor"
)

// DBConfig describes database connection settings.
//
// DBConfig 描述数据库连接配置。
type DBConfig struct {
	Driver  string `mapstructure:"driver"`
	Backend string `mapstructure:"backend"`
	DSN     string `mapstructure:"dsn"`

	MaxOpenConns int `mapstructure:"max_open_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime string `mapstructure:"conn_max_idletime"`
}

// RuntimeBackend identifies the database runtime backend.
//
// RuntimeBackend 标识数据库运行时后端类型。
type RuntimeBackend string

const (
	RuntimeBackendGorm RuntimeBackend = "gorm"
	RuntimeBackendSQLX RuntimeBackend = "sqlx"
	RuntimeBackendEnt  RuntimeBackend = "ent"
)

// NormalizeBackendName normalizes a backend name into a supported runtime backend.
//
// NormalizeBackendName 将后端名称归一化为受支持的运行时后端。
func NormalizeBackendName(name string) RuntimeBackend {
	switch name {
	case "", string(RuntimeBackendGorm):
		// Default to Gorm so an empty backend still lands on the mainstream ORM path.
		// 默认回退到 Gorm，保证空后端名仍能落到主流 ORM 路径。
		return RuntimeBackendGorm
	case string(RuntimeBackendSQLX):
		return RuntimeBackendSQLX
	case string(RuntimeBackendEnt):
		return RuntimeBackendEnt
	default:
		// Unknown backends are normalized conservatively to Gorm for compatibility.
		// 未知后端名保守归一化为 Gorm，以保持兼容行为。
		return RuntimeBackendGorm
	}
}

// EntClientFactory defines how an Ent client should be created from the container.
//
// EntClientFactory 定义如何从容器构建 Ent client。
type EntClientFactory interface {
	// CreateEntClient creates an Ent client from the current container.
	//
	// CreateEntClient 从当前容器创建 Ent client。
	CreateEntClient(c runtimecontract.Container) (any, error)
}

// Migrator defines the minimal schema migration capability.
//
// Migrator 定义最小化的结构迁移能力。
type Migrator interface {
	// AutoMigrate executes schema auto migration for the given models.
	//
	// AutoMigrate 对给定模型执行自动迁移。
	//
	// ⚠ 适用范围：仅限开发/测试环境。生产环境请使用版本化迁移工具（如 golang-migrate）。
	// 原因：AutoMigrate 直接 ALTER TABLE，存在丢数据、锁表风险，无法回滚，无法追踪历史。
	// 开发期：改 model → 重启服务 → 表结构自动同步，快速迭代。
	// 生产期：用 golang-migrate 管理版本化迁移文件，review SQL，支持 up/down。
	AutoMigrate(dst ...any) error
}

// SQLExecutor defines the minimal SQL execution capability needed by the framework.
//
// SQLExecutor 定义框架所需的最小 SQL 执行能力。
type SQLExecutor interface {
	// ExecContext executes a SQL statement in context.
	//
	// ExecContext 在指定 context 中执行 SQL 语句。
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// GormDB defines the minimal Gorm context-binding capability.
//
// GormDB 定义最小 Gorm context 绑定能力。
type GormDB interface {
	// WithContext returns a context-bound Gorm handle.
	//
	// WithContext 返回绑定 context 的 Gorm 句柄。
	WithContext(ctx context.Context) any
}

// SQLX defines the minimal SQLX execution capability.
//
// SQLX 定义最小 SQLX 执行能力。
type SQLX interface {
	// ExecContext executes a SQL statement through SQLX.
	//
	// ExecContext 通过 SQLX 执行 SQL 语句。
	ExecContext(ctx context.Context, query string, args ...any) (any, error)
}

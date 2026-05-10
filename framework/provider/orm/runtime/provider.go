// Package runtime provides ORM backend abstraction and unified runtime services.
// This provider binds multiple contracts based on the configured backend:
//
// ORM 运行时包，提供 ORM 后端抽象和统一的运行时服务。
// 此 Provider 根据配置的后端绑定多个契约：
//   - ORMBackendKey: The backend name (gorm/sqlx/ent)
//   - DBRuntimeKey: The actual DB client (GORM/SQLX/Ent)
//   - MigratorKey: The migrator (GORM, except Ent)
//   - SQLExecutorKey: The SQL executor (SQLX)
//
// Configuration via config.yaml:
//
// 通过 config.yaml 配置：
//
//	database:
//	  backend: gorm  # or sqlx, ent
//
// Eg:
//
//	// 注册 Provider
//	app.Register(runtime.NewProvider())
//
//	// 获取统一运行时服务
//	backend := c.MustMake(datacontract.ORMBackendKey).(string)
//	dbRuntime := c.MustMake(datacontract.DBRuntimeKey)
package runtime

import (
	"os"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers ORM backend abstraction and unified runtime services.
//
// Provider 注册 ORM 后端抽象和统一的运行时服务。
type Provider struct{}

// NewProvider creates a new runtime provider instance.
//
// NewProvider 创建新的运行时 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "orm.runtime".
//
// Name 返回 Provider 名称 "orm.runtime"。
func (p *Provider) Name() string { return "orm.runtime" }

// IsDefer returns false, runtime should be initialized immediately.
//
// IsDefer 返回 false，运行时服务应立即初始化。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the list of contracts this provider provides.
//
// Provides 返回此 Provider 提供的契约列表。
func (p *Provider) Provides() []string {
	return []string{
		datacontract.ORMBackendKey,
		datacontract.DBRuntimeKey,
		datacontract.MigratorKey,
		datacontract.SQLExecutorKey,
	}
}

// Register binds multiple factories to the container based on backend config.
// Core logic: Determine backend from config, then bind appropriate implementations.
//
// Register 根据后端配置将多个工厂绑定到容器。
// 核心逻辑：从配置确定后端，然后绑定相应的实现。
func (p *Provider) Register(c runtimecontract.Container) error {
	// Bind ORMBackendKey: resolve backend name from config, default to "gorm".
	//
	// 绑定 ORMBackendKey：从配置解析后端名称，默认为 "gorm"。
	c.Bind(datacontract.ORMBackendKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return string(datacontract.RuntimeBackendGorm), nil
		}
		cfg, ok := cfgAny.(datacontract.Config)
		if !ok {
			return string(datacontract.RuntimeBackendGorm), nil
		}

		var dbc datacontract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return string(datacontract.RuntimeBackendGorm), nil
		}
		return string(datacontract.NormalizeBackendName(dbc.Backend)), nil
	}, true)

	// Bind DBRuntimeKey: return appropriate DB client based on backend.
	//
	// 绑定 DBRuntimeKey：根据后端返回适当的数据库客户端。
	c.Bind(datacontract.DBRuntimeKey, func(c runtimecontract.Container) (any, error) {
		backendAny, err := c.Make(datacontract.ORMBackendKey)
		if err != nil {
			return c.Make(datacontract.GormKey)
		}
		switch datacontract.NormalizeBackendName(backendAny.(string)) {
		case datacontract.RuntimeBackendSQLX:
			return c.Make(datacontract.SQLXKey)
		case datacontract.RuntimeBackendEnt:
			return c.Make(datacontract.EntClientKey)
		case datacontract.RuntimeBackendGorm:
			fallthrough
		default:
			return c.Make(datacontract.GormKey)
		}
	}, true)

	// Bind MigratorKey: return GORM for migration (Ent has no migrator).
	//
	// 绑定 MigratorKey：返回 GORM 用于迁移（Ent 无迁移器）。
	c.Bind(datacontract.MigratorKey, func(c runtimecontract.Container) (any, error) {
		backendAny, err := c.Make(datacontract.ORMBackendKey)
		if err == nil && datacontract.NormalizeBackendName(backendAny.(string)) == datacontract.RuntimeBackendEnt {
			return nil, os.ErrInvalid
		}
		return c.Make(datacontract.GormKey)
	}, true)

	// Bind SQLExecutorKey: always use SQLX for raw SQL execution.
	//
	// 绑定 SQLExecutorKey：始终使用 SQLX 进行原生 SQL 执行。
	c.Bind(datacontract.SQLExecutorKey, func(c runtimecontract.Container) (any, error) {
		return c.Make(datacontract.SQLXKey)
	}, true)

	return nil
}

// Boot is a no-op for runtime provider.
//
// Boot 运行时 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
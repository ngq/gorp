package runtime

import (
	"os"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 ORM runtime capability 绑定。
//
// 中文说明：
// - 统一在 framework 层绑定 ORMBackendKey / DBRuntimeKey / MigratorKey / SQLExecutorKey；
// - 这样 cmd 层不再重复做 capability 绑定，只负责注册 provider；
// - 当前行为与历史实现保持一致：
//  1. backend 默认回退 gorm
//  2. ent backend 暂不提供 migrator（返回 os.ErrInvalid）
//  3. SQL 执行能力统一走 sqlx。
type Provider struct{}

// NewProvider 创建 ORM runtime provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "orm.runtime" }

// IsDefer 表示 ORM runtime 在应用启动期直接注册。
func (p *Provider) IsDefer() bool { return false }

// Provides 返回 ORM runtime 对外暴露的统一能力 key。
func (p *Provider) Provides() []string {
	return []string{
		contract.ORMBackendKey,
		contract.DBRuntimeKey,
		contract.MigratorKey,
		contract.SQLExecutorKey,
	}
}

// Register 绑定 ORM runtime 统一能力。
//
// 中文说明：
// - ORMBackendKey 负责根据 database.backend 解析当前主后端；
// - DBRuntimeKey 负责把业务真正拿到的运行时对象收敛成一个入口；
// - MigratorKey 与 SQLExecutorKey 则把迁移、执行 SQL 的常见入口统一成 capability；
// - 这样业务和 starter 都不必再自己分支判断 gorm / sqlx / ent。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ORMBackendKey, func(c contract.Container) (any, error) {
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return string(contract.RuntimeBackendGorm), nil
		}
		cfg, ok := cfgAny.(contract.Config)
		if !ok {
			return string(contract.RuntimeBackendGorm), nil
		}

		var dbc contract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return string(contract.RuntimeBackendGorm), nil
		}
		return string(contract.NormalizeBackendName(dbc.Backend)), nil
	}, true)

	c.Bind(contract.DBRuntimeKey, func(c contract.Container) (any, error) {
		backendAny, err := c.Make(contract.ORMBackendKey)
		if err != nil {
			return c.Make(contract.GormKey)
		}
		switch contract.NormalizeBackendName(backendAny.(string)) {
		case contract.RuntimeBackendSQLX:
			return c.Make(contract.SQLXKey)
		case contract.RuntimeBackendEnt:
			return c.Make(contract.EntClientKey)
		case contract.RuntimeBackendGorm:
			fallthrough
		default:
			return c.Make(contract.GormKey)
		}
	}, true)

	c.Bind(contract.MigratorKey, func(c contract.Container) (any, error) {
		backendAny, err := c.Make(contract.ORMBackendKey)
		if err == nil && contract.NormalizeBackendName(backendAny.(string)) == contract.RuntimeBackendEnt {
			return nil, os.ErrInvalid
		}
		return c.Make(contract.GormKey)
	}, true)

	c.Bind(contract.SQLExecutorKey, func(c contract.Container) (any, error) {
		return c.Make(contract.SQLXKey)
	}, true)

	return nil
}

// Boot ORM runtime provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }

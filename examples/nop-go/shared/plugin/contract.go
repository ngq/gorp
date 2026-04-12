// Package plugin nop-go 插件系统核心接口
//
// 中文说明:
// - 插件机制属于产品层设计,框架层 ServiceProvider + Container 已足够;
// - 所有业务插件都实现 Plugin 接口;
// - 通过 ToServiceProvider() 转换后注册到 gorp Container;
// - 复用框架已有的 ServiceProvider 生命周期管理。
package plugin

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Plugin 产品级插件基础接口
//
// 中文说明:
// - 所有业务插件(payment/shipping/widget等)都必须实现这个接口;
// - 插件通过 ToServiceProvider() 转换为 gorp ServiceProvider;
// - 这样可以复用框架已有的 ServiceProvider 生命周期管理(Register/Boot);
// - Install/Uninstall 用于数据库迁移和初始化配置。
type Plugin interface {
	// Meta 返回插件元数据
	//
	// 中文说明:
	// - 元数据从 plugin.json 加载,包含插件基本信息;
	// - 包含: Group/FriendlyName/SystemName/Version 等。
	Meta() *PluginMeta

	// PluginType 返回插件类型
	//
	// 中文说明:
	// - 类型用于分类管理,例如: "payment", "shipping", "widget";
	// - Registry 可按类型查找插件。
	PluginType() string

	// Install 插件安装时执行
	//
	// 中文说明:
	// - 用于创建数据库表、初始化配置、写入默认数据;
	// - 通常在首次启用插件时调用一次;
	// - 应包含数据库迁移逻辑。
	Install(ctx context.Context, c contract.Container) error

	// Uninstall 插件卸载时执行
	//
	// 中文说明:
	// - 用于清理数据、删除表(谨慎操作);
	// - 用户明确卸载时调用;
	// - 通常建议保留数据,只标记为已卸载。
	Uninstall(ctx context.Context, c contract.Container) error

	// Boot 插件启动时执行
	//
	// 中文说明:
	// - 每次服务启动时调用;
	// - 用于初始化运行时状态、读取配置、启动 goroutine;
	// - 在 ServiceProvider.Boot 中被调用。
	Boot(ctx context.Context, c contract.Container) error

	// ToServiceProvider 转换为 gorp ServiceProvider
	//
	// 中文说明:
	// - 这是连接产品插件和框架容器的桥梁;
	// - 返回的 ServiceProvider 会被注册到 Container;
	// - 实现中应返回一个包装了插件本身的 ServiceProvider。
	ToServiceProvider() contract.ServiceProvider
}

// PluginMeta 插件元数据
//
// 中文说明:
// - 对应 plugin.json 文件内容;
// - 用于插件的发现、展示和版本管理;
// - SystemName 是唯一标识,命名格式: {Type}.{Provider}。
type PluginMeta struct {
	// Group 插件分组
	//
	// 中文说明:
	// - 用于在管理界面分组展示;
	// - 例如: "Payment", "Shipping", "Misc", "Widgets"。
	Group string `json:"group"`

	// FriendlyName 友好名称
	//
	// 中文说明:
	// - 显示给用户的名称;
	// - 例如: "支付宝支付", "顺丰速运"。
	FriendlyName string `json:"friendly_name"`

	// SystemName 系统名称
	//
	// 中文说明:
	// - 唯一标识,用于查找和配置;
	// - 命名格式: {Type}.{Provider};
	// - 例如: "Payment.Alipay", "Shipping.FedEx"。
	SystemName string `json:"system_name"`

	// Version 插件版本
	//
	// 中文说明:
	// - 遵循语义化版本规范;
	// - 例如: "1.0.0", "2.1.0"。
	Version string `json:"version"`

	// SupportedVersions 支持的 nop-go 版本
	//
	// 中文说明:
	// - 指明兼容的主版本;
	// - 例如: ["1.0", "1.1"]。
	SupportedVersions []string `json:"supported_versions"`

	// Author 作者
	Author string `json:"author"`

	// DisplayOrder 显示顺序
	//
	// 中文说明:
	// - 用于列表排序;
	// - 数字越小排在越前面。
	DisplayOrder int `json:"display_order"`

	// Description 插件描述
	Description string `json:"description"`

	// DependsOn 依赖的其他插件
	//
	// 中文说明:
	// - 命明前置依赖的 SystemName 列表;
	// - Manager 会按依赖顺序加载。
	DependsOn []string `json:"depends_on"`

	// FileName 编译后的文件名
	//
	// 中文说明:
	// - 用于 Phase 2 动态加载(.so);
	// - Phase 1 编译进主程序时可为空。
	FileName string `json:"file_name"`

	// Installed 是否已安装
	//
	// 中文说明:
	// - 由 Manager 维护;
	// - true 表示 Install 已执行过。
	Installed bool `json:"installed"`

	// InstalledVersion 已安装的版本
	//
	// 中文说明:
	// - 用于检测版本更新;
	// - 与 Version 不同时可能需要升级。
	InstalledVersion string `json:"installed_version"`
}
// Package plugin 小部件插件接口
package plugin

import (
	"context"
)

// WidgetPlugin 小部件插件接口
//
// 中文说明:
// - 用于向前端页面注入自定义内容块;
// - 例如: 广告位、推荐商品、统计代码等;
// - 继承基础 Plugin 接口,增加小部件特有能力;
// - GetWidgetZones 返回可注入的页面区域;
// - RenderWidget 渲染小部件 HTML 内容。
type WidgetPlugin interface {
	Plugin

	// GetWidgetZones 获取可用的小部件区域
	//
	// 中文说明:
	// - 定义小部件可以显示的页面位置;
	// - 例如: "home_top", "product_detail_bottom", "cart_summary";
	// - 管理后台可配置小部件显示在哪些区域。
	GetWidgetZones() []string

	// RenderWidget 渲染小部件内容
	//
	// 中文说明:
	// - 根据区域和参数生成 HTML 内容;
	// - zone 指定显示位置;
	// - params 包含额外配置参数;
	// - 返回的 HTML 会直接注入页面。
	RenderWidget(ctx context.Context, zone string, params map[string]interface{}) (string, error)

	// GetWidgetSettings 获取小部件默认配置
	//
	// 中文说明:
	// - 返回可在管理后台调整的配置项;
	// - 例如: 广告位图片、推荐商品数量等。
	GetWidgetSettings() []WidgetSettingItem
}

// WidgetZone 小部件区域定义
//
// 中文说明:
// - 预定义的页面注入点;
// - 插件可选择支持的区域;
// - 系统会按区域组织小部件列表。
type WidgetZone struct {
	// SystemName 区域系统名称
	//
	// 中文说明:
	// - 唯一标识;
	// - 例如: "home_top", "product_detail_sidebar"。
	SystemName string

	// FriendlyName 友好名称
	//
	// 中文说明:
	// - 显示给管理员的名称;
	// - 例如: "首页顶部", "商品详情侧边栏"。
	FriendlyName string

	// Description 区域描述
	Description string
}

// WidgetSettingItem 小部件配置项
//
// 中文说明:
// - 描述一个配置字段的元信息;
// - 用于在管理后台生成配置表单。
type WidgetSettingItem struct {
	// Name 配置字段名
	Name string

	// Label 显示标签
	Label string

	// Type 输入类型
	//
	// 中文说明:
	// - text: 单行文本;
	// - textarea: 多行文本;
	// - number: 数字输入;
	// - boolean: 开关;
	// - select: 下拉选择;
	// - image: 图片上传。
	Type string

	// Required 是否必填
	Required bool

	// Default 默认值
	Default string

	// Options 下拉选项(仅 select 类型)
	Options []WidgetSettingOption

	// HelpText 帮助文本
	HelpText string
}

// WidgetSettingOption 下拉选项
type WidgetSettingOption struct {
	Value string
	Label string
}

// 预定义的小部件区域
//
// 中文说明:
// - 系统内置的页面注入点;
// - 插件可选择支持其中部分或全部。
var BuiltInWidgetZones = []WidgetZone{
	{SystemName: "home_top", FriendlyName: "首页顶部", Description: "首页顶部区域,适合放置广告或公告"},
	{SystemName: "home_bottom", FriendlyName: "首页底部", Description: "首页底部区域"},
	{SystemName: "home_sidebar", FriendlyName: "首页侧边栏", Description: "首页侧边栏区域"},
	{SystemName: "product_list_top", FriendlyName: "商品列表顶部", Description: "商品列表页顶部"},
	{SystemName: "product_list_bottom", FriendlyName: "商品列表底部", Description: "商品列表页底部"},
	{SystemName: "product_detail_top", FriendlyName: "商品详情顶部", Description: "商品详情页顶部"},
	{SystemName: "product_detail_bottom", FriendlyName: "商品详情底部", Description: "商品详情页底部,适合放置推荐商品"},
	{SystemName: "product_detail_sidebar", FriendlyName: "商品详情侧边栏", Description: "商品详情页侧边栏"},
	{SystemName: "cart_top", FriendlyName: "购物车顶部", Description: "购物车页顶部"},
	{SystemName: "cart_bottom", FriendlyName: "购物车底部", Description: "购物车页底部"},
	{SystemName: "checkout_top", FriendlyName: "结算页顶部", Description: "结算页顶部"},
	{SystemName: "checkout_bottom", FriendlyName: "结算页底部", Description: "结算页底部"},
	{SystemName: "order_detail_top", FriendlyName: "订单详情顶部", Description: "订单详情页顶部"},
	{SystemName: "order_detail_bottom", FriendlyName: "订单详情底部", Description: "订单详情页底部"},
	{SystemName: "footer", FriendlyName: "页脚", Description: "全局页脚区域,适合放置统计代码"},
	{SystemName: "head", FriendlyName: "HEAD区域", Description: "HTML HEAD 区域,适合放置脚本标签"},
}
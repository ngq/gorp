// Package alipay 支付宝支付插件
//
// 中文说明:
// - 实现支付宝网页支付、APP支付、扫码支付;
// - 支持 PaymentMethod 接口;
// - 通过 ToServiceProvider 注册到 gorp Container;
// - 配置项包括: app_id、private_key、public_key、sandbox 模式。
package alipay

import (
	"context"
	"fmt"
	"time"

	"nop-go/shared/plugin"

	"github.com/ngq/gorp/framework/contract"
)

// AlipayPlugin 支付宝支付插件实现
//
// 中文说明:
// - 实现 plugin.Plugin 和 plugin.PaymentMethod 接口;
// - 通过 config 字段存储从配置服务读取的参数;
// - sandbox 模式用于开发测试。
type AlipayPlugin struct {
	meta   *plugin.PluginMeta
	config map[string]string
}

// New 创建支付宝插件实例
//
// 中文说明:
// - 创建时初始化默认元数据;
// - 配置从 gorp Config 服务读取;
// - 通常在服务启动时创建并注册。
func New() *AlipayPlugin {
	return &AlipayPlugin{
		meta: &plugin.PluginMeta{
			Group:             "Payment",
			FriendlyName:      "支付宝支付",
			SystemName:        "Payment.Alipay",
			Version:           "1.0.0",
			SupportedVersions: []string{"1.0"},
			Author:            "nop-go Team",
			DisplayOrder:      1,
			Description:       "支付宝网页支付、APP支付、扫码支付",
		},
		config: make(map[string]string),
	}
}

// Meta 返回插件元数据
func (p *AlipayPlugin) Meta() *plugin.PluginMeta {
	return p.meta
}

// PluginType 返回插件类型
func (p *AlipayPlugin) PluginType() string {
	return "payment"
}

// Install 插件安装时执行
//
// 中文说明:
// - 创建支付宝支付方式记录到数据库;
// - 实际项目中应该用 GORM 创建 payment_methods 表记录;
// - 这里简化为打印日志。
func (p *AlipayPlugin) Install(ctx context.Context, c contract.Container) error {
	// 实际项目中应该:
	// 1. 创建 payment_methods 记录
	// 2. 创建默认配置
	// 3. 通知用户去配置支付参数

	// 获取日志服务(可选)
	logger, err := c.Make(contract.LogKey)
	if err == nil {
		if log, ok := logger.(contract.Logger); ok {
			log.Info("Installing Alipay payment plugin...")
		}
	}

	fmt.Println("[Alipay] Plugin installed")
	return nil
}

// Uninstall 插件卸载时执行
//
// 中文说明:
// - 标记 payment_methods 记录为已卸载;
// - 通常不建议删除数据。
func (p *AlipayPlugin) Uninstall(ctx context.Context, c contract.Container) error {
	fmt.Println("[Alipay] Plugin uninstalled")
	return nil
}

// Boot 插件启动时执行
//
// 中文说明:
// - 从配置服务读取支付宝参数;
// - 参数包括: app_id、private_key、public_key、sandbox;
// - 初始化支付宝 SDK 客户端。
func (p *AlipayPlugin) Boot(ctx context.Context, c contract.Container) error {
	// 从配置服务读取参数
	cfg, err := c.Make(contract.ConfigKey)
	if err != nil {
		// 配置服务不可用时使用空配置
		// 实际项目中应该返回错误或使用默认值
		fmt.Println("[Alipay] Config service not available, using defaults")
		return nil
	}

	config, ok := cfg.(contract.Config)
	if !ok {
		return nil
	}

	// 读取支付宝配置
	p.config["app_id"] = config.GetString("plugins.alipay.app_id")
	p.config["private_key"] = config.GetString("plugins.alipay.private_key")
	p.config["public_key"] = config.GetString("plugins.alipay.public_key")
	p.config["sandbox"] = config.GetString("plugins.alipay.sandbox")

	fmt.Printf("[Alipay] Plugin booted, app_id: %s, sandbox: %s\n",
		p.config["app_id"], p.config["sandbox"])

	return nil
}

// ToServiceProvider 转换为 gorp ServiceProvider
//
// 中文说明:
// - 这是连接产品插件和框架容器的桥梁;
// - 返回的 ServiceProvider 会被注册到 Container;
// - Boot 时将插件注册到全局 Registry。
func (p *AlipayPlugin) ToServiceProvider() contract.ServiceProvider {
	return &AlipayServiceProvider{plugin: p}
}

// ProcessPayment 处理支付
//
// 中文说明:
// - 创建支付宝支付订单;
// - 返回支付链接供前端跳转;
// - 实际项目中应调用支付宝 SDK。
func (p *AlipayPlugin) ProcessPayment(ctx context.Context, req *plugin.ProcessPaymentRequest) (*plugin.ProcessPaymentResult, error) {
	// 检查配置
	if p.config["app_id"] == "" {
		return nil, fmt.Errorf("alipay app_id not configured")
	}

	// 实际项目中应该:
	// 1. 调用支付宝 SDK 创建订单
	// 2. 生成支付链接或二维码
	// 3. 记录请求日志

	// 简化实现:生成模拟支付链接
	txnID := fmt.Sprintf("ALI%s%d", time.Now().Format("20060102150405"), req.OrderID)

	// 判断沙箱模式
	baseURL := "https://openapi.alipay.com/gateway.do"
	if p.config["sandbox"] == "true" {
		baseURL = "https://openapi.alipaydev.com/gateway.do"
	}

	redirectURL := fmt.Sprintf("%s?app_id=%s&order_id=%d&amount=%.2f",
		baseURL, p.config["app_id"], req.OrderID, req.Amount)

	return &plugin.ProcessPaymentResult{
		Success:       true,
		TransactionID: txnID,
		RedirectURL:   redirectURL,
		RawData: map[string]interface{}{
			"app_id":     p.config["app_id"],
			"order_id":   req.OrderID,
			"amount":     req.Amount,
			"currency":   req.Currency,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// Refund 退款
//
// 中文说明:
// - 支付宝支持全额和部分退款;
// - 需要原交易流水号;
// - 返回退款流水号。
func (p *AlipayPlugin) Refund(ctx context.Context, req *plugin.RefundRequest) (*plugin.RefundResult, error) {
	if p.config["app_id"] == "" {
		return nil, fmt.Errorf("alipay app_id not configured")
	}

	// 实际项目中应该调用支付宝退款接口
	refundTxnID := fmt.Sprintf("REF%s%s", time.Now().Format("20060102150405"), req.TransactionID)

	return &plugin.RefundResult{
		Success:             true,
		RefundTransactionID: refundTxnID,
	}, nil
}

// Capture 捕获预授权
//
// 中文说明:
// - 支付宝不支持预授权模式;
// - 返回错误提示。
func (p *AlipayPlugin) Capture(ctx context.Context, req *plugin.CaptureRequest) (*plugin.CaptureResult, error) {
	return nil, fmt.Errorf("alipay does not support capture (pre-authorization)")
}

// Void 取消预授权
//
// 中文说明:
// - 支付宝不支持预授权模式;
// - 返回错误提示。
func (p *AlipayPlugin) Void(ctx context.Context, req *plugin.VoidRequest) (*plugin.VoidResult, error) {
	return nil, fmt.Errorf("alipay does not support void (cancel pre-authorization)")
}

// GetConfiguration 获取支付配置项
//
// 中文说明:
// - 返回需要在管理后台配置的字段列表;
// - 用于生成动态配置表单;
// - Type 决定输入控件类型。
func (p *AlipayPlugin) GetConfiguration() []plugin.PaymentConfigItem {
	return []plugin.PaymentConfigItem{
		{
			Name:     "app_id",
			Label:    "应用ID (AppID)",
			Type:     "text",
			Required: true,
			HelpText: "在支付宝开放平台创建应用后获取",
		},
		{
			Name:     "private_key",
			Label:    "应用私钥",
			Type:     "textarea",
			Required: true,
			HelpText: "使用 RSA2 算法生成的应用私钥",
		},
		{
			Name:     "public_key",
			Label:    "支付宝公钥",
			Type:     "textarea",
			Required: true,
			HelpText: "支付宝应用的公钥,用于验签",
		},
		{
			Name:     "sandbox",
			Label:    "沙箱模式",
			Type:     "boolean",
			Required: false,
			Default:  "false",
			HelpText: "开启后使用支付宝沙箱环境进行测试",
		},
		{
			Name:     "notify_url",
			Label:    "异步通知URL",
			Type:     "text",
			Required: false,
			HelpText: "支付结果异步通知地址,需外网可访问",
		},
		{
			Name:     "return_url",
			Label:    "同步返回URL",
			Type:     "text",
			Required: false,
			HelpText: "支付完成后浏览器跳转地址",
		},
	}
}

// ValidateConfiguration 验证配置是否正确
//
// 中文说明:
// - 保存配置前验证;
// - 检查必填项、格式等;
// - 返回具体的错误信息便于修正。
func (p *AlipayPlugin) ValidateConfiguration(config map[string]string) error {
	if config["app_id"] == "" {
		return fmt.Errorf("app_id 是必填项")
	}
	if len(config["app_id"]) < 16 {
		return fmt.Errorf("app_id 格式不正确,应为16位数字")
	}
	if config["private_key"] == "" {
		return fmt.Errorf("private_key 是必填项")
	}
	if config["public_key"] == "" {
		return fmt.Errorf("public_key 是必填项")
	}

	// 实际项目中还可以:
	// 1. 验证私钥格式
	// 2. 调用支付宝接口验证配置有效性
	// 3. 检查 notify_url 是否外网可访问

	return nil
}

// AlipayServiceProvider gorp ServiceProvider 实现
//
// 中文说明:
// - 包装 AlipayPlugin 实现 ServiceProvider 接口;
// - Register 时绑定插件实例到 Container;
// - Boot 时注册到全局 Registry。
type AlipayServiceProvider struct {
	plugin *AlipayPlugin
}

// Name 返回 ServiceProvider 名称
func (sp *AlipayServiceProvider) Name() string {
	return "plugin.payment.alipay"
}

// IsDefer 是否延迟加载
//
// 中文说明:
// - 支付插件通常需要立即加载;
// - 返回 false 表示注册时立即执行 Register/Boot。
func (sp *AlipayServiceProvider) IsDefer() bool {
	return false
}

// Provides 返回提供的服务 Key 列表
//
// 中文说明:
// - 声明此 Provider 提供哪些服务;
// - Container.Make("plugin.payment.alipay") 可获取插件实例。
func (sp *AlipayServiceProvider) Provides() []string {
	return []string{"plugin.payment.alipay"}
}

// Register 注册阶段
//
// 中文说明:
// - 将插件实例绑定到 Container;
// - singleton=true 表示单例模式;
// - 业务代码可通过 Make 获取插件。
func (sp *AlipayServiceProvider) Register(c contract.Container) error {
	c.Bind("plugin.payment.alipay", func(c contract.Container) (interface{}, error) {
		return sp.plugin, nil
	}, true)
	return nil
}

// Boot 启动阶段
//
// 中文说明:
// - 注册到全局 Registry;
// - 这样其他服务可通过 Registry 查找插件;
// - 例如: GetRegistry().GetPaymentMethod("Payment.Alipay")。
func (sp *AlipayServiceProvider) Boot(c contract.Container) error {
	// 注册到全局 Registry
	plugin.GetRegistry().Register(sp.plugin)

	// 执行插件的 Boot 方法
	ctx := context.Background()
	return sp.plugin.Boot(ctx, c)
}
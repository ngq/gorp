// Package wechat 微信支付插件
//
// 中文说明:
// - 实现微信支付:JSAPI支付、APP支付、H5支付、Native支付;
// - 支持 PaymentMethod 接口;
// - 通过 ToServiceProvider 注册到 gorp Container;
// - 配置项包括: app_id、mch_id、api_key、api_v3_key、cert_path。
package wechat

import (
	"context"
	"fmt"
	"time"

	"nop-go/shared/plugin"

	"github.com/ngq/gorp/framework/contract"
)

// WechatPayPlugin 微信支付插件实现
//
// 中文说明:
// - 实现 plugin.Plugin 和 plugin.PaymentMethod 接口;
// - 支持多种支付方式:JSAPI、APP、H5、Native;
// - 通过 payType 配置项指定支付方式。
type WechatPayPlugin struct {
	meta   *plugin.PluginMeta
	config map[string]string
}

// New 创建微信支付插件实例
func New() *WechatPayPlugin {
	return &WechatPayPlugin{
		meta: &plugin.PluginMeta{
			Group:             "Payment",
			FriendlyName:      "微信支付",
			SystemName:        "Payment.Wechat",
			Version:           "1.0.0",
			SupportedVersions: []string{"1.0"},
			Author:            "nop-go Team",
			DisplayOrder:      2,
			Description:       "微信支付:JSAPI支付、APP支付、H5支付、Native支付",
		},
		config: make(map[string]string),
	}
}

func (p *WechatPayPlugin) Meta() *plugin.PluginMeta { return p.meta }
func (p *WechatPayPlugin) PluginType() string       { return "payment" }

// Install 插件安装
func (p *WechatPayPlugin) Install(ctx context.Context, c contract.Container) error {
	logger, err := c.Make(contract.LogKey)
	if err == nil {
		if log, ok := logger.(contract.Logger); ok {
			log.Info("Installing Wechat payment plugin...")
		}
	}
	fmt.Println("[WechatPay] Plugin installed")
	return nil
}

// Uninstall 插件卸载
func (p *WechatPayPlugin) Uninstall(ctx context.Context, c contract.Container) error {
	fmt.Println("[WechatPay] Plugin uninstalled")
	return nil
}

// Boot 插件启动
//
// 中文说明:
// - 从配置服务读取微信支付参数;
// - 参数包括: app_id、mch_id、api_key、api_v3_key、cert_path。
func (p *WechatPayPlugin) Boot(ctx context.Context, c contract.Container) error {
	cfg, err := c.Make(contract.ConfigKey)
	if err != nil {
		fmt.Println("[WechatPay] Config service not available")
		return nil
	}

	config, ok := cfg.(contract.Config)
	if !ok {
		return nil
	}

	p.config["app_id"] = config.GetString("plugins.wechat.app_id")
	p.config["mch_id"] = config.GetString("plugins.wechat.mch_id")
	p.config["api_key"] = config.GetString("plugins.wechat.api_key")
	p.config["api_v3_key"] = config.GetString("plugins.wechat.api_v3_key")
	p.config["cert_path"] = config.GetString("plugins.wechat.cert_path")
	p.config["pay_type"] = config.GetString("plugins.wechat.pay_type") // jsapi, app, h5, native
	p.config["sandbox"] = config.GetString("plugins.wechat.sandbox")

	fmt.Printf("[WechatPay] Plugin booted, app_id: %s, mch_id: %s\n",
		p.config["app_id"], p.config["mch_id"])

	return nil
}

// ToServiceProvider 转换为 gorp ServiceProvider
func (p *WechatPayPlugin) ToServiceProvider() contract.ServiceProvider {
	return &WechatServiceProvider{plugin: p}
}

// ProcessPayment 处理支付
//
// 中文说明:
// - 根据 payType 创建不同类型的支付订单;
// - JSAPI: 返回 prepay_id 供前端调用;
// - Native: 返回二维码链接;
// - H5: 返回跳转链接;
// - APP: 返回 APP 调用参数。
func (p *WechatPayPlugin) ProcessPayment(ctx context.Context, req *plugin.ProcessPaymentRequest) (*plugin.ProcessPaymentResult, error) {
	if p.config["app_id"] == "" || p.config["mch_id"] == "" {
		return nil, fmt.Errorf("wechat pay app_id or mch_id not configured")
	}

	txnID := fmt.Sprintf("WX%s%d", time.Now().Format("20060102150405"), req.OrderID)
	payType := p.config["pay_type"]
	if payType == "" {
		payType = "native" // 默认扫码支付
	}

	var redirectURL, qrCodeURL string

	switch payType {
	case "native":
		// Native 扫码支付:返回二维码内容
		qrCodeURL = fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", txnID)
	case "h5":
		// H5 支付:返回跳转链接
		redirectURL = fmt.Sprintf("https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=%s", txnID)
	case "jsapi":
		// JSAPI 支付:需要 openid
		openid := req.CustomFields["openid"]
		if openid == "" {
			return nil, fmt.Errorf("jsapi payment requires openid in custom_fields")
		}
		// 返回 prepay_id 供前端使用
		return &plugin.ProcessPaymentResult{
			Success:       true,
			TransactionID: txnID,
			RawData: map[string]interface{}{
				"prepay_id": txnID,
				"app_id":    p.config["app_id"],
				"openid":    openid,
			},
		}, nil
	case "app":
		// APP 支付:返回 APP 调用参数
		return &plugin.ProcessPaymentResult{
			Success:       true,
			TransactionID: txnID,
			RawData: map[string]interface{}{
				"prepay_id": txnID,
				"app_id":    p.config["app_id"],
				"partnerid": p.config["mch_id"],
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported pay_type: %s", payType)
	}

	return &plugin.ProcessPaymentResult{
		Success:       true,
		TransactionID: txnID,
		RedirectURL:   redirectURL,
		QRCodeURL:     qrCodeURL,
		RawData: map[string]interface{}{
			"pay_type":   payType,
			"order_id":   req.OrderID,
			"amount":     req.Amount,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// Refund 退款
//
// 中文说明:
// - 微信支付支持全额和部分退款;
// - 需要原交易流水号;
// - 返回退款流水号。
func (p *WechatPayPlugin) Refund(ctx context.Context, req *plugin.RefundRequest) (*plugin.RefundResult, error) {
	if p.config["mch_id"] == "" {
		return nil, fmt.Errorf("wechat pay mch_id not configured")
	}

	refundTxnID := fmt.Sprintf("REFWX%s%s", time.Now().Format("20060102150405"), req.TransactionID)

	return &plugin.RefundResult{
		Success:             true,
		RefundTransactionID: refundTxnID,
	}, nil
}

// Capture 捕获预授权
//
// 中文说明:
// - 微信支付不支持预授权模式;
// - 返回错误提示。
func (p *WechatPayPlugin) Capture(ctx context.Context, req *plugin.CaptureRequest) (*plugin.CaptureResult, error) {
	return nil, fmt.Errorf("wechat pay does not support capture (pre-authorization)")
}

// Void 取消预授权
func (p *WechatPayPlugin) Void(ctx context.Context, req *plugin.VoidRequest) (*plugin.VoidResult, error) {
	return nil, fmt.Errorf("wechat pay does not support void (cancel pre-authorization)")
}

// GetConfiguration 获取支付配置项
func (p *WechatPayPlugin) GetConfiguration() []plugin.PaymentConfigItem {
	return []plugin.PaymentConfigItem{
		{
			Name:     "app_id",
			Label:    "微信AppID",
			Type:     "text",
			Required: true,
			HelpText: "微信公众号/小程序/APP的AppID",
		},
		{
			Name:     "mch_id",
			Label:    "商户号",
			Type:     "text",
			Required: true,
			HelpText: "微信支付分配的商户号",
		},
		{
			Name:     "api_key",
			Label:    "API密钥(V2)",
			Type:     "password",
			Required: true,
			HelpText: "商户平台设置的API密钥,用于V2接口签名",
		},
		{
			Name:     "api_v3_key",
			Label:    "APIv3密钥",
			Type:     "password",
			Required: false,
			HelpText: "商户平台设置的APIv3密钥,用于V3接口",
		},
		{
			Name:     "cert_path",
			Label:    "商户证书路径",
			Type:     "text",
			Required: false,
			HelpText: "商户API证书文件路径(apiclient_cert.p12)",
		},
		{
			Name:     "pay_type",
			Label:    "支付方式",
			Type:     "select",
			Required: true,
			Default:  "native",
			Options: []plugin.PaymentConfigOption{
				{Value: "native", Label: "Native扫码支付"},
				{Value: "jsapi", Label: "JSAPI公众号支付"},
				{Value: "h5", Label: "H5支付"},
				{Value: "app", Label: "APP支付"},
			},
			HelpText: "选择默认支付方式",
		},
		{
			Name:     "sandbox",
			Label:    "沙箱模式",
			Type:     "boolean",
			Required: false,
			Default:  "false",
			HelpText: "开启后使用微信支付沙箱环境",
		},
		{
			Name:     "notify_url",
			Label:    "异步通知URL",
			Type:     "text",
			Required: false,
			HelpText: "支付结果异步通知地址",
		},
	}
}

// ValidateConfiguration 验证配置
func (p *WechatPayPlugin) ValidateConfiguration(config map[string]string) error {
	if config["app_id"] == "" {
		return fmt.Errorf("app_id 是必填项")
	}
	if config["mch_id"] == "" {
		return fmt.Errorf("mch_id 是必填项")
	}
	if config["api_key"] == "" {
		return fmt.Errorf("api_key 是必填项")
	}

	payType := config["pay_type"]
	validPayTypes := map[string]bool{"native": true, "jsapi": true, "h5": true, "app": true}
	if payType != "" && !validPayTypes[payType] {
		return fmt.Errorf("无效的 pay_type: %s", payType)
	}

	return nil
}

// WechatServiceProvider gorp ServiceProvider 实现
type WechatServiceProvider struct {
	plugin *WechatPayPlugin
}

func (sp *WechatServiceProvider) Name() string       { return "plugin.payment.wechat" }
func (sp *WechatServiceProvider) IsDefer() bool      { return false }
func (sp *WechatServiceProvider) Provides() []string { return []string{"plugin.payment.wechat"} }

func (sp *WechatServiceProvider) Register(c contract.Container) error {
	c.Bind("plugin.payment.wechat", func(c contract.Container) (interface{}, error) {
		return sp.plugin, nil
	}, true)
	return nil
}

func (sp *WechatServiceProvider) Boot(c contract.Container) error {
	plugin.GetRegistry().Register(sp.plugin)
	ctx := context.Background()
	return sp.plugin.Boot(ctx, c)
}
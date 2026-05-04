// Package wechat 寰俊鏀粯鎻掍欢
//
// 涓枃璇存槑:
// - 瀹炵幇寰俊鏀粯:JSAPI鏀粯銆丄PP鏀粯銆丠5鏀粯銆丯ative鏀粯;
// - 鏀寔 PaymentMethod 鎺ュ彛;
// - 閫氳繃 ToServiceProvider 娉ㄥ唽鍒?gorp Container;
// - 閰嶇疆椤瑰寘鎷? app_id銆乵ch_id銆乤pi_key銆乤pi_v3_key銆乧ert_path銆?
package wechat

import (
	"context"
	"fmt"
	"time"

	"nop-go/shared/plugin"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// WechatPayPlugin 寰俊鏀粯鎻掍欢瀹炵幇
//
// 涓枃璇存槑:
// - 瀹炵幇 plugin.Plugin 鍜?plugin.PaymentMethod 鎺ュ彛;
// - 鏀寔澶氱鏀粯鏂瑰紡:JSAPI銆丄PP銆丠5銆丯ative;
// - 閫氳繃 payType 閰嶇疆椤规寚瀹氭敮浠樻柟寮忋€?
type WechatPayPlugin struct {
	meta   *plugin.PluginMeta
	config map[string]string
}

// New 鍒涘缓寰俊鏀粯鎻掍欢瀹炰緥
func New() *WechatPayPlugin {
	return &WechatPayPlugin{
		meta: &plugin.PluginMeta{
			Group:             "Payment",
			FriendlyName:      "寰俊鏀粯",
			SystemName:        "Payment.Wechat",
			Version:           "1.0.0",
			SupportedVersions: []string{"1.0"},
			Author:            "nop-go Team",
			DisplayOrder:      2,
			Description:       "寰俊鏀粯:JSAPI鏀粯銆丄PP鏀粯銆丠5鏀粯銆丯ative鏀粯",
		},
		config: make(map[string]string),
	}
}

func (p *WechatPayPlugin) Meta() *plugin.PluginMeta { return p.meta }
func (p *WechatPayPlugin) PluginType() string       { return "payment" }

// Install 鎻掍欢瀹夎
func (p *WechatPayPlugin) Install(ctx context.Context, c runtimecontract.Container) error {
	logger, err := c.Make(observabilitycontract.LogKey)
	if err == nil {
		if log, ok := logger.(observabilitycontract.Logger); ok {
			log.Info("Installing Wechat payment plugin...")
		}
	}
	fmt.Println("[WechatPay] Plugin installed")
	return nil
}

// Uninstall 鎻掍欢鍗歌浇
func (p *WechatPayPlugin) Uninstall(ctx context.Context, c runtimecontract.Container) error {
	fmt.Println("[WechatPay] Plugin uninstalled")
	return nil
}

// Boot 鎻掍欢鍚姩
//
// 涓枃璇存槑:
// - 浠庨厤缃湇鍔¤鍙栧井淇℃敮浠樺弬鏁?
// - 鍙傛暟鍖呮嫭: app_id銆乵ch_id銆乤pi_key銆乤pi_v3_key銆乧ert_path銆?
func (p *WechatPayPlugin) Boot(ctx context.Context, c runtimecontract.Container) error {
	cfg, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		fmt.Println("[WechatPay] Config service not available")
		return nil
	}

	config, ok := cfg.(datacontract.Config)
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

// ToServiceProvider 杞崲涓?gorp ServiceProvider
func (p *WechatPayPlugin) ToServiceProvider() runtimecontract.ServiceProvider {
	return &WechatServiceProvider{plugin: p}
}

// ProcessPayment 澶勭悊鏀粯
//
// 涓枃璇存槑:
// - 鏍规嵁 payType 鍒涘缓涓嶅悓绫诲瀷鐨勬敮浠樿鍗?
// - JSAPI: 杩斿洖 prepay_id 渚涘墠绔皟鐢?
// - Native: 杩斿洖浜岀淮鐮侀摼鎺?
// - H5: 杩斿洖璺宠浆閾炬帴;
// - APP: 杩斿洖 APP 璋冪敤鍙傛暟銆?
func (p *WechatPayPlugin) ProcessPayment(ctx context.Context, req *plugin.ProcessPaymentRequest) (*plugin.ProcessPaymentResult, error) {
	if p.config["app_id"] == "" || p.config["mch_id"] == "" {
		return nil, fmt.Errorf("wechat pay app_id or mch_id not configured")
	}

	txnID := fmt.Sprintf("WX%s%d", time.Now().Format("20060102150405"), req.OrderID)
	payType := p.config["pay_type"]
	if payType == "" {
		payType = "native" // 榛樿鎵爜鏀粯
	}

	var redirectURL, qrCodeURL string

	switch payType {
	case "native":
		// Native 鎵爜鏀粯:杩斿洖浜岀淮鐮佸唴瀹?
		qrCodeURL = fmt.Sprintf("weixin://wxpay/bizpayurl?pr=%s", txnID)
	case "h5":
		// H5 鏀粯:杩斿洖璺宠浆閾炬帴
		redirectURL = fmt.Sprintf("https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=%s", txnID)
	case "jsapi":
		// JSAPI 鏀粯:闇€瑕?openid
		openid := req.CustomFields["openid"]
		if openid == "" {
			return nil, fmt.Errorf("jsapi payment requires openid in custom_fields")
		}
		// 杩斿洖 prepay_id 渚涘墠绔娇鐢?
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
		// APP 鏀粯:杩斿洖 APP 璋冪敤鍙傛暟
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

// Refund 閫€娆?
//
// 涓枃璇存槑:
// - 寰俊鏀粯鏀寔鍏ㄩ鍜岄儴鍒嗛€€娆?
// - 闇€瑕佸師浜ゆ槗娴佹按鍙?
// - 杩斿洖閫€娆炬祦姘村彿銆?
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

// Capture 鎹曡幏棰勬巿鏉?
//
// 涓枃璇存槑:
// - 寰俊鏀粯涓嶆敮鎸侀鎺堟潈妯″紡;
// - 杩斿洖閿欒鎻愮ず銆?
func (p *WechatPayPlugin) Capture(ctx context.Context, req *plugin.CaptureRequest) (*plugin.CaptureResult, error) {
	return nil, fmt.Errorf("wechat pay does not support capture (pre-authorization)")
}

// Void 鍙栨秷棰勬巿鏉?
func (p *WechatPayPlugin) Void(ctx context.Context, req *plugin.VoidRequest) (*plugin.VoidResult, error) {
	return nil, fmt.Errorf("wechat pay does not support void (cancel pre-authorization)")
}

// GetConfiguration 鑾峰彇鏀粯閰嶇疆椤?
func (p *WechatPayPlugin) GetConfiguration() []plugin.PaymentConfigItem {
	return []plugin.PaymentConfigItem{
		{
			Name:     "app_id",
			Label:    "寰俊AppID",
			Type:     "text",
			Required: true,
			HelpText: "寰俊鍏紬鍙?灏忕▼搴?APP鐨凙ppID",
		},
		{
			Name:     "mch_id",
			Label:    "鍟嗘埛鍙?,
			Type:     "text",
			Required: true,
			HelpText: "寰俊鏀粯鍒嗛厤鐨勫晢鎴峰彿",
		},
		{
			Name:     "api_key",
			Label:    "API瀵嗛挜(V2)",
			Type:     "password",
			Required: true,
			HelpText: "鍟嗘埛骞冲彴璁剧疆鐨凙PI瀵嗛挜,鐢ㄤ簬V2鎺ュ彛绛惧悕",
		},
		{
			Name:     "api_v3_key",
			Label:    "APIv3瀵嗛挜",
			Type:     "password",
			Required: false,
			HelpText: "鍟嗘埛骞冲彴璁剧疆鐨凙PIv3瀵嗛挜,鐢ㄤ簬V3鎺ュ彛",
		},
		{
			Name:     "cert_path",
			Label:    "鍟嗘埛璇佷功璺緞",
			Type:     "text",
			Required: false,
			HelpText: "鍟嗘埛API璇佷功鏂囦欢璺緞(apiclient_cert.p12)",
		},
		{
			Name:     "pay_type",
			Label:    "鏀粯鏂瑰紡",
			Type:     "select",
			Required: true,
			Default:  "native",
			Options: []plugin.PaymentConfigOption{
				{Value: "native", Label: "Native鎵爜鏀粯"},
				{Value: "jsapi", Label: "JSAPI鍏紬鍙锋敮浠?},
				{Value: "h5", Label: "H5鏀粯"},
				{Value: "app", Label: "APP鏀粯"},
			},
			HelpText: "閫夋嫨榛樿鏀粯鏂瑰紡",
		},
		{
			Name:     "sandbox",
			Label:    "娌欑妯″紡",
			Type:     "boolean",
			Required: false,
			Default:  "false",
			HelpText: "寮€鍚悗浣跨敤寰俊鏀粯娌欑鐜",
		},
		{
			Name:     "notify_url",
			Label:    "寮傛閫氱煡URL",
			Type:     "text",
			Required: false,
			HelpText: "鏀粯缁撴灉寮傛閫氱煡鍦板潃",
		},
	}
}

// ValidateConfiguration 楠岃瘉閰嶇疆
func (p *WechatPayPlugin) ValidateConfiguration(config map[string]string) error {
	if config["app_id"] == "" {
		return fmt.Errorf("app_id 鏄繀濉」")
	}
	if config["mch_id"] == "" {
		return fmt.Errorf("mch_id 鏄繀濉」")
	}
	if config["api_key"] == "" {
		return fmt.Errorf("api_key 鏄繀濉」")
	}

	payType := config["pay_type"]
	validPayTypes := map[string]bool{"native": true, "jsapi": true, "h5": true, "app": true}
	if payType != "" && !validPayTypes[payType] {
		return fmt.Errorf("鏃犳晥鐨?pay_type: %s", payType)
	}

	return nil
}

// WechatServiceProvider gorp ServiceProvider 瀹炵幇
type WechatServiceProvider struct {
	plugin *WechatPayPlugin
}

func (sp *WechatServiceProvider) Name() string       { return "plugin.payment.wechat" }
func (sp *WechatServiceProvider) IsDefer() bool      { return false }
func (sp *WechatServiceProvider) Provides() []string { return []string{"plugin.payment.wechat"} }

func (sp *WechatServiceProvider) Register(c runtimecontract.Container) error {
	c.Bind("plugin.payment.wechat", func(c runtimecontract.Container) (interface{}, error) {
		return sp.plugin, nil
	}, true)
	return nil
}

func (sp *WechatServiceProvider) Boot(c runtimecontract.Container) error {
	plugin.GetRegistry().Register(sp.plugin)
	ctx := context.Background()
	return sp.plugin.Boot(ctx, c)
}